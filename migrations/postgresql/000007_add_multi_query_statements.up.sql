-- Create query_transaction_statements table
CREATE TABLE IF NOT EXISTS query_transaction_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES query_transactions(id) ON DELETE CASCADE,
    sequence INT NOT NULL,
    query_text TEXT NOT NULL,
    operation_type VARCHAR(20) NOT NULL CHECK (operation_type IN ('SELECT', 'INSERT', 'UPDATE', 'DELETE', 'CREATE_TABLE', 'DROP_TABLE', 'ALTER_TABLE')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed')),
    affected_rows INT DEFAULT 0,
    error_message TEXT,
    preview_data JSONB DEFAULT '[]',
    before_data JSONB DEFAULT '[]',
    after_data JSONB DEFAULT '[]',
    execution_time_ms INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(transaction_id, sequence)
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_statements_transaction_id ON query_transaction_statements(transaction_id);
CREATE INDEX IF NOT EXISTS idx_statements_sequence ON query_transaction_statements(transaction_id, sequence);

-- Add multi-query columns to query_transactions
ALTER TABLE query_transactions 
    ADD COLUMN IF NOT EXISTS is_multi_query BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS statement_count INT DEFAULT 1;
