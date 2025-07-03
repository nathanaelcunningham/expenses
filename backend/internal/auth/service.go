package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/family"
	"expenses-backend/internal/logger"
	"expenses-backend/internal/security"
	authv1 "expenses-backend/pkg/auth/v1"

	"golang.org/x/crypto/bcrypt"
)

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

	// Create user
	now := time.Now()
	createParams := masterdb.CreateUserParams{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Name:         strings.TrimSpace(req.Name),
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	sqlcUser, err := s.dbManager.GetMasterQueries().CreateUser(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Follow NOTES.md registration pattern: create family db -> create family -> create family membership
	if req.InviteCode != "" {
		// User is joining an existing family
		if err := s.joinExistingFamily(ctx, sqlcUser, req.InviteCode); err != nil {
			s.logger.Warn("Failed to join family with invite code - user can join later",
				err,
				logger.Int64("user_id", sqlcUser.ID),
				logger.Str("email", sqlcUser.Email),
				logger.Str("invite_code", req.InviteCode),
			)
		}
	} else {
		// Create new family for the user
		if err := s.createNewFamily(ctx, sqlcUser); err != nil {
			s.logger.Warn("Failed to create family for new user - user can create one later",
				err,
				logger.Int64("user_id", sqlcUser.ID),
				logger.Str("email", sqlcUser.Email),
			)
		}
	}

	s.logger.Info("User registered successfully",
		logger.Int64("user_id", sqlcUser.ID),
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
			logger.Int64("user_id", user.ID))
	}

	// Create session
	session, err := s.createSession(ctx, user.ID, familyID, userRole, userAgent, ipAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("User logged in successfully",
		logger.Int64("user_id", user.ID),
		logger.Int64("session_id", session.ID),
		logger.Int64("family_id", familyID),
	)

	return session, user, nil
}

// ValidateSessionInternal checks if a session is valid and returns session info (for internal use)
func (s *Service) ValidateSessionInternal(ctx context.Context, sessionID int64) (*authv1.SessionValidationResult, error) {
	return s.validateSession(ctx, sessionID)
}

// ValidateSessionByToken checks if a session token is valid and returns session info
func (s *Service) ValidateSessionByToken(ctx context.Context, sessionToken string) (*authv1.SessionValidationResult, error) {
	return s.validateSessionByToken(ctx, sessionToken)
}

// validateSession checks if a session is valid and returns session info
func (s *Service) validateSession(ctx context.Context, sessionID int64) (*authv1.SessionValidationResult, error) {
	if sessionID == 0 {
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
			logger.Int64("session_id", sessionID))
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

// validateSessionByToken checks if a session token is valid and returns session info
func (s *Service) validateSessionByToken(ctx context.Context, sessionToken string) (*authv1.SessionValidationResult, error) {
	if sessionToken == "" {
		return &authv1.SessionValidationResult{
			Valid: false,
		}, nil
	}

	// Validate token format
	if err := security.ValidateTokenFormat(sessionToken); err != nil {
		return &authv1.SessionValidationResult{
			Valid: false,
		}, nil
	}

	// Get session from database
	session, err := s.getSessionByToken(ctx, sessionToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return &authv1.SessionValidationResult{
				Valid: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		s.deleteSessionByToken(ctx, sessionToken)
		return &authv1.SessionValidationResult{
			Valid: false,
		}, nil
	}

	// Update last active time
	if err := s.updateSessionActivityByToken(ctx, sessionToken); err != nil {
		s.logger.Warn("Failed to update session activity", err,
			logger.Str("session_token", sessionToken[:8]+"...")) // Log only first 8 chars for security
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
func (s *Service) logout(ctx context.Context, sessionID int64) error {
	if err := s.deleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.Info("User logged out successfully",
		logger.Int64("session_id", sessionID))

	return nil
}

// refreshSession extends a session's expiry time
func (s *Service) refreshSession(ctx context.Context, sessionID int64) (*masterdb.UserSession, error) {
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

	err = s.dbManager.GetMasterQueries().RefreshSession(ctx, refreshParams)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	session.ExpiresAt = newExpiresAt
	session.LastActive = now

	return session, nil
}

// refreshSessionByToken extends a session's expiry time using the session token
func (s *Service) refreshSessionByToken(ctx context.Context, sessionToken string) (*masterdb.UserSession, error) {
	// Get current session
	session, err := s.getSessionByToken(ctx, sessionToken)
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

	refreshParams := masterdb.RefreshSessionByTokenParams{
		ExpiresAt:    newExpiresAt,
		LastActive:   now,
		SessionToken: &sessionToken,
	}

	err = s.dbManager.GetMasterQueries().RefreshSessionByToken(ctx, refreshParams)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	session.ExpiresAt = newExpiresAt
	session.LastActive = now

	return session, nil
}

// UpdateUserFamily updates the user's family association when they join/leave a family
func (s *Service) UpdateUserFamily(ctx context.Context, userID, familyID int64, role string) error {
	// Update all active sessions for this user
	updateParams := masterdb.UpdateUserFamilySessionsParams{
		FamilyID:  familyID,
		UserRole:  role,
		UserID:    userID,
		ExpiresAt: time.Now(), // Only sessions that expire after now
	}

	err := s.dbManager.GetMasterQueries().UpdateUserFamilySessions(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update user sessions: %w", err)
	}

	s.logger.Info("Updated user family association in sessions",
		logger.Int64("user_id", userID),
		logger.Int64("family_id", familyID),
		logger.Str("role", role))

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *Service) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	// We can't get rows affected from sqlc's exec method directly, so we'll use the raw DB
	// for this specific case since it needs to return the count
	query := `DELETE FROM user_sessions WHERE expires_at < ?`

	result, err := s.dbManager.GetMasterDB().ExecContext(ctx, query, time.Now())
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
	count, err := s.dbManager.GetMasterQueries().CheckUserExists(ctx, strings.ToLower(email))
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

func (s *Service) getUserByEmail(ctx context.Context, email string) (*masterdb.User, error) {
	sqlcUser, err := s.dbManager.GetMasterQueries().GetUserByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}

	return sqlcUser, nil
}

func (s *Service) getUserByID(ctx context.Context, userID int64) (*masterdb.User, error) {
	sqlcUser, err := s.dbManager.GetMasterQueries().GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sqlcUser, nil
}

func (s *Service) getUserFamilyInfo(ctx context.Context, userID int64) (familyID int64, role string, err error) {
	result, err := s.dbManager.GetMasterQueries().GetUserFamilyInfo(ctx, &userID)
	if err != nil {
		return 0, "", err
	}

	var famID int64
	if result.FamilyID != nil {
		famID = *result.FamilyID
	}

	return famID, result.Role, nil
}

func (s *Service) createSession(ctx context.Context, userID, familyID int64, userRole, userAgent, ipAddress string) (*masterdb.UserSession, error) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hour sessions

	// Generate secure session token
	sessionToken, err := security.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	createParams := masterdb.CreateUserSessionParams{
		UserID:       userID,
		FamilyID:     familyID,
		UserRole:     userRole,
		SessionToken: &sessionToken,
		CreatedAt:    now,
		LastActive:   now,
		ExpiresAt:    expiresAt,
		UserAgent:    &userAgent,
		IpAddress:    &ipAddress,
	}

	sqlcSession, err := s.dbManager.GetMasterQueries().CreateUserSession(ctx, createParams)
	if err != nil {
		return nil, err
	}

	return sqlcSession, nil
}

func (s *Service) getSession(ctx context.Context, sessionID int64) (*masterdb.UserSession, error) {
	sqlcSession, err := s.dbManager.GetMasterQueries().GetUserSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return sqlcSession, nil
}

func (s *Service) getSessionByToken(ctx context.Context, sessionToken string) (*masterdb.UserSession, error) {
	sqlcSession, err := s.dbManager.GetMasterQueries().GetUserSessionByToken(ctx, &sessionToken)
	if err != nil {
		return nil, err
	}

	return sqlcSession, nil
}

func (s *Service) updateSessionActivity(ctx context.Context, sessionID int64) error {
	updateParams := masterdb.UpdateSessionActivityParams{
		LastActive: time.Now(),
		ID:         sessionID,
	}

	return s.dbManager.GetMasterQueries().UpdateSessionActivity(ctx, updateParams)
}

func (s *Service) updateSessionActivityByToken(ctx context.Context, sessionToken string) error {
	updateParams := masterdb.UpdateSessionActivityByTokenParams{
		LastActive:   time.Now(),
		SessionToken: &sessionToken,
	}

	return s.dbManager.GetMasterQueries().UpdateSessionActivityByToken(ctx, updateParams)
}

func (s *Service) deleteSession(ctx context.Context, sessionID int64) error {
	return s.dbManager.GetMasterQueries().DeleteUserSession(ctx, sessionID)
}

func (s *Service) deleteSessionByToken(ctx context.Context, sessionToken string) error {
	return s.dbManager.GetMasterQueries().DeleteUserSessionByToken(ctx, &sessionToken)
}

// joinExistingFamily handles joining an existing family with invite code
func (s *Service) joinExistingFamily(ctx context.Context, user *masterdb.User, inviteCode string) error {
	joinRequest := family.JoinFamilyRequest{
		InviteCode: inviteCode,
		UserID:     user.ID,
		UserName:   user.Name,
		UserEmail:  user.Email,
	}

	userFamily, err := s.familyService.JoinFamily(ctx, joinRequest)
	if err != nil {
		return fmt.Errorf("failed to join family: %w", err)
	}

	s.logger.Info("User joined family during registration",
		logger.Int64("user_id", user.ID),
		logger.Int64("family_id", userFamily.ID),
		logger.Str("family_name", userFamily.Name),
		logger.Str("invite_code", inviteCode),
	)

	// Update user sessions with family information
	if err := s.UpdateUserFamily(ctx, user.ID, userFamily.ID, "member"); err != nil {
		s.logger.Warn("Failed to update user sessions with family info",
			err,
			logger.Int64("user_id", user.ID),
			logger.Int64("family_id", userFamily.ID),
		)
	}

	return nil
}

func (s *Service) createNewFamily(ctx context.Context, user *masterdb.User) error {
	familyRequest := family.CreateFamilyRequest{
		Name:         fmt.Sprintf("%s's Family", user.Name),
		ManagerID:    user.ID,
		ManagerName:  user.Name,
		ManagerEmail: user.Email,
	}

	userFamily, err := s.familyService.CreateFamily(ctx, familyRequest)
	if err != nil {
		return fmt.Errorf("failed to create family: %w", err)
	}

	s.logger.Info("Family created for new user",
		logger.Int64("user_id", user.ID),
		logger.Int64("family_id", userFamily.ID),
		logger.Str("family_name", userFamily.Name),
	)

	// Update user sessions with family information
	if err := s.UpdateUserFamily(ctx, user.ID, userFamily.ID, "manager"); err != nil {
		s.logger.Warn("Failed to update user sessions with family info",
			err,
			logger.Int64("user_id", user.ID),
			logger.Int64("family_id", userFamily.ID),
		)
	}

	return nil
}
