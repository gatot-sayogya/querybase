-- ============================================
-- Drop Schema for QueryBase (Rollback)
-- ============================================

-- Drop views
DROP VIEW IF EXISTS pending_approvals;
DROP VIEW IF EXISTS user_permissions;

-- Drop triggers
DROP TRIGGER IF EXISTS update_notification_configs_updated_at ON notification_configs;
DROP TRIGGER IF EXISTS update_approval_requests_updated_at ON approval_requests;
DROP TRIGGER IF EXISTS update_queries_updated_at ON queries;
DROP TRIGGER IF EXISTS update_data_sources_updated_at ON data_sources;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (they will be dropped with tables, but being explicit)
DROP INDEX IF EXISTS idx_notifications_approval_request_id;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_approval_reviews_reviewed_by;
DROP INDEX IF EXISTS idx_approval_reviews_approval_request_id;
DROP INDEX IF EXISTS idx_approval_requests_created_at;
DROP INDEX IF EXISTS idx_approval_requests_status;
DROP INDEX IF EXISTS idx_approval_requests_requested_by;
DROP INDEX IF EXISTS idx_query_history_executed_at;
DROP INDEX IF EXISTS idx_query_history_data_source_id;
DROP INDEX IF EXISTS idx_query_history_user_id;
DROP INDEX IF EXISTS idx_query_results_cached_at;
DROP INDEX IF EXISTS idx_query_results_query_id;
DROP INDEX IF EXISTS idx_queries_requires_approval;
DROP INDEX IF EXISTS idx_queries_created_at;
DROP INDEX IF EXISTS idx_queries_status;
DROP INDEX IF EXISTS idx_queries_user_id;
DROP INDEX IF EXISTS idx_queries_data_source_id;
DROP INDEX IF EXISTS idx_data_source_permissions_group_id;
DROP INDEX IF EXISTS idx_data_source_permissions_data_source_id;
DROP INDEX IF EXISTS idx_data_sources_is_active;
DROP INDEX IF EXISTS idx_data_sources_type;
DROP INDEX IF EXISTS idx_user_groups_group_id;
DROP INDEX IF EXISTS idx_user_groups_user_id;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (in correct order respecting foreign keys)
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS notification_configs CASCADE;
DROP TABLE IF EXISTS approval_reviews CASCADE;
DROP TABLE IF EXISTS approval_requests CASCADE;
DROP TABLE IF EXISTS query_history CASCADE;
DROP TABLE IF EXISTS query_results CASCADE;
DROP TABLE IF EXISTS queries CASCADE;
DROP TABLE IF EXISTS data_source_permissions CASCADE;
DROP TABLE IF EXISTS data_sources CASCADE;
DROP TABLE IF EXISTS user_groups CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS groups CASCADE;

-- Drop enums
DROP TYPE IF EXISTS notification_type CASCADE;
DROP TYPE IF EXISTS notification_status CASCADE;
DROP TYPE IF EXISTS approval_status CASCADE;
DROP TYPE IF EXISTS operation_type CASCADE;
DROP TYPE IF EXISTS query_status CASCADE;
DROP TYPE IF EXISTS data_source_type CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;

-- Note: We don't drop the uuid-ossp extension as it may be used by other tables
-- Uncomment the line below if you want to drop the extension:
-- DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
