-- Migration: Add audit execution support for approval workflow
-- Version: 000006

-- Add audit columns to query_transactions table
ALTER TABLE query_transactions
    ADD COLUMN IF NOT EXISTS audit_mode VARCHAR(20) NOT NULL DEFAULT 'count_only',
    ADD COLUMN IF NOT EXISTS before_data JSONB,
    ADD COLUMN IF NOT EXISTS after_data JSONB,
    ADD COLUMN IF NOT EXISTS estimated_rows INTEGER NOT NULL DEFAULT 0;

-- Add audit configuration to data_sources table
ALTER TABLE data_sources
    ADD COLUMN IF NOT EXISTS audit_row_threshold INTEGER NOT NULL DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS audit_capability VARCHAR(20) NOT NULL DEFAULT 'unknown';

-- Add comments for documentation
COMMENT ON COLUMN query_transactions.audit_mode IS 'Audit capture mode: full, sample, count_only';
COMMENT ON COLUMN query_transactions.before_data IS 'JSONB snapshot of data before query execution';
COMMENT ON COLUMN query_transactions.after_data IS 'JSONB snapshot of data after query execution';
COMMENT ON COLUMN query_transactions.estimated_rows IS 'Estimated number of rows affected (pre-execution count)';
COMMENT ON COLUMN data_sources.audit_row_threshold IS 'Row count threshold that triggers large-change caution warning';
COMMENT ON COLUMN data_sources.audit_capability IS 'DDL audit capability: full, count_only, unknown';
