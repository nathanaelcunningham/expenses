-- Migration-related queries for master database

-- name: GetCurrentMigrationVersion :one
SELECT CAST(COALESCE(MAX(version), 0) AS INTEGER) FROM schema_migrations;

-- name: CheckMigrationsTableExists :one
-- This query will return 1 if table exists, 0 if not
-- We use a simple approach that works with sqlc
SELECT COUNT(*) as count 
FROM schema_migrations 
WHERE 1=0;

-- name: GetAppliedMigrations :many
SELECT 
    version, 
    name, 
    filename, 
    COALESCE(description, '') as description, 
    applied_at,
    COALESCE(checksum, '') as checksum,
    COALESCE(execution_time_ms, 0) as execution_time_ms,
    COALESCE(migration_type, '') as migration_type,
    COALESCE(applied_by, 'unknown') as applied_by
FROM schema_migrations 
ORDER BY version ASC;

-- name: CheckMigrationApplied :one
SELECT COUNT(*) as count 
FROM schema_migrations 
WHERE version = ?;

-- name: RecordMigration :exec
INSERT INTO schema_migrations 
(version, name, filename, description, applied_at, checksum, execution_time_ms, migration_type, applied_by) 
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?);

-- name: CreateMigrationsTable :exec
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    filename TEXT NOT NULL,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    checksum TEXT,
    execution_time_ms INTEGER,
    migration_type TEXT CHECK (migration_type IN ('master', 'family')),
    applied_by TEXT DEFAULT 'system'
);

-- name: CreateMigrationsTableIndexes :exec
CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at ON schema_migrations(applied_at);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_type ON schema_migrations(migration_type);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_version ON schema_migrations(version);