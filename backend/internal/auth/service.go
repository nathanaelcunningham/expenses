package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations
type Service struct {
	db     *sql.DB
	logger zerolog.Logger
}

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose password hash
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session represents an active user session
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FamilyID   string    `json:"family_id"`
	UserRole   string    `json:"user_role"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
	UserAgent  string    `json:"user_agent,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// SessionValidationResult contains session validation results
type SessionValidationResult struct {
	Valid    bool     `json:"valid"`
	Session  *Session `json:"session,omitempty"`
	User     *User    `json:"user,omitempty"`
	FamilyID string   `json:"family_id,omitempty"`
}

// AuthError represents authentication-specific errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error [%s]: %s", e.Code, e.Message)
}

// Common auth error codes
var (
	ErrInvalidCredentials = &AuthError{"INVALID_CREDENTIALS", "Invalid email or password"}
	ErrUserExists         = &AuthError{"USER_EXISTS", "User with this email already exists"}
	ErrInvalidSession     = &AuthError{"INVALID_SESSION", "Session is invalid or expired"}
	ErrUserNotFound       = &AuthError{"USER_NOT_FOUND", "User not found"}
	ErrWeakPassword       = &AuthError{"WEAK_PASSWORD", "Password must be at least 8 characters long"}
	ErrInvalidEmail       = &AuthError{"INVALID_EMAIL", "Invalid email format"}
)

// NewService creates a new authentication service
func NewService(db *sql.DB, logger zerolog.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger.With().Str("component", "auth-service").Logger(),
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	// Validate input
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Check if user exists
	exists, err := s.userExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID, err := s.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Create user
	user := &User{
		ID:           userID,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Name:         strings.TrimSpace(req.Name),
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = s.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.PasswordHash,
		user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Msg("User registered successfully")

	return user, nil
}

// Login authenticates a user and creates a session
func (s *Service) Login(ctx context.Context, req LoginRequest, userAgent, ipAddress string) (*Session, *User, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, nil, ErrInvalidCredentials
	}

	// Get user by email
	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if !s.verifyPassword(req.Password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	// Get user's family info
	familyID, userRole, err := s.getUserFamilyInfo(ctx, user.ID)
	if err != nil {
		// User might not have a family yet - that's okay for new users
		s.logger.Debug().
			Str("user_id", user.ID).
			Msg("User has no family membership yet")
	}

	// Create session
	session, err := s.createSession(ctx, user.ID, familyID, userRole, userAgent, ipAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID).
		Str("session_id", session.ID).
		Str("family_id", familyID).
		Msg("User logged in successfully")

	return session, user, nil
}

// ValidateSession checks if a session is valid and returns session info
func (s *Service) ValidateSession(ctx context.Context, sessionID string) (*SessionValidationResult, error) {
	if sessionID == "" {
		return &SessionValidationResult{Valid: false}, nil
	}

	// Get session from database
	session, err := s.getSession(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &SessionValidationResult{Valid: false}, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		s.deleteSession(ctx, sessionID)
		return &SessionValidationResult{Valid: false}, nil
	}

	// Update last active time
	if err := s.updateSessionActivity(ctx, sessionID); err != nil {
		s.logger.Warn().
			Err(err).
			Str("session_id", sessionID).
			Msg("Failed to update session activity")
	}

	// Get user info
	user, err := s.getUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &SessionValidationResult{
		Valid:    true,
		Session:  session,
		User:     user,
		FamilyID: session.FamilyID,
	}, nil
}

// Logout invalidates a session
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if err := s.deleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.Info().
		Str("session_id", sessionID).
		Msg("User logged out successfully")

	return nil
}

// RefreshSession extends a session's expiry time
func (s *Service) RefreshSession(ctx context.Context, sessionID string) (*Session, error) {
	// Get current session
	session, err := s.getSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrInvalidSession
	}

	// Extend expiry
	newExpiresAt := time.Now().Add(24 * time.Hour)
	query := `UPDATE user_sessions SET expires_at = ?, last_active = CURRENT_TIMESTAMP WHERE id = ?`
	
	_, err = s.db.ExecContext(ctx, query, newExpiresAt, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	session.ExpiresAt = newExpiresAt
	session.LastActive = time.Now()

	return session, nil
}

// UpdateUserFamily updates the user's family association when they join/leave a family
func (s *Service) UpdateUserFamily(ctx context.Context, userID, familyID, role string) error {
	// Update all active sessions for this user
	query := `
		UPDATE user_sessions 
		SET family_id = ?, user_role = ?
		WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP`

	_, err := s.db.ExecContext(ctx, query, familyID, role, userID)
	if err != nil {
		return fmt.Errorf("failed to update user sessions: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Str("family_id", familyID).
		Str("role", role).
		Msg("Updated user family association in sessions")

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *Service) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP`
	
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	
	if rowsAffected > 0 {
		s.logger.Info().
			Int64("sessions_deleted", rowsAffected).
			Msg("Cleaned up expired sessions")
	}

	return rowsAffected, nil
}

// Helper methods

func (s *Service) validateRegisterRequest(req RegisterRequest) error {
	if req.Email == "" {
		return ErrInvalidEmail
	}
	if req.Name == "" {
		return &AuthError{"INVALID_NAME", "Name is required"}
	}
	if len(req.Password) < 8 {
		return ErrWeakPassword
	}
	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		return ErrInvalidEmail
	}
	return nil
}

func (s *Service) userExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = ?`
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(&count)
	return count > 0, err
}

func (s *Service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *Service) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *Service) generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) getUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = ?`
	
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt)
	
	return user, err
}

func (s *Service) getUserByID(ctx context.Context, userID string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = ?`
	
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt)
	
	return user, err
}

func (s *Service) getUserFamilyInfo(ctx context.Context, userID string) (familyID, role string, err error) {
	query := `SELECT family_id, role FROM family_memberships WHERE user_id = ?`
	err = s.db.QueryRowContext(ctx, query, userID).Scan(&familyID, &role)
	return
}

func (s *Service) createSession(ctx context.Context, userID, familyID, userRole, userAgent, ipAddress string) (*Session, error) {
	sessionID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hour sessions

	session := &Session{
		ID:         sessionID,
		UserID:     userID,
		FamilyID:   familyID,
		UserRole:   userRole,
		CreatedAt:  now,
		LastActive: now,
		ExpiresAt:  expiresAt,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}

	query := `
		INSERT INTO user_sessions (id, user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.FamilyID, session.UserRole,
		session.CreatedAt, session.LastActive, session.ExpiresAt,
		session.UserAgent, session.IPAddress)

	return session, err
}

func (s *Service) getSession(ctx context.Context, sessionID string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address
		FROM user_sessions WHERE id = ?`

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.FamilyID, &session.UserRole,
		&session.CreatedAt, &session.LastActive, &session.ExpiresAt,
		&session.UserAgent, &session.IPAddress)

	return session, err
}

func (s *Service) updateSessionActivity(ctx context.Context, sessionID string) error {
	query := `UPDATE user_sessions SET last_active = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, sessionID)
	return err
}

func (s *Service) deleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM user_sessions WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, sessionID)
	return err
}

// SecureCompare performs a constant-time comparison of two strings
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}