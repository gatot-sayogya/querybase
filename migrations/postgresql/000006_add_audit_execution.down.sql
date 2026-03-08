-- Rollback: Remove audit execution support
-- Version: 000006

-- Remove audit columns from query_transactions
ALTER TABLE query_transactions
    DROP COLUMN IF EXISTS audit_mode,
    DROP COLUMN IF EXISTS before_data,
    DROP COLUMN IF EXISTS after_data,
    DROP COLUMN IF EXISTS estimated_rows;

-- Remove audit configuration from data_sources
ALTER TABLE data_sources
    DROP COLUMN IF EXISTS audit_row_threshold,
    DROP COLUMN IF EXISTS audit_capability;
