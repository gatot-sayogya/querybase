-- Drop indexes
DROP INDEX IF EXISTS idx_approval_comments_approval_request_id;
DROP INDEX IF EXISTS idx_approval_comments_user_id;
DROP INDEX IF EXISTS idx_approval_comments_created_at;

-- Drop approval_comments table
DROP TABLE IF EXISTS approval_comments;
