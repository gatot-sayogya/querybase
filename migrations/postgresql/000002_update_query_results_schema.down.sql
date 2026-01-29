-- Rollback: Revert query_results table back to TEXT[] columns

-- Drop the updated table
DROP TABLE IF EXISTS query_results CASCADE;

-- Recreate with original TEXT[] columns
CREATE TABLE query_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    data JSONB NOT NULL, -- Results as JSON array of objects
    column_names TEXT[] NOT NULL,
    column_types TEXT[] NOT NULL,
    row_count INTEGER NOT NULL,
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    size_bytes INTEGER
);

-- Recreate indexes
CREATE INDEX idx_query_results_query_id ON query_results(query_id);
CREATE INDEX idx_query_results_cached_at ON query_results(cached_at);
