-- Drop group_role_policies table
DROP TABLE IF EXISTS group_role_policies;

-- Remove columns from user_groups
ALTER TABLE user_groups
DROP COLUMN role_in_group,
DROP COLUMN joined_at;
