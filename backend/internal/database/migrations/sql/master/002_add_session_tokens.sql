-- Add session_token column to user_sessions table for secure token-based authentication
-- This migration adds a cryptographically secure token column while maintaining the existing ID column for database relations

-- Add the session_token column
ALTER TABLE user_sessions ADD COLUMN session_token TEXT;

-- Create a unique index on session_token for fast lookups and to prevent duplicates
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);

-- Add an index on the existing ID column for continued performance on ID-based queries
CREATE INDEX IF NOT EXISTS idx_user_sessions_id ON user_sessions(id);