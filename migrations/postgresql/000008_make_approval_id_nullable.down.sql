-- Migration: Revert approval_id to NOT NULL (down migration)
-- Version: 000008

-- First, we need to handle any NULL values - delete them or set them to a dummy value
-- Since we can't have NULL values with NOT NULL constraint, we'll delete orphaned transactions
DELETE FROM query_transactions WHERE approval_id IS NULL;

-- Drop the foreign key constraint
ALTER TABLE query_transactions DROP CONSTRAINT IF EXISTS query_transactions_approval_id_fkey;

-- Make approval_id NOT NULL again
ALTER TABLE query_transactions ALTER COLUMN approval_id SET NOT NULL;

-- Recreate foreign key constraint
ALTER TABLE query_transactions ADD CONSTRAINT query_transactions_approval_id_fkey 
    FOREIGN KEY (approval_id) REFERENCES approval_requests(id) ON DELETE CASCADE;
