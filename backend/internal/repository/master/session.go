package master

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/models"
	"expenses-backend/internal/repository/interfaces"
	"time"
)

type sessionRepository struct {
	queries *masterdb.Queries
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) interfaces.SessionRepository {
	return &sessionRepository{
		queries: masterdb.New(db),
	}
}

// NewSessionRepositoryWithTx creates a new session repository with a transaction
func NewSessionRepositoryWithTx(tx *sql.Tx) interfaces.SessionRepository {
	return &sessionRepository{
		queries: masterdb.New(tx),
	}
}

// WithTx returns a new session repository using the provided transaction
func (r *sessionRepository) WithTx(tx *sql.Tx) interface{} {
	return &sessionRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *models.UserSession) (*models.UserSession, error) {
	var familyID, userRole string
	if session.FamilyID != nil {
		familyID = *session.FamilyID
	}
	if session.UserRole != nil {
		userRole = *session.UserRole
	}

	params := masterdb.CreateUserSessionParams{
		ID:         session.ID,
		UserID:     session.UserID,
		FamilyID:   familyID,
		UserRole:   userRole,
		CreatedAt:  session.CreatedAt,
		LastActive: session.LastActive,
		ExpiresAt:  session.ExpiresAt,
		UserAgent:  session.UserAgent,
		IpAddress:  session.IPAddress,
	}

	result, err := r.queries.CreateUserSession(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCSessionToModel(result), nil
}

func (r *sessionRepository) GetSession(ctx context.Context, id string) (*models.UserSession, error) {
	result, err := r.queries.GetUserSession(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertSQLCSessionToModel(result), nil
}

func (r *sessionRepository) GetUserActiveSessions(ctx context.Context, userID string, limit int) ([]*models.UserSession, error) {
	params := masterdb.GetUserActiveSessionsParams{
		UserID:    userID,
		ExpiresAt: time.Now(),
	}

	results, err := r.queries.GetUserActiveSessions(ctx, params)
	if err != nil {
		return nil, err
	}

	// Apply limit manually since it's not in the SQL query
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	sessions := make([]*models.UserSession, len(results))
	for i, result := range results {
		sessions[i] = convertSQLCSessionToModel(result)
	}

	return sessions, nil
}

func (r *sessionRepository) UpdateSessionActivity(ctx context.Context, id string) error {
	params := masterdb.UpdateSessionActivityParams{
		ID:         id,
		LastActive: time.Now(),
	}

	return r.queries.UpdateSessionActivity(ctx, params)
}

func (r *sessionRepository) DeleteSession(ctx context.Context, id string) error {
	return r.queries.DeleteUserSession(ctx, id)
}

func (r *sessionRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	err := r.queries.DeleteExpiredSessions(ctx, time.Now())
	if err != nil {
		return 0, err
	}
	// Note: SQLite doesn't return affected rows for DELETE, so we return 0
	// In a production system, you might want to count before deleting
	return 0, nil
}

// Helper functions
func convertSQLCSessionToModel(session *masterdb.UserSession) *models.UserSession {
	var familyID, userRole *string
	if session.FamilyID != "" {
		familyID = &session.FamilyID
	}
	if session.UserRole != "" {
		userRole = &session.UserRole
	}

	return &models.UserSession{
		ID:         session.ID,
		UserID:     session.UserID,
		FamilyID:   familyID,
		UserRole:   userRole,
		CreatedAt:  session.CreatedAt,
		LastActive: session.LastActive,
		ExpiresAt:  session.ExpiresAt,
		UserAgent:  session.UserAgent,
		IPAddress:  session.IpAddress,
	}
}