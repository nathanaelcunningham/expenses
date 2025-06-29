-- name: CreateFamily :one
INSERT INTO families (name, invite_code, database_url, manager_id, schema_version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetFamilyByID :one
SELECT * FROM families WHERE id = ?;

-- name: GetFamilyByInviteCode :one
SELECT * FROM families WHERE invite_code = ?;

-- name: UpdateFamily :one
UPDATE families 
SET name = ?, database_url = ?, schema_version = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteFamily :exec
DELETE FROM families WHERE id = ?;
