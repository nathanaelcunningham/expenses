// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: user_sessions.sql

package masterdb

import (
	"context"
	"time"
)

const cleanupExpiredSessions = `-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < ?
`

func (q *Queries) CleanupExpiredSessions(ctx context.Context, expiresAt time.Time) error {
	_, err := q.db.ExecContext(ctx, cleanupExpiredSessions, expiresAt)
	return err
}

const createUserSession = `-- name: CreateUserSession :one
INSERT INTO user_sessions (user_id, family_id, user_role, session_token, created_at, last_active, expires_at, user_agent, ip_address)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address, session_token
`

type CreateUserSessionParams struct {
	UserID       int64     `json:"user_id"`
	FamilyID     int64     `json:"family_id"`
	UserRole     string    `json:"user_role"`
	SessionToken *string   `json:"session_token"`
	CreatedAt    time.Time `json:"created_at"`
	LastActive   time.Time `json:"last_active"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserAgent    *string   `json:"user_agent"`
	IpAddress    *string   `json:"ip_address"`
}

func (q *Queries) CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (*UserSession, error) {
	row := q.db.QueryRowContext(ctx, createUserSession,
		arg.UserID,
		arg.FamilyID,
		arg.UserRole,
		arg.SessionToken,
		arg.CreatedAt,
		arg.LastActive,
		arg.ExpiresAt,
		arg.UserAgent,
		arg.IpAddress,
	)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.FamilyID,
		&i.UserRole,
		&i.CreatedAt,
		&i.LastActive,
		&i.ExpiresAt,
		&i.UserAgent,
		&i.IpAddress,
		&i.SessionToken,
	)
	return &i, err
}

const deleteExpiredSessions = `-- name: DeleteExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < ?
`

func (q *Queries) DeleteExpiredSessions(ctx context.Context, expiresAt time.Time) error {
	_, err := q.db.ExecContext(ctx, deleteExpiredSessions, expiresAt)
	return err
}

const deleteUserSession = `-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE id = ?
`

func (q *Queries) DeleteUserSession(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteUserSession, id)
	return err
}

const deleteUserSessionByToken = `-- name: DeleteUserSessionByToken :exec
DELETE FROM user_sessions WHERE session_token = ?
`

func (q *Queries) DeleteUserSessionByToken(ctx context.Context, sessionToken *string) error {
	_, err := q.db.ExecContext(ctx, deleteUserSessionByToken, sessionToken)
	return err
}

const getUserActiveSessions = `-- name: GetUserActiveSessions :many
SELECT id, user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address, session_token FROM user_sessions 
WHERE user_id = ? AND expires_at > ?
ORDER BY last_active DESC
`

type GetUserActiveSessionsParams struct {
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) GetUserActiveSessions(ctx context.Context, arg GetUserActiveSessionsParams) ([]*UserSession, error) {
	rows, err := q.db.QueryContext(ctx, getUserActiveSessions, arg.UserID, arg.ExpiresAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*UserSession{}
	for rows.Next() {
		var i UserSession
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.FamilyID,
			&i.UserRole,
			&i.CreatedAt,
			&i.LastActive,
			&i.ExpiresAt,
			&i.UserAgent,
			&i.IpAddress,
			&i.SessionToken,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserSession = `-- name: GetUserSession :one
SELECT id, user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address, session_token FROM user_sessions WHERE id = ?
`

func (q *Queries) GetUserSession(ctx context.Context, id int64) (*UserSession, error) {
	row := q.db.QueryRowContext(ctx, getUserSession, id)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.FamilyID,
		&i.UserRole,
		&i.CreatedAt,
		&i.LastActive,
		&i.ExpiresAt,
		&i.UserAgent,
		&i.IpAddress,
		&i.SessionToken,
	)
	return &i, err
}

const getUserSessionByToken = `-- name: GetUserSessionByToken :one
SELECT id, user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address, session_token FROM user_sessions WHERE session_token = ?
`

func (q *Queries) GetUserSessionByToken(ctx context.Context, sessionToken *string) (*UserSession, error) {
	row := q.db.QueryRowContext(ctx, getUserSessionByToken, sessionToken)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.FamilyID,
		&i.UserRole,
		&i.CreatedAt,
		&i.LastActive,
		&i.ExpiresAt,
		&i.UserAgent,
		&i.IpAddress,
		&i.SessionToken,
	)
	return &i, err
}

const refreshSession = `-- name: RefreshSession :exec
UPDATE user_sessions 
SET expires_at = ?, last_active = ? 
WHERE id = ?
`

type RefreshSessionParams struct {
	ExpiresAt  time.Time `json:"expires_at"`
	LastActive time.Time `json:"last_active"`
	ID         int64     `json:"id"`
}

func (q *Queries) RefreshSession(ctx context.Context, arg RefreshSessionParams) error {
	_, err := q.db.ExecContext(ctx, refreshSession, arg.ExpiresAt, arg.LastActive, arg.ID)
	return err
}

const refreshSessionByToken = `-- name: RefreshSessionByToken :exec
UPDATE user_sessions 
SET expires_at = ?, last_active = ? 
WHERE session_token = ?
`

type RefreshSessionByTokenParams struct {
	ExpiresAt    time.Time `json:"expires_at"`
	LastActive   time.Time `json:"last_active"`
	SessionToken *string   `json:"session_token"`
}

func (q *Queries) RefreshSessionByToken(ctx context.Context, arg RefreshSessionByTokenParams) error {
	_, err := q.db.ExecContext(ctx, refreshSessionByToken, arg.ExpiresAt, arg.LastActive, arg.SessionToken)
	return err
}

const updateSessionActivity = `-- name: UpdateSessionActivity :exec
UPDATE user_sessions 
SET last_active = ?
WHERE id = ?
`

type UpdateSessionActivityParams struct {
	LastActive time.Time `json:"last_active"`
	ID         int64     `json:"id"`
}

func (q *Queries) UpdateSessionActivity(ctx context.Context, arg UpdateSessionActivityParams) error {
	_, err := q.db.ExecContext(ctx, updateSessionActivity, arg.LastActive, arg.ID)
	return err
}

const updateSessionActivityByToken = `-- name: UpdateSessionActivityByToken :exec
UPDATE user_sessions 
SET last_active = ?
WHERE session_token = ?
`

type UpdateSessionActivityByTokenParams struct {
	LastActive   time.Time `json:"last_active"`
	SessionToken *string   `json:"session_token"`
}

func (q *Queries) UpdateSessionActivityByToken(ctx context.Context, arg UpdateSessionActivityByTokenParams) error {
	_, err := q.db.ExecContext(ctx, updateSessionActivityByToken, arg.LastActive, arg.SessionToken)
	return err
}

const updateUserFamilySessions = `-- name: UpdateUserFamilySessions :exec
UPDATE user_sessions 
SET family_id = ?, user_role = ?
WHERE user_id = ? AND expires_at > ?
`

type UpdateUserFamilySessionsParams struct {
	FamilyID  int64     `json:"family_id"`
	UserRole  string    `json:"user_role"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) UpdateUserFamilySessions(ctx context.Context, arg UpdateUserFamilySessionsParams) error {
	_, err := q.db.ExecContext(ctx, updateUserFamilySessions,
		arg.FamilyID,
		arg.UserRole,
		arg.UserID,
		arg.ExpiresAt,
	)
	return err
}
