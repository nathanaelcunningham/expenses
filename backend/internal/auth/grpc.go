package auth

import (
	"context"
	"crypto/subtle"
	"database/sql"

	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/logger"
	authv1 "expenses-backend/pkg/auth/v1"

	"connectrpc.com/connect"
)

// Service handles authentication operations
type Service struct {
	db      *sql.DB
	queries *masterdb.Queries
	logger  logger.Logger
}

// NewService creates a new authentication service
func NewService(db *sql.DB, queries *masterdb.Queries, log logger.Logger) *Service {
	return &Service{
		db:      db,
		queries: queries,
		logger:  log.With(logger.Str("component", "auth-service")),
	}
}

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

func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
