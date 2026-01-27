-- Rollback: Recreate query_results with caching columns

DROP TABLE IF EXISTS query_results CASCADE;

CREATE TABLE query_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    column_names JSONB NOT NULL,
    column_types JSONB NOT NULL,
    row_count INTEGER NOT NULL,
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    size_bytes INTEGER
);

CREATE INDEX idx_query_results_query_id ON query_results(query_id);
CREATE INDEX idx_query_results_cached_at ON query_results(cached_at);
