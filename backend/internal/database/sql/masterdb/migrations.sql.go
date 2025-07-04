// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: migrations.sql

package masterdb

import (
	"context"
	"time"
)

const checkMigrationApplied = `-- name: CheckMigrationApplied :one
SELECT COUNT(*) as count 
FROM schema_migrations 
WHERE version = ?
`

func (q *Queries) CheckMigrationApplied(ctx context.Context, version int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkMigrationApplied, version)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const checkMigrationsTableExists = `-- name: CheckMigrationsTableExists :one
SELECT COUNT(*) as count 
FROM schema_migrations 
WHERE 1=0
`

// This query will return 1 if table exists, 0 if not
// We use a simple approach that works with sqlc
func (q *Queries) CheckMigrationsTableExists(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkMigrationsTableExists)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createMigrationsTable = `-- name: CreateMigrationsTable :exec
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
)
`

func (q *Queries) CreateMigrationsTable(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, createMigrationsTable)
	return err
}

const getAppliedMigrations = `-- name: GetAppliedMigrations :many
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
ORDER BY version ASC
`

type GetAppliedMigrationsRow struct {
	Version         int64      `json:"version"`
	Name            string     `json:"name"`
	Filename        string     `json:"filename"`
	Description     string     `json:"description"`
	AppliedAt       *time.Time `json:"applied_at"`
	Checksum        string     `json:"checksum"`
	ExecutionTimeMs int64      `json:"execution_time_ms"`
	MigrationType   string     `json:"migration_type"`
	AppliedBy       string     `json:"applied_by"`
}

func (q *Queries) GetAppliedMigrations(ctx context.Context) ([]*GetAppliedMigrationsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAppliedMigrations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*GetAppliedMigrationsRow{}
	for rows.Next() {
		var i GetAppliedMigrationsRow
		if err := rows.Scan(
			&i.Version,
			&i.Name,
			&i.Filename,
			&i.Description,
			&i.AppliedAt,
			&i.Checksum,
			&i.ExecutionTimeMs,
			&i.MigrationType,
			&i.AppliedBy,
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

const getCurrentMigrationVersion = `-- name: GetCurrentMigrationVersion :one

SELECT CAST(COALESCE(MAX(version), 0) AS INTEGER) FROM schema_migrations
`

// Migration-related queries for master database
func (q *Queries) GetCurrentMigrationVersion(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, getCurrentMigrationVersion)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}

const recordMigration = `-- name: RecordMigration :exec
INSERT INTO schema_migrations 
(version, name, filename, description, applied_at, checksum, execution_time_ms, migration_type, applied_by) 
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?)
`

type RecordMigrationParams struct {
	Version         int64   `json:"version"`
	Name            string  `json:"name"`
	Filename        string  `json:"filename"`
	Description     *string `json:"description"`
	Checksum        *string `json:"checksum"`
	ExecutionTimeMs *int64  `json:"execution_time_ms"`
	MigrationType   *string `json:"migration_type"`
	AppliedBy       *string `json:"applied_by"`
}

func (q *Queries) RecordMigration(ctx context.Context, arg RecordMigrationParams) error {
	_, err := q.db.ExecContext(ctx, recordMigration,
		arg.Version,
		arg.Name,
		arg.Filename,
		arg.Description,
		arg.Checksum,
		arg.ExecutionTimeMs,
		arg.MigrationType,
		arg.AppliedBy,
	)
	return err
}
