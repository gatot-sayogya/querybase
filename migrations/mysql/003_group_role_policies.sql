-- Add role_in_group to user_groups
ALTER TABLE user_groups
ADD COLUMN role_in_group VARCHAR(20) NOT NULL DEFAULT 'member',
ADD COLUMN joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Create group_role_policies table
CREATE TABLE IF NOT EXISTS group_role_policies (
  id              CHAR(36) PRIMARY KEY,
  group_id        CHAR(36) NOT NULL,
  data_source_id  CHAR(36),
  role_in_group   VARCHAR(20) NOT NULL,
  allow_select    BOOLEAN NOT NULL DEFAULT true,
  allow_insert    BOOLEAN NOT NULL DEFAULT false,
  allow_update    BOOLEAN NOT NULL DEFAULT false,
  allow_delete    BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE (group_id, data_source_id, role_in_group),
  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (data_source_id) REFERENCES data_sources(id) ON DELETE CASCADE
);
