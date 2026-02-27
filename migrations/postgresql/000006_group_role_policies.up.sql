-- Add role_in_group to user_groups
ALTER TABLE user_groups
ADD COLUMN role_in_group VARCHAR(20) NOT NULL DEFAULT 'member',
ADD COLUMN joined_at TIMESTAMPTZ DEFAULT NOW();

-- Create group_role_policies table
CREATE TABLE group_role_policies (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id        UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  data_source_id  UUID REFERENCES data_sources(id) ON DELETE CASCADE,
  role_in_group   VARCHAR(20) NOT NULL,
  allow_select    BOOLEAN NOT NULL DEFAULT true,
  allow_insert    BOOLEAN NOT NULL DEFAULT false,
  allow_update    BOOLEAN NOT NULL DEFAULT false,
  allow_delete    BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (group_id, data_source_id, role_in_group)
);
