# Plan: Remove Group Roles & Policies

## Goal
Simplify the application's permission model by removing the concept of "Role in Group" and deleting the Group Policies tab entirely. Access will be managed strictly via the new Data Sources tab (Read/Write/Approve).

## Proposed Changes

### Database & Models
- `internal/models/group.go`: Remove the `RoleInGroup` property from `UserGroup`.
- `internal/models/group.go`: Delete the entire `GroupRolePolicy` struct.
- Migration: Create or add to auto-migrate a step to drop the `group_role_policies` table and drop the `role_in_group` column from `user_groups` (or just let GORM ignore it, but we should clean it).

### Backend API
- `internal/api/routes/routes.go`: Remove endpoints `GET /:id/policies` and `PUT /:id/policies`.
- `internal/api/handlers/group.go`:
  - Delete `GetRolePolicies` and `SetRolePolicy`.
  - Update `AddUserToGroup` to not require or store `RoleInGroup`.
  - Update `UpdateGroupMemberRole` (probably delete this entire endpoint since members no longer have roles to update).
- `internal/api/dto/group.go`:
  - Remove `GroupRolePolicyRequest` and `GroupRolePolicyResponse`.
  - Update `AddGroupMemberRequest` to remove `RoleInGroup`.
  - Remove `UpdateGroupMemberRoleRequest`.
- `internal/service/query.go`:
  - Update `GetEffectivePermissions` to ONLY rely on `DataSourcePermission` (the checkboxes UI). If `CanRead` is true, grant `CanSelect`. If `CanWrite` is true, grant `CanInsert`, `CanUpdate`, `CanDelete`.

### Frontend
- `web/src/types/index.ts`:
  - Remove `GroupRolePolicy`.
  - Update `GroupMember` to remove `role_in_group`.
- `web/src/lib/api-client.ts`:
  - Remove `getGroupPolicies` and `setGroupPolicy`.
  - Update `assignUserGroups` and `addUserToGroup` if they passed the role.
- `web/src/components/admin/GroupManager.tsx`:
  - Remove the "Policies" tab entirely.
- `web/src/components/admin/GroupPoliciesTab.tsx`:
  - Delete file.
- `web/src/components/admin/GroupMembersTab.tsx`:
  - Remove the dropdown UI that allowed changing a member's role (Admin vs Viewer).

## Verification Plan
1. `go build` succeeds.
2. `npm run lint` succeeds.
3. Groups can still be created and users added.
4. Queries execute correctly based on Data Source permissions, not policies.
