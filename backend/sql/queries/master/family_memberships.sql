-- name: CreateFamilyMembership :one
INSERT INTO family_memberships (family_id, user_id, role, joined_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetFamilyMembership :one
SELECT * FROM family_memberships WHERE family_id = ? AND user_id = ?;

-- name: ListFamilyMemberships :many
SELECT * FROM family_memberships WHERE family_id = ?;

-- name: ListUserMemberships :many
SELECT * FROM family_memberships WHERE user_id = ?;

-- name: UpdateFamilyMembershipRole :one
UPDATE family_memberships 
SET role = ?
WHERE family_id = ? AND user_id = ?
RETURNING *;

-- name: DeleteFamilyMembership :exec
DELETE FROM family_memberships WHERE family_id = ? AND user_id = ?;

-- name: GetUserFamilyInfo :one
SELECT family_id, role FROM family_memberships WHERE user_id = ?;