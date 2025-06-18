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

	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/logger"
	authv1 "expenses-backend/pkg/auth/v1"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations
type Service struct {
	db      *sql.DB
	queries *masterdb.Queries
	logger  logger.Logger
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
func NewService(db *sql.DB, queries *masterdb.Queries, log logger.Logger) *Service {
	return &Service{
		db:      db,
		queries: queries,
		logger:  log.With(logger.Str("component", "auth-service")),
	}
}

// register creates a new user account
func (s *Service) register(ctx context.Context, req *authv1.RegisterRequest) (*masterdb.User, error) {
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
	now := time.Now()
	createParams := masterdb.CreateUserParams{
		ID:           userID,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Name:         strings.TrimSpace(req.Name),
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	sqlcUser, err := s.queries.CreateUser(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User registered successfully",
		logger.Str("user_id", sqlcUser.ID),
		logger.Str("email", sqlcUser.Email),
	)

	return sqlcUser, nil
}

// login authenticates a user and creates a session
func (s *Service) login(ctx context.Context, req *authv1.LoginRequest, userAgent, ipAddress string) (*masterdb.UserSession, *masterdb.User, error) {
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
		s.logger.Debug("User has no family membership yet",
			logger.Str("user_id", user.ID))
	}

	// Create session
	session, err := s.createSession(ctx, user.ID, familyID, userRole, userAgent, ipAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("User logged in successfully",
		logger.Str("user_id", user.ID),
		logger.Str("session_id", session.ID),
		logger.Str("family_id", familyID),
	)

	return session, user, nil
}

// ValidateSessionInternal checks if a session is valid and returns session info (for internal use)
func (s *Service) ValidateSessionInternal(ctx context.Context, sessionID string) (*authv1.SessionValidationResult, error) {
	return s.validateSession(ctx, sessionID)
}

// validateSession checks if a session is valid and returns session info
func (s *Service) validateSession(ctx context.Context, sessionID string) (*authv1.SessionValidationResult, error) {
	if sessionID == "" {
		return &authv1.SessionValidationResult{
			Valid: false,
		}, nil
	}

	// Get session from database
	session, err := s.getSession(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &authv1.SessionValidationResult{
				Valid: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		s.deleteSession(ctx, sessionID)
		return &authv1.SessionValidationResult{
			Valid: false,
		}, nil
	}

	// Update last active time
	if err := s.updateSessionActivity(ctx, sessionID); err != nil {
		s.logger.Warn("Failed to update session activity", err,
			logger.Str("session_id", sessionID))
	}

	// Get user info
	user, err := s.getUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &authv1.SessionValidationResult{
		Valid:    true,
		Session:  s.sessionToProto(session),
		User:     s.userToProto(user),
		FamilyId: session.FamilyID,
	}, nil
}

// logout invalidates a session
func (s *Service) logout(ctx context.Context, sessionID string) error {
	if err := s.deleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.Info("User logged out successfully",
		logger.Str("session_id", sessionID))

	return nil
}

// refreshSession extends a session's expiry time
func (s *Service) refreshSession(ctx context.Context, sessionID string) (*masterdb.UserSession, error) {
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
	now := time.Now()
	newExpiresAt := now.Add(24 * time.Hour)

	refreshParams := masterdb.RefreshSessionParams{
		ExpiresAt:  newExpiresAt,
		LastActive: now,
		ID:         sessionID,
	}

	err = s.queries.RefreshSession(ctx, refreshParams)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	session.ExpiresAt = newExpiresAt
	session.LastActive = now

	return session, nil
}

// UpdateUserFamily updates the user's family association when they join/leave a family
func (s *Service) UpdateUserFamily(ctx context.Context, userID, familyID, role string) error {
	// Update all active sessions for this user
	updateParams := masterdb.UpdateUserFamilySessionsParams{
		FamilyID:  familyID,
		UserRole:  role,
		UserID:    userID,
		ExpiresAt: time.Now(), // Only sessions that expire after now
	}

	err := s.queries.UpdateUserFamilySessions(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update user sessions: %w", err)
	}

	s.logger.Info("Updated user family association in sessions",
		logger.Str("user_id", userID),
		logger.Str("family_id", familyID),
		logger.Str("role", role))

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *Service) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	// We can't get rows affected from sqlc's exec method directly, so we'll use the raw DB
	// for this specific case since it needs to return the count
	query := `DELETE FROM user_sessions WHERE expires_at < ?`

	result, err := s.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected > 0 {
		s.logger.Info("Cleaned up expired sessions",
			logger.Int64("sessions_deleted", rowsAffected))
	}

	return rowsAffected, nil
}

// Helper methods

func (s *Service) validateRegisterRequest(req *authv1.RegisterRequest) error {
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
	count, err := s.queries.CheckUserExists(ctx, strings.ToLower(email))
	if err != nil {
		return false, err
	}
	return count > 0, nil
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

func (s *Service) getUserByEmail(ctx context.Context, email string) (*masterdb.User, error) {
	sqlcUser, err := s.queries.GetUserByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}

	return sqlcUser, nil
}

func (s *Service) getUserByID(ctx context.Context, userID string) (*masterdb.User, error) {
	sqlcUser, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sqlcUser, nil
}

func (s *Service) getUserFamilyInfo(ctx context.Context, userID string) (familyID, role string, err error) {
	result, err := s.queries.GetUserFamilyInfo(ctx, &userID)
	if err != nil {
		return "", "", err
	}

	var famID string
	if result.FamilyID != nil {
		famID = *result.FamilyID
	}

	return famID, result.Role, nil
}

func (s *Service) createSession(ctx context.Context, userID, familyID, userRole, userAgent, ipAddress string) (*masterdb.UserSession, error) {
	sessionID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hour sessions

	createParams := masterdb.CreateUserSessionParams{
		ID:         sessionID,
		UserID:     userID,
		FamilyID:   familyID,
		UserRole:   userRole,
		CreatedAt:  now,
		LastActive: now,
		ExpiresAt:  expiresAt,
		UserAgent:  &userAgent,
		IpAddress:  &ipAddress,
	}

	sqlcSession, err := s.queries.CreateUserSession(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return sqlcSession, nil
}

func (s *Service) getSession(ctx context.Context, sessionID string) (*masterdb.UserSession, error) {
	sqlcSession, err := s.queries.GetUserSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return sqlcSession, nil
}

func (s *Service) updateSessionActivity(ctx context.Context, sessionID string) error {
	updateParams := masterdb.UpdateSessionActivityParams{
		LastActive: time.Now(),
		ID:         sessionID,
	}

	return s.queries.UpdateSessionActivity(ctx, updateParams)
}

func (s *Service) deleteSession(ctx context.Context, sessionID string) error {
	return s.queries.DeleteUserSession(ctx, sessionID)
}

// SecureCompare performs a constant-time comparison of two strings
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// Connect interface implementation

// Register implements the Connect AuthServiceHandler interface
func (s *Service) Register(ctx context.Context, req *connect.Request[authv1.RegisterRequest]) (*connect.Response[authv1.RegisterResponse], error) {
	// Call internal business logic
	user, err := s.register(ctx, req.Msg)
	if err != nil {
		// Handle auth errors specifically
		if authErr, ok := err.(*AuthError); ok {
			return connect.NewResponse(&authv1.RegisterResponse{
				Error: &authv1.AuthError{
					Code:    authErr.Code,
					Message: authErr.Message,
				},
			}), nil
		}
		// Return other errors as Connect errors
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert user to protobuf type
	pbUser := s.userToProto(user)

	return connect.NewResponse(&authv1.RegisterResponse{
		User: pbUser,
	}), nil
}

// Login implements the Connect AuthServiceHandler interface
func (s *Service) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	// Extract user agent and IP from metadata (simplified for now)
	userAgent := req.Header().Get("User-Agent")
	// In a real implementation, you'd extract the real IP from headers
	ipAddress := "127.0.0.1"

	// Call internal business logic
	session, user, err := s.login(ctx, req.Msg, userAgent, ipAddress)
	if err != nil {
		// Handle auth errors specifically
		if authErr, ok := err.(*AuthError); ok {
			return connect.NewResponse(&authv1.LoginResponse{
				Error: &authv1.AuthError{
					Code:    authErr.Code,
					Message: authErr.Message,
				},
			}), nil
		}
		// Return other errors as Connect errors
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf types
	pbSession := s.sessionToProto(session)
	pbUser := s.userToProto(user)

	return connect.NewResponse(&authv1.LoginResponse{
		Session: pbSession,
		User:    pbUser,
	}), nil
}

// Logout implements the Connect AuthServiceHandler interface
func (s *Service) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	err := s.logout(ctx, req.Msg.SessionId)
	if err != nil {
		// Handle auth errors specifically
		if authErr, ok := err.(*AuthError); ok {
			return connect.NewResponse(&authv1.LogoutResponse{
				Success: false,
				Error: &authv1.AuthError{
					Code:    authErr.Code,
					Message: authErr.Message,
				},
			}), nil
		}
		// Return other errors as Connect errors
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.LogoutResponse{
		Success: true,
	}), nil
}

// RefreshSession implements the Connect AuthServiceHandler interface
func (s *Service) RefreshSession(ctx context.Context, req *connect.Request[authv1.RefreshSessionRequest]) (*connect.Response[authv1.RefreshSessionResponse], error) {
	session, err := s.refreshSession(ctx, req.Msg.SessionId)
	if err != nil {
		// Handle auth errors specifically
		if authErr, ok := err.(*AuthError); ok {
			return connect.NewResponse(&authv1.RefreshSessionResponse{
				Error: &authv1.AuthError{
					Code:    authErr.Code,
					Message: authErr.Message,
				},
			}), nil
		}
		// Return other errors as Connect errors
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbSession := s.sessionToProto(session)

	return connect.NewResponse(&authv1.RefreshSessionResponse{
		Session: pbSession,
	}), nil
}

// ValidateSession implements the Connect AuthServiceHandler interface
func (s *Service) ValidateSession(ctx context.Context, req *connect.Request[authv1.ValidateSessionRequest]) (*connect.Response[authv1.ValidateSessionResponse], error) {
	result, err := s.validateSession(ctx, req.Msg.SessionId)
	if err != nil {
		// Handle auth errors specifically
		if authErr, ok := err.(*AuthError); ok {
			return connect.NewResponse(&authv1.ValidateSessionResponse{
				Error: &authv1.AuthError{
					Code:    authErr.Code,
					Message: authErr.Message,
				},
			}), nil
		}
		// Return other errors as Connect errors
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.ValidateSessionResponse{
		Result: result,
	}), nil
}

// Type conversion helpers
func (s *Service) userToProto(user *masterdb.User) *authv1.User {
	return &authv1.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}

func (s *Service) sessionToProto(session *masterdb.UserSession) *authv1.Session {
	userAgent := ""
	if session.UserAgent != nil {
		userAgent = *session.UserAgent
	}
	ipAddress := ""
	if session.IpAddress != nil {
		ipAddress = *session.IpAddress
	}
	return &authv1.Session{
		Id:         session.ID,
		UserId:     session.UserID,
		FamilyId:   session.FamilyID,
		UserRole:   session.UserRole,
		CreatedAt:  session.CreatedAt.Unix(),
		LastActive: session.LastActive.Unix(),
		ExpiresAt:  session.ExpiresAt.Unix(),
		UserAgent:  userAgent,
		IpAddress:  ipAddress,
	}
}

