-- Remove caching-specific columns and rename for result storage (MySQL)

-- Drop existing query_results table
DROP TABLE IF EXISTS query_results CASCADE;

-- Recreate without caching-specific columns
CREATE TABLE query_results (
    id CHAR(36) PRIMARY KEY,
    query_id CHAR(36) NOT NULL,
    data JSON NOT NULL,
    column_names JSON NOT NULL,
    column_types JSON NOT NULL,
    row_count INT NOT NULL,
    stored_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    size_bytes INT,
    FOREIGN KEY (query_id) REFERENCES queries(id) ON DELETE CASCADE,
    INDEX idx_query_results_query_id (query_id),
    INDEX idx_query_results_stored_at (stored_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

COMMENT ON TABLE query_results IS 'Stored query results for result history (not a cache)';
