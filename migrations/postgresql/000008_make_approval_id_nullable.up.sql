-- Migration: Make approval_id nullable in query_transactions to support direct multi-query execution
-- Version: 000008

-- Drop the NOT NULL constraint and foreign key constraint on approval_id
ALTER TABLE query_transactions DROP CONSTRAINT IF EXISTS query_transactions_approval_id_fkey;

-- Make approval_id nullable
ALTER TABLE query_transactions ALTER COLUMN approval_id DROP NOT NULL;

-- Recreate foreign key constraint (nullable)
ALTER TABLE query_transactions ADD CONSTRAINT query_transactions_approval_id_fkey 
    FOREIGN KEY (approval_id) REFERENCES approval_requests(id) ON DELETE CASCADE;

-- Keep the unique index but allow NULL values (PostgreSQL allows multiple NULLs in unique index)
-- The unique constraint is already on the column, no change needed

COMMENT ON COLUMN query_transactions.approval_id IS 'Reference to approval request (nullable for direct multi-query execution without approval)';
