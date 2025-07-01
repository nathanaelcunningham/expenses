package context

import (
	"context"
	"database/sql"
	"fmt"

	"connectrpc.com/connect"
)

// AuthContext holds authentication information for the request
type AuthContext struct {
	UserID    int64   `json:"user_id"`
	FamilyID  int64   `json:"family_id"`
	UserRole  string  `json:"user_role"`
	SessionID int64   `json:"session_id"`
	FamilyDB  *sql.DB `json:"-"`
}

// ContextKey is used for storing auth context in request context
type ContextKey string

const (
	AuthContextKey ContextKey = "auth_context"
)

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

	if authCtx.FamilyID == 0 {
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