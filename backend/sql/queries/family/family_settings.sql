-- name: CreateFamilySetting :one 
INSERT INTO family_settings (setting_key, setting_value, data_type)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListFamilySettings :many
SELECT * FROM family_settings;

-- name: GetFamilySettingByKey :one
SELECT * FROM family_settings
WHERE setting_key = ?;

-- name: UpdateFamilySetting :one
UPDATE family_settings
SET setting_value = ?, data_type = ?
WHERE id = ?
RETURNING *;

-- name: DeleteFamilySetting :exec
DELETE FROM family_settings WHERE id = ?;
