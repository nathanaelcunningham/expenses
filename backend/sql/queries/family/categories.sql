-- name: CreateCategory :one
INSERT INTO categories (id, name, description, color, icon, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = ?;

-- name: ListCategories :many
SELECT * FROM categories ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE categories 
SET name = ?, description = ?, color = ?, icon = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;