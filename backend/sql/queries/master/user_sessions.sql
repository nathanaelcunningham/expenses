-- name: CreateUserSession :one
INSERT INTO user_sessions (user_id, family_id, user_role, created_at, last_active, expires_at, user_agent, ip_address)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserSession :one
SELECT * FROM user_sessions WHERE id = ?;

-- name: UpdateSessionActivity :exec
UPDATE user_sessions 
SET last_active = ?
WHERE id = ?;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE id = ?;

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

-- name: UpdateUserFamilySessions :exec
UPDATE user_sessions 
SET family_id = ?, user_role = ?
WHERE user_id = ? AND expires_at > ?;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < ?;
