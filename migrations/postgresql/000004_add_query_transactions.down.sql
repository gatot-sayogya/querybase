-- Migration Rollback: Remove query_transactions table
-- Version: 000004

-- Drop the table
DROP TABLE IF EXISTS query_transactions CASCADE;
