package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"expenses-backend/internal/auth"
	"expenses-backend/internal/database"

	"connectrpc.com/connect"
	"github.com/rs/zerolog"
)

// AuthInterceptor provides session-based authentication for Connect RPC
type AuthInterceptor struct {
	authService *auth.Service
	dbManager   *database.Manager
	logger      zerolog.Logger
}

// AuthContext holds authentication information for the request
type AuthContext struct {
	UserID     string `json:"user_id"`
	FamilyID   string `json:"family_id"`
	UserRole   string `json:"user_role"`
	SessionID  string `json:"session_id"`
	FamilyDB   *sql.DB `json:"-"`
}

// ContextKey is used for storing auth context in request context
type ContextKey string

const (
	AuthContextKey ContextKey = "auth_context"
	SessionHeader            = "Authorization"
)

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(authService *auth.Service, dbManager *database.Manager, logger zerolog.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
		dbManager:   dbManager,
		logger:      logger.With().Str("component", "auth-interceptor").Logger(),
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
			ai.logger.Warn().
				Err(err).
				Str("procedure", req.Spec().Procedure).
				Msg("Authentication failed")
			return nil, err
		}

		// Add auth context to request context
		ctx = context.WithValue(ctx, AuthContextKey, authCtx)

		ai.logger.Debug().
			Str("user_id", authCtx.UserID).
			Str("family_id", authCtx.FamilyID).
			Str("procedure", req.Spec().Procedure).
			Msg("Request authenticated successfully")

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
			ai.logger.Warn().
				Err(err).
				Str("procedure", conn.Spec().Procedure).
				Msg("Authentication failed for streaming connection")
			return err
		}

		// Add auth context to request context
		ctx = context.WithValue(ctx, AuthContextKey, authCtx)

		ai.logger.Debug().
			Str("user_id", authCtx.UserID).
			Str("family_id", authCtx.FamilyID).
			Str("procedure", conn.Spec().Procedure).
			Msg("Streaming connection authenticated successfully")

		return next(ctx, conn)
	}
}

// authenticateRequest extracts and validates the session from request headers
func (ai *AuthInterceptor) authenticateRequest(ctx context.Context, headers http.Header) (*AuthContext, error) {
	// Extract session token from Authorization header
	sessionToken := ai.extractSessionToken(headers)
	if sessionToken == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, 
			fmt.Errorf("missing or invalid session token"))
	}

	// Validate session
	validation, err := ai.authService.ValidateSession(ctx, sessionToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, 
			fmt.Errorf("failed to validate session: %w", err))
	}

	if !validation.Valid {
		return nil, connect.NewError(connect.CodeUnauthenticated, 
			fmt.Errorf("invalid or expired session"))
	}

	// Get family database connection if user has a family
	var familyDB *sql.DB
	if validation.FamilyID != "" {
		familyDB, err = ai.dbManager.GetFamilyDatabase(ctx, validation.FamilyID)
		if err != nil {
			ai.logger.Error().
				Err(err).
				Str("family_id", validation.FamilyID).
				Msg("Failed to get family database connection")
			return nil, connect.NewError(connect.CodeInternal, 
				fmt.Errorf("failed to access family database: %w", err))
		}
	}

	return &AuthContext{
		UserID:    validation.User.ID,
		FamilyID:  validation.FamilyID,
		UserRole:  validation.Session.UserRole,
		SessionID: validation.Session.ID,
		FamilyDB:  familyDB,
	}, nil
}

// extractSessionToken extracts the session token from request headers
func (ai *AuthInterceptor) extractSessionToken(headers http.Header) string {
	authHeader := headers.Get(SessionHeader)
	if authHeader == "" {
		return ""
	}

	// Support both "Bearer <token>" and just "<token>" formats
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
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

	for _, endpoint := range publicEndpoints {
		if procedure == endpoint {
			return true
		}
	}

	return false
}

// GetAuthContext extracts the authentication context from the request context
func GetAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(AuthContextKey).(*AuthContext)
	return authCtx, ok
}

// RequireAuth is a helper function that returns an error if no auth context is found
func RequireAuth(ctx context.Context) (*AuthContext, error) {
	authCtx, ok := GetAuthContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, 
			fmt.Errorf("authentication required"))
	}
	return authCtx, nil
}

// RequireFamily is a helper function that returns an error if user is not in a family
func RequireFamily(ctx context.Context) (*AuthContext, error) {
	authCtx, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if authCtx.FamilyID == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, 
			fmt.Errorf("user must be a member of a family to access this resource"))
	}

	return authCtx, nil
}

// RequireFamilyManager is a helper function that returns an error if user is not a family manager
func RequireFamilyManager(ctx context.Context) (*AuthContext, error) {
	authCtx, err := RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	if authCtx.UserRole != "manager" {
		return nil, connect.NewError(connect.CodePermissionDenied, 
			fmt.Errorf("only family managers can perform this action"))
	}

	return authCtx, nil
}

// LoggingInterceptor provides request logging for Connect RPC
type LoggingInterceptor struct {
	logger zerolog.Logger
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(logger zerolog.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger.With().Str("component", "rpc-logger").Logger(),
	}
}

// WrapUnary creates a unary interceptor for logging
func (li *LoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()
		
		// Get auth context if available
		var userID, familyID string
		if authCtx, ok := GetAuthContext(ctx); ok {
			userID = authCtx.UserID
			familyID = authCtx.FamilyID
		}

		// Execute request
		resp, err := next(ctx, req)
		
		duration := time.Since(start)
		
		// Log the request
		logEvent := li.logger.Info()
		if err != nil {
			logEvent = li.logger.Error().Err(err)
		}
		
		logEvent.
			Str("procedure", req.Spec().Procedure).
			Str("user_id", userID).
			Str("family_id", familyID).
			Dur("duration", duration).
			Msg("RPC request completed")

		return resp, err
	}
}

// WrapStreamingClient creates a streaming client interceptor for logging
func (li *LoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		start := time.Now()
		conn := next(ctx, spec)
		
		li.logger.Debug().
			Str("procedure", spec.Procedure).
			Dur("setup_duration", time.Since(start)).
			Msg("RPC streaming client connection established")
		
		return conn
	}
}

// WrapStreamingHandler creates a streaming handler interceptor for logging
func (li *LoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()
		
		// Get auth context if available
		var userID, familyID string
		if authCtx, ok := GetAuthContext(ctx); ok {
			userID = authCtx.UserID
			familyID = authCtx.FamilyID
		}

		// Execute streaming connection
		err := next(ctx, conn)
		
		duration := time.Since(start)
		
		// Log the streaming connection
		logEvent := li.logger.Info()
		if err != nil {
			logEvent = li.logger.Error().Err(err)
		}
		
		logEvent.
			Str("procedure", conn.Spec().Procedure).
			Str("user_id", userID).
			Str("family_id", familyID).
			Dur("duration", duration).
			Msg("RPC streaming handler completed")

		return err
	}
}