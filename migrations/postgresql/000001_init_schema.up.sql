-- ============================================
-- Database Schema for QueryBase
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- ENUMS
-- ============================================

CREATE TYPE user_role AS ENUM ('admin', 'user', 'viewer');
CREATE TYPE data_source_type AS ENUM ('postgresql', 'mysql');
CREATE TYPE query_status AS ENUM ('pending', 'running', 'completed', 'failed');
CREATE TYPE operation_type AS ENUM ('select', 'insert', 'update', 'delete', 'create_table', 'drop_table', 'alter_table');
CREATE TYPE approval_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed');
CREATE TYPE notification_type AS ENUM ('approval_request', 'approval_status_change', 'query_result', 'error');

-- ============================================
-- USERS & GROUPS
-- ============================================

-- Groups (similar to Redash groups)
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role user_role NOT NULL DEFAULT 'user',
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- User-Group Membership (Many-to-Many)
CREATE TABLE user_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, group_id)
);

-- ============================================
-- DATA SOURCES
-- ============================================

CREATE TABLE data_sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type data_source_type NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    -- Encrypted connection string (encrypted at rest)
    encrypted_password TEXT NOT NULL,
    -- Additional connection parameters (SSL, etc.)
    connection_params JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Group permissions for data sources
CREATE TABLE data_source_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    can_read BOOLEAN DEFAULT TRUE,
    can_write BOOLEAN DEFAULT FALSE,
    can_approve BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(data_source_id, group_id)
);

-- ============================================
-- QUERIES & RESULTS
-- ============================================

CREATE TABLE queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query_text TEXT NOT NULL,
    operation_type operation_type NOT NULL,
    name VARCHAR(500),
    description TEXT,
    status query_status DEFAULT 'pending',
    row_count INTEGER,
    execution_time_ms INTEGER,
    error_message TEXT,
    -- For write operations requiring approval
    requires_approval BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Cached query results (stored as JSONB for flexibility)
CREATE TABLE query_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    data JSONB NOT NULL, -- Results as JSON array of objects
    column_names TEXT[] NOT NULL,
    column_types TEXT[] NOT NULL,
    row_count INTEGER NOT NULL,
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    size_bytes INTEGER
);

-- Query history (for tracking all executions)
CREATE TABLE query_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID REFERENCES queries(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    data_source_id UUID NOT NULL REFERENCES data_sources(id),
    query_text TEXT NOT NULL,
    operation_type operation_type NOT NULL,
    status query_status NOT NULL,
    row_count INTEGER,
    execution_time_ms INTEGER,
    error_message TEXT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- APPROVAL WORKFLOW
-- ============================================

CREATE TABLE approval_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID REFERENCES queries(id) ON DELETE CASCADE,
    -- Alternative: Direct query execution without saved query
    direct_query_id UUID REFERENCES queries(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL REFERENCES users(id),
    operation_type operation_type NOT NULL,
    query_text TEXT NOT NULL,
    data_source_id UUID NOT NULL REFERENCES data_sources(id),
    status approval_status DEFAULT 'pending',
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE approval_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    approval_request_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    reviewed_by UUID NOT NULL REFERENCES users(id),
    status approval_status NOT NULL,
    comments TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(approval_request_id, reviewed_by)
);

-- ============================================
-- NOTIFICATIONS (Google Chat)
-- ============================================

CREATE TABLE notification_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    webhook_url TEXT NOT NULL, -- Google Chat webhook URL
    is_active BOOLEAN DEFAULT TRUE,
    notification_events TEXT[] NOT NULL DEFAULT ARRAY['approval_request', 'approval_status_change', 'query_result'],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, webhook_url)
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_config_id UUID REFERENCES notification_configs(id) ON DELETE SET NULL,
    approval_request_id UUID REFERENCES approval_requests(id) ON DELETE SET NULL,
    query_id UUID REFERENCES queries(id) ON DELETE SET NULL,
    type notification_type NOT NULL,
    status notification_status DEFAULT 'pending',
    payload JSONB NOT NULL,
    retry_count INTEGER DEFAULT 0,
    last_error TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- INDEXES
-- ============================================

-- Users & Groups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_user_groups_user_id ON user_groups(user_id);
CREATE INDEX idx_user_groups_group_id ON user_groups(group_id);

-- Data Sources
CREATE INDEX idx_data_sources_type ON data_sources(type);
CREATE INDEX idx_data_sources_is_active ON data_sources(is_active);
CREATE INDEX idx_data_source_permissions_data_source_id ON data_source_permissions(data_source_id);
CREATE INDEX idx_data_source_permissions_group_id ON data_source_permissions(group_id);

-- Queries
CREATE INDEX idx_queries_data_source_id ON queries(data_source_id);
CREATE INDEX idx_queries_user_id ON queries(user_id);
CREATE INDEX idx_queries_status ON queries(status);
CREATE INDEX idx_queries_created_at ON queries(created_at DESC);
CREATE INDEX idx_queries_requires_approval ON queries(requires_approval);
CREATE INDEX idx_query_results_query_id ON query_results(query_id);
CREATE INDEX idx_query_results_cached_at ON query_results(cached_at);
CREATE INDEX idx_query_history_user_id ON query_history(user_id);
CREATE INDEX idx_query_history_data_source_id ON query_history(data_source_id);
CREATE INDEX idx_query_history_executed_at ON query_history(executed_at DESC);

-- Approvals
CREATE INDEX idx_approval_requests_requested_by ON approval_requests(requested_by);
CREATE INDEX idx_approval_requests_status ON approval_requests(status);
CREATE INDEX idx_approval_requests_created_at ON approval_requests(created_at DESC);
CREATE INDEX idx_approval_reviews_approval_request_id ON approval_reviews(approval_request_id);
CREATE INDEX idx_approval_reviews_reviewed_by ON approval_reviews(reviewed_by);

-- Notifications
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_approval_request_id ON notifications(approval_request_id);

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Update updated_at timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_groups_updated_at BEFORE UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_data_sources_updated_at BEFORE UPDATE ON data_sources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_queries_updated_at BEFORE UPDATE ON queries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_approval_requests_updated_at BEFORE UPDATE ON approval_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_configs_updated_at BEFORE UPDATE ON notification_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- VIEWS
-- ============================================

-- View for user permissions
CREATE VIEW user_permissions AS
SELECT
    u.id AS user_id,
    u.email,
    u.role,
    g.id AS group_id,
    g.name AS group_name,
    dsp.data_source_id,
    ds.name AS data_source_name,
    dsp.can_read,
    dsp.can_write,
    dsp.can_approve
FROM users u
JOIN user_groups ug ON u.id = ug.user_id
JOIN groups g ON ug.group_id = g.id
JOIN data_source_permissions dsp ON g.id = dsp.group_id
JOIN data_sources ds ON dsp.data_source_id = ds.id
WHERE u.deleted_at IS NULL
  AND g.deleted_at IS NULL
  AND ds.deleted_at IS NULL;

-- View for pending approvals
CREATE VIEW pending_approvals AS
SELECT
    ar.id,
    ar.query_id,
    ar.operation_type,
    ar.query_text,
    ds.name AS data_source_name,
    u.email AS requester_email,
    u.full_name AS requester_name,
    ar.created_at,
    ar.status
FROM approval_requests ar
JOIN users u ON ar.requested_by = u.id
JOIN data_sources ds ON ar.data_source_id = ds.id
WHERE ar.status = 'pending'
ORDER BY ar.created_at DESC;

-- ============================================
-- INITIAL DATA
-- ============================================

-- Insert default admin user (password: admin123 - should be changed on first login)
INSERT INTO users (email, username, password_hash, full_name, role)
VALUES (
    'admin@querybase.local',
    'admin',
    '$2a$10$VJCO5iGNpmHnD1g0z/RIIOZkXf9BaJ8FH9jEDlzWdulEC.WsxnGLa',
    'System Administrator',
    'admin'
);

-- Insert default groups
INSERT INTO groups (name, description) VALUES
('Admins', 'Full system access'),
('Data Analysts', 'Read and write access to data sources'),
('Data Viewers', 'Read-only access to data sources');

-- Assign admin to Admins group
INSERT INTO user_groups (user_id, group_id)
SELECT u.id, g.id
FROM users u, groups g
WHERE u.username = 'admin' AND g.name = 'Admins';
