-- ============================================
-- Database Schema for QueryBase (MySQL)
-- ============================================

-- ============================================
-- ENUMS (MySQL uses ENUM type)
-- ============================================

-- ============================================
-- USERS & GROUPS
-- ============================================

-- Groups (similar to Redash groups)
CREATE TABLE groups (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_groups_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Users
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role ENUM('admin', 'user', 'viewer') NOT NULL DEFAULT 'user',
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_users_email (email),
    INDEX idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User-Group Membership (Many-to-Many)
CREATE TABLE user_groups (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    group_id CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_group (user_id, group_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    INDEX idx_user_groups_user_id (user_id),
    INDEX idx_user_groups_group_id (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- DATA SOURCES
-- ============================================

CREATE TABLE data_sources (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type ENUM('postgresql', 'mysql') NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    -- Encrypted connection string (encrypted at rest)
    encrypted_password TEXT NOT NULL,
    -- Additional connection parameters (SSL, etc.)
    connection_params JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_by CHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_data_sources_type (type),
    INDEX idx_data_sources_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Group permissions for data sources
CREATE TABLE data_source_permissions (
    id CHAR(36) PRIMARY KEY,
    data_source_id CHAR(36) NOT NULL,
    group_id CHAR(36) NOT NULL,
    can_read BOOLEAN DEFAULT TRUE,
    can_write BOOLEAN DEFAULT FALSE,
    can_approve BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_datasource_group (data_source_id, group_id),
    FOREIGN KEY (data_source_id) REFERENCES data_sources(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    INDEX idx_data_source_permissions_ds_id (data_source_id),
    INDEX idx_data_source_permissions_group_id (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- QUERIES & RESULTS
-- ============================================

CREATE TABLE queries (
    id CHAR(36) PRIMARY KEY,
    data_source_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    query_text TEXT NOT NULL,
    operation_type ENUM('select', 'insert', 'update', 'delete', 'create_table', 'drop_table', 'alter_table') NOT NULL,
    name VARCHAR(500),
    description TEXT,
    status ENUM('pending', 'running', 'completed', 'failed') DEFAULT 'pending',
    row_count INT,
    execution_time_ms INT,
    error_message TEXT,
    -- For write operations requiring approval
    requires_approval BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (data_source_id) REFERENCES data_sources(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_queries_data_source_id (data_source_id),
    INDEX idx_queries_user_id (user_id),
    INDEX idx_queries_status (status),
    INDEX idx_queries_created_at (created_at),
    INDEX idx_queries_requires_approval (requires_approval)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Cached query results (stored as JSON for flexibility)
CREATE TABLE query_results (
    id CHAR(36) PRIMARY KEY,
    query_id CHAR(36) NOT NULL,
    data JSON NOT NULL,
    column_names JSON NOT NULL,
    column_types JSON NOT NULL,
    row_count INT NOT NULL,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    size_bytes INT,
    FOREIGN KEY (query_id) REFERENCES queries(id) ON DELETE CASCADE,
    INDEX idx_query_results_query_id (query_id),
    INDEX idx_query_results_cached_at (cached_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Query history (for tracking all executions)
CREATE TABLE query_history (
    id CHAR(36) PRIMARY KEY,
    query_id CHAR(36),
    user_id CHAR(36) NOT NULL,
    data_source_id CHAR(36) NOT NULL,
    query_text TEXT NOT NULL,
    operation_type ENUM('select', 'insert', 'update', 'delete', 'create_table', 'drop_table', 'alter_table') NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed') NOT NULL,
    row_count INT,
    execution_time_ms INT,
    error_message TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (query_id) REFERENCES queries(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (data_source_id) REFERENCES data_sources(id),
    INDEX idx_query_history_user_id (user_id),
    INDEX idx_query_history_data_source_id (data_source_id),
    INDEX idx_query_history_executed_at (executed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- APPROVAL WORKFLOW
-- ============================================

CREATE TABLE approval_requests (
    id CHAR(36) PRIMARY KEY,
    query_id CHAR(36),
    direct_query_id CHAR(36),
    requested_by CHAR(36) NOT NULL,
    operation_type ENUM('select', 'insert', 'update', 'delete', 'create_table', 'drop_table', 'alter_table') NOT NULL,
    query_text TEXT NOT NULL,
    data_source_id CHAR(36) NOT NULL,
    status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (query_id) REFERENCES queries(id) ON DELETE CASCADE,
    FOREIGN KEY (direct_query_id) REFERENCES queries(id) ON DELETE CASCADE,
    FOREIGN KEY (requested_by) REFERENCES users(id),
    FOREIGN KEY (data_source_id) REFERENCES data_sources(id),
    INDEX idx_approval_requests_requested_by (requested_by),
    INDEX idx_approval_requests_status (status),
    INDEX idx_approval_requests_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE approval_reviews (
    id CHAR(36) PRIMARY KEY,
    approval_request_id CHAR(36) NOT NULL,
    reviewed_by CHAR(36) NOT NULL,
    status ENUM('pending', 'approved', 'rejected') NOT NULL,
    comments TEXT,
    reviewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_approval_review (approval_request_id, reviewed_by),
    FOREIGN KEY (approval_request_id) REFERENCES approval_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewed_by) REFERENCES users(id),
    INDEX idx_approval_reviews_approval_request_id (approval_request_id),
    INDEX idx_approval_reviews_reviewed_by (reviewed_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- NOTIFICATIONS (Google Chat)
-- ============================================

CREATE TABLE notification_configs (
    id CHAR(36) PRIMARY KEY,
    group_id CHAR(36) NOT NULL,
    webhook_url TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    notification_events JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_group_webhook (group_id, webhook_url(255)),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE notifications (
    id CHAR(36) PRIMARY KEY,
    notification_config_id CHAR(36),
    approval_request_id CHAR(36),
    query_id CHAR(36),
    type ENUM('approval_request', 'approval_status_change', 'query_result', 'error') NOT NULL,
    status ENUM('pending', 'sent', 'failed') DEFAULT 'pending',
    payload JSON NOT NULL,
    retry_count INT DEFAULT 0,
    last_error TEXT,
    sent_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (notification_config_id) REFERENCES notification_configs(id) ON DELETE SET NULL,
    FOREIGN KEY (approval_request_id) REFERENCES approval_requests(id) ON DELETE SET NULL,
    FOREIGN KEY (query_id) REFERENCES queries(id) ON DELETE SET NULL,
    INDEX idx_notifications_status (status),
    INDEX idx_notifications_created_at (created_at),
    INDEX idx_notifications_approval_request_id (approval_request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
INNER JOIN user_groups ug ON u.id = ug.user_id
INNER JOIN groups g ON ug.group_id = g.id
INNER JOIN data_source_permissions dsp ON g.id = dsp.group_id
INNER JOIN data_sources ds ON dsp.data_source_id = ds.id
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
INNER JOIN users u ON ar.requested_by = u.id
INNER JOIN data_sources ds ON ar.data_source_id = ds.id
WHERE ar.status = 'pending'
ORDER BY ar.created_at DESC;

-- ============================================
-- INITIAL DATA
-- ============================================

-- Insert default admin user (password: admin123)
-- Note: This hash is for 'admin123' - change in production
INSERT INTO users (id, email, username, password_hash, full_name, role)
VALUES (
    UUID(),
    'admin@querybase.local',
    'admin',
    '$2a$10$VJCO5iGNpmHnD1g0z/RIIOZkXf9BaJ8FH9jEDlzWdulEC.WsxnGLa',
    'System Administrator',
    'admin'
);

-- Insert default groups
INSERT INTO groups (id, name, description) VALUES
(UUID(), 'Admins', 'Full system access'),
(UUID(), 'Data Analysts', 'Read and write access to data sources'),
(UUID(), 'Data Viewers', 'Read-only access to data sources');

-- Assign admin to Admins group
INSERT INTO user_groups (id, user_id, group_id)
SELECT UUID(), u.id, g.id
FROM users u, groups g
WHERE u.username = 'admin' AND g.name = 'Admins';
