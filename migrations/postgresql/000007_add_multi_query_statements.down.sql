-- Drop indexes
DROP INDEX IF EXISTS idx_statements_sequence;
DROP INDEX IF EXISTS idx_statements_transaction_id;

-- Drop table
DROP TABLE IF EXISTS query_transaction_statements;

-- Remove columns from query_transactions
ALTER TABLE query_transactions 
    DROP COLUMN IF EXISTS is_multi_query,
    DROP COLUMN IF EXISTS statement_count;
