-- Description: Create schema_migrations table for tracking applied migrations in family databases
-- This must be the first migration (version 000) to bootstrap the migration system

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

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at ON schema_migrations(applied_at);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_type ON schema_migrations(migration_type);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_version ON schema_migrations(version);

-- Insert this migration record itself
INSERT OR IGNORE INTO schema_migrations 
(version, name, filename, description, applied_at, migration_type, applied_by) 
VALUES (0, 'migration tracking', '000_migration_tracking.sql', 'Create schema_migrations table for tracking applied migrations', CURRENT_TIMESTAMP, 'family', 'bootstrap');