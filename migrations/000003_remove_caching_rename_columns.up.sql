-- Remove caching-specific columns and rename for result storage

-- Drop existing query_results table
DROP TABLE IF EXISTS query_results CASCADE;

-- Recreate without caching-specific columns
CREATE TABLE query_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    data JSONB NOT NULL, -- Results as JSON array of objects
    column_names JSONB NOT NULL, -- Column names stored as JSON array
    column_types JSONB NOT NULL, -- Column types stored as JSON array
    row_count INTEGER NOT NULL,
    stored_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    size_bytes INTEGER
);

-- Recreate indexes
CREATE INDEX idx_query_results_query_id ON query_results(query_id);
CREATE INDEX idx_query_results_stored_at ON query_results(stored_at);

COMMENT ON TABLE query_results IS 'Stored query results for result history (not a cache)';
COMMENT ON COLUMN query_results.stored_at IS 'When the result was stored';
