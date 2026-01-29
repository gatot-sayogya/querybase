-- Drop index
DROP INDEX IF EXISTS idx_data_sources_is_healthy;

-- Remove columns
ALTER TABLE data_sources DROP COLUMN IF EXISTS last_schema_sync;
ALTER TABLE data_sources DROP COLUMN IF EXISTS is_healthy;
ALTER TABLE data_sources DROP COLUMN IF EXISTS last_health_check;
