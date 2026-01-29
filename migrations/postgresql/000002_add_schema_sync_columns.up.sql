-- Add columns for schema sync tracking
ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_schema_sync TIMESTAMP;
ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS is_healthy BOOLEAN DEFAULT true;
ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_health_check TIMESTAMP;

-- Add index for health checks
CREATE INDEX IF NOT EXISTS idx_data_sources_is_healthy ON data_sources(is_healthy) WHERE is_healthy = false;
