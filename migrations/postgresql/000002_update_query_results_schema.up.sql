-- Migration: Update query_results table to use JSONB for column metadata
-- This fixes the issue where column names and types need to be stored as JSON
-- instead of text arrays for better compatibility with the Go code

-- Drop existing query_results table and recreate with JSONB columns
DROP TABLE IF EXISTS query_results CASCADE;

-- Recreate query_results with JSONB columns
CREATE TABLE query_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    data JSONB NOT NULL, -- Results as JSON array of objects
    column_names JSONB NOT NULL, -- Column names stored as JSON array
    column_types JSONB NOT NULL, -- Column types stored as JSON array
    row_count INTEGER NOT NULL,
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    size_bytes INTEGER
);

-- Recreate indexes
CREATE INDEX idx_query_results_query_id ON query_results(query_id);
CREATE INDEX idx_query_results_cached_at ON query_results(cached_at);

-- Add comment for documentation
COMMENT ON TABLE query_results IS 'Cached query results with column metadata stored as JSONB';
COMMENT ON COLUMN query_results.column_names IS 'Column names stored as JSONB array for compatibility with Go JSON serialization';
COMMENT ON COLUMN query_results.column_types IS 'Column types stored as JSONB array for compatibility with Go JSON serialization';
