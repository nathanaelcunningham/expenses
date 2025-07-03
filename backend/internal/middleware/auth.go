package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"expenses-backend/internal/auth"
	appcontext "expenses-backend/internal/context"
	"expenses-backend/internal/database"
	"expenses-backend/internal/logger"

	"connectrpc.com/connect"
)

// AuthInterceptor provides session-based authentication for Connect RPC
type AuthInterceptor struct {
	authService *auth.Service
	dbManager   *database.DatabaseManager
	logger      logger.Logger
}

// ContextKey is used for storing auth context in request context
type ContextKey string

const (
	SessionHeader ContextKey = "Authorization"
)

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(authService *auth.Service, dbManager *database.DatabaseManager, log logger.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
		dbManager:   dbManager,
		logger:      log.With(logger.Str("component", "auth-interceptor")),
	}
}

// WrapUnary creates a unary interceptor for authentication
func (ai *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Check if this endpoint requires authentication
		if ai.isPublicEndpoint(req.Spec().Procedure) {
			return next(ctx, req)
		}

		// Extract and validate session
		authCtx, err := ai.authenticateRequest(ctx, req.Header())
		if err != nil {
			ai.logger.Warn("Authentication failed", err,
				logger.Str("procedure", req.Spec().Procedure))
			return nil, err
		}

		// Add auth context to request context
		ctx = context.WithValue(ctx, appcontext.AuthContextKey, authCtx)

		ai.logger.Debug("Request authenticated successfully",
			logger.Int64("user_id", authCtx.UserID),
			logger.Int64("family_id", authCtx.FamilyID),
			logger.Str("procedure", req.Spec().Procedure))

		return next(ctx, req)
	}
}

// WrapStreamingClient creates a streaming client interceptor for authentication
func (ai *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// Add session token to outgoing requests if available
		return next(ctx, spec)
	}
}

// WrapStreamingHandler creates a streaming handler interceptor for authentication
func (ai *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// Check if this endpoint requires authentication
		if ai.isPublicEndpoint(conn.Spec().Procedure) {
			return next(ctx, conn)
		}

		// Extract and validate session
		authCtx, err := ai.authenticateRequest(ctx, conn.RequestHeader())
		if err != nil {
			ai.logger.Warn("Authentication failed for streaming connection", err,
				logger.Str("procedure", conn.Spec().Procedure))
			return err
		}

		// Add auth context to request context
		ctx = context.WithValue(ctx, appcontext.AuthContextKey, authCtx)

		ai.logger.Debug("Streaming connection authenticated successfully",
			logger.Int64("user_id", authCtx.UserID),
			logger.Int64("family_id", authCtx.FamilyID),
			logger.Str("procedure", conn.Spec().Procedure))

		return next(ctx, conn)
	}
}

// authenticateRequest extracts and validates the session from request headers
func (ai *AuthInterceptor) authenticateRequest(ctx context.Context, headers http.Header) (*appcontext.AuthContext, error) {
	// Extract session token from Authorization header
	sessionToken := ai.extractSessionToken(headers)
	if sessionToken == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated,
			fmt.Errorf("missing or invalid session token"))
	}

	// Try token-based authentication first (new secure method)
	validation, err := ai.authService.ValidateSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("failed to validate session: %w", err))
	}

	// If token-based validation fails, try legacy ID-based validation for backward compatibility
	if !validation.Valid {
		// Check if it's a numeric ID (legacy format)
		if sessionID, parseErr := strconv.ParseInt(sessionToken, 10, 64); parseErr == nil {
			validation, err = ai.authService.ValidateSessionInternal(ctx, sessionID)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal,
					fmt.Errorf("failed to validate session: %w", err))
			}
		}
	}

	if !validation.Valid {
		return nil, connect.NewError(connect.CodeUnauthenticated,
			fmt.Errorf("invalid or expired session"))
	}

	// Get family database connection if user has a family
	var familyDB *sql.DB
	if validation.FamilyId != 0 {
		familyDB, err = ai.dbManager.GetFamilyDatabase(ctx, int(validation.FamilyId))
		if err != nil {
			ai.logger.Error("Failed to get family database connection", err,
				logger.Int64("family_id", validation.FamilyId))
			return nil, connect.NewError(connect.CodeInternal,
				fmt.Errorf("failed to access family database: %w", err))
		}
	}

	return &appcontext.AuthContext{
		UserID:    validation.User.Id,
		FamilyID:  validation.FamilyId,
		UserRole:  validation.Session.UserRole,
		SessionID: validation.Session.Id,
		FamilyDB:  familyDB,
	}, nil
}

// extractSessionToken extracts the session token from request headers
func (ai *AuthInterceptor) extractSessionToken(headers http.Header) string {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Support both "Bearer <token>" and just "<token>" formats
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		return after
	}

	return authHeader
}

// isPublicEndpoint checks if an endpoint doesn't require authentication
func (ai *AuthInterceptor) isPublicEndpoint(procedure string) bool {
	publicEndpoints := []string{
		"/auth.v1.AuthService/Register",
		"/auth.v1.AuthService/Login",
		"/health.v1.HealthService/Check",
	}

	return slices.Contains(publicEndpoints, procedure)
}


// LoggingInterceptor provides request logging for Connect RPC
type LoggingInterceptor struct {
	logger logger.Logger
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(log logger.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: log.With(logger.Str("component", "rpc-logger")),
	}
}

// WrapUnary creates a unary interceptor for logging
func (li *LoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		// Get auth context if available
		var userID, familyID int64
		if authCtx, ok := appcontext.GetAuthContext(ctx); ok {
			userID = authCtx.UserID
			familyID = authCtx.FamilyID
		}

		// Execute request
		resp, err := next(ctx, req)

		duration := time.Since(start)

		// Log the request
		if err != nil {
			li.logger.Error("RPC request completed", err,
				logger.Str("procedure", req.Spec().Procedure),
				logger.Int64("user_id", userID),
				logger.Int64("family_id", familyID),
				logger.Duration("duration", duration))
		} else {
			li.logger.Info("RPC request completed",
				logger.Str("procedure", req.Spec().Procedure),
				logger.Int64("user_id", userID),
				logger.Int64("family_id", familyID),
				logger.Duration("duration", duration))
		}

		return resp, err
	}
}

// WrapStreamingClient creates a streaming client interceptor for logging
func (li *LoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		start := time.Now()
		conn := next(ctx, spec)

		li.logger.Debug("RPC streaming client connection established",
			logger.Str("procedure", spec.Procedure),
			logger.Duration("setup_duration", time.Since(start)))

		return conn
	}
}

// WrapStreamingHandler creates a streaming handler interceptor for logging
func (li *LoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()

		// Get auth context if available
		var userID, familyID int64
		if authCtx, ok := appcontext.GetAuthContext(ctx); ok {
			userID = authCtx.UserID
			familyID = authCtx.FamilyID
		}

		// Execute streaming connection
		err := next(ctx, conn)

		duration := time.Since(start)

		// Log the streaming connection
		if err != nil {
			li.logger.Error("RPC streaming handler completed", err,
				logger.Str("procedure", conn.Spec().Procedure),
				logger.Int64("user_id", userID),
				logger.Int64("family_id", familyID),
				logger.Duration("duration", duration))
		} else {
			li.logger.Info("RPC streaming handler completed",
				logger.Str("procedure", conn.Spec().Procedure),
				logger.Int64("user_id", userID),
				logger.Int64("family_id", familyID),
				logger.Duration("duration", duration))
		}

		return err
	}
}
