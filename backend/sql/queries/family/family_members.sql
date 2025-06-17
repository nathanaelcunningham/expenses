-- name: CreateFamilyMember :one
INSERT INTO family_members (id, name, email, role, joined_at, is_active)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetFamilyMemberByID :one
SELECT * FROM family_members WHERE id = ?;

-- name: GetFamilyMemberByEmail :one
SELECT * FROM family_members WHERE email = ?;

-- name: ListFamilyMembers :many
SELECT * FROM family_members 
WHERE is_active = TRUE
ORDER BY name ASC;

-- name: ListAllFamilyMembers :many
SELECT * FROM family_members ORDER BY name ASC;

-- name: UpdateFamilyMember :one
UPDATE family_members 
SET name = ?, email = ?, role = ?
WHERE id = ?
RETURNING *;

-- name: DeactivateFamilyMember :exec
UPDATE family_members 
SET is_active = FALSE
WHERE id = ?;

-- name: DeleteFamilyMember :exec
DELETE FROM family_members WHERE id = ?;