# Implementing Data Source Permissions UI

## Goal
The goal was to create a UI so Admins can manage which Data Sources a Group has access to (`can_read`, `can_write`, `can_approve` connection permissions). Previously, "Group Policies" controlled granular SQL statement permissions (SELECT, UPDATE) but there was no way to actually grant connecting access to the Database in the UI.

## Implementation Details

We built the UI into the existing **Group Edit** modal (rather than the Data Source editing flow) because it made more structural sense alongside the existing "Policies" and "Members" tabs.

**Backend Additions:**
- Added `GetGroupDataSourcePermissions(groupId)` to fetch an array of all data sources and a specific group's access.
- Added `SetGroupDataSourcePermission(...)` to toggle `can_read`/`can_write`/`can_approve` per data source.
- Bound these to `GET /api/v1/groups/:id/datasource_permissions` and `PUT /api/v1/groups/:id/datasource_permissions`.

**Frontend Additions:**
- Updated `api-client.ts` to include `getGroupDataSourcePermissions` and `setGroupDataSourcePermission`.
- Reused existing exact table styles from `GroupPoliciesTab` to generate `GroupDataSourcesTab`.
- Placed it between "Members" and "Policies" inside `GroupManager.tsx`.

## Verification Steps
1. The Go backend compiles without errors (`go build -v ./...`).
2. The TypeScript frontend builds cleanly (`npx tsc --noEmit` & `npm run lint`).
3. You can verify this in the UI by opening the admin dashboard, navigating to Groups, clicking "Edit" on a test group, and granting "Read Access" in the new **Data Sources** tab.
