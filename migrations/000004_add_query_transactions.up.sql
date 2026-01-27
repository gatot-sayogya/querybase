-- Migration: Add query_transactions table for transaction-based approval workflow
-- Version: 000004

-- Create query_transactions table
CREATE TABLE IF NOT EXISTS query_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    approval_id UUID NOT NULL UNIQUE REFERENCES approval_requests(id) ON DELETE CASCADE,
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    query_text TEXT NOT NULL,
    started_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    preview_data JSONB,
    affected_rows INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better query performance
CREATE INDEX idx_query_transactions_approval_id ON query_transactions(approval_id);
CREATE INDEX idx_query_transactions_data_source_id ON query_transactions(data_source_id);
CREATE INDEX idx_query_transactions_status ON query_transactions(status);
CREATE INDEX idx_query_transactions_started_by ON query_transactions(started_by);
CREATE INDEX idx_query_transactions_started_at ON query_transactions(started_at);

-- Add comments for documentation
COMMENT ON TABLE query_transactions IS 'Active database transactions for preview in approval workflow';
COMMENT ON COLUMN query_transactions.status IS 'Transaction status: active, committed, rolled_back, failed';
COMMENT ON COLUMN query_transactions.preview_data IS 'JSONB data containing query result preview';
COMMENT ON COLUMN query_transactions.affected_rows IS 'Number of rows affected by the query';
