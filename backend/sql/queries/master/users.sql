-- name: CreateUser :one
INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: UpdateUser :one
UPDATE users 
SET name = ?, password_hash = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;