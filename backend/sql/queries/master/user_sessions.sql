-- name: CreateUserSession :one
INSERT INTO user_sessions (user_id, family_id, user_role, session_token, created_at, last_active, expires_at, user_agent, ip_address)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserSession :one
SELECT * FROM user_sessions WHERE id = ?;

-- name: GetUserSessionByToken :one
SELECT * FROM user_sessions WHERE session_token = ?;

-- name: UpdateSessionActivity :exec
UPDATE user_sessions 
SET last_active = ?
WHERE id = ?;

-- name: UpdateSessionActivityByToken :exec
UPDATE user_sessions 
SET last_active = ?
WHERE session_token = ?;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE id = ?;

-- name: DeleteUserSessionByToken :exec
DELETE FROM user_sessions WHERE session_token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < ?;

-- name: GetUserActiveSessions :many
SELECT * FROM user_sessions 
WHERE user_id = ? AND expires_at > ?
ORDER BY last_active DESC;

-- name: RefreshSession :exec
UPDATE user_sessions 
SET expires_at = ?, last_active = ? 
WHERE id = ?;

-- name: RefreshSessionByToken :exec
UPDATE user_sessions 
SET expires_at = ?, last_active = ? 
WHERE session_token = ?;

-- name: UpdateUserFamilySessions :exec
UPDATE user_sessions 
SET family_id = ?, user_role = ?
WHERE user_id = ? AND expires_at > ?;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < ?;
