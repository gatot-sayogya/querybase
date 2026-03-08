# Group Data Source Access Implementation Plan

## Overview

Currently, the frontend allows configuring SQL operation policies (`GroupPoliciesTab.tsx`), but disconnected from the actual data source access provisioning (`can_read`, `can_write`, `can_approve` from `data_source_permissions`). As identified, the most intuitive place for admins to manage which databases a group can connect to is inside the **Group Edit Modal**. This plan implements a "Data Sources" tab in the Group manager to map groups to data sources.

## Project Type

WEB (with Backend API additions)

## Success Criteria

- Admins can view which data sources a specific group has access to from the Group Edit modal.
- Admins can toggle `can_read`, `can_write`, and `can_approve` for the group per data source.
- Avoids N+1 query problems by creating a bulk-fetch and bulk-update endpoint for groups.

## Tech Stack

- Go / Gin / GORM (Backend)
- React / Next.js / Tailwind CSS (Frontend)

## File Structure

- `internal/api/handlers/group.go` [MODIFY]
- `internal/api/routes/routes.go` [MODIFY]
- `web/src/lib/api-client.ts` [MODIFY]
- `web/src/components/admin/GroupDataSourcesTab.tsx` [NEW]
- `web/src/components/admin/GroupManager.tsx` [MODIFY]

## Task Breakdown

### Task 1: Backend Group Data Source API

- **Agent**: `backend-specialist`
- **Skill**: `api-patterns`
- **Action**: Add `GetGroupDataSourcePermissions` and `SetGroupDataSourcePermissions` to `group.go`.
  - `GET /api/v1/groups/:id/datasource_permissions`: Returns an array of permissions the group has across all data sources.
  - `PUT /api/v1/groups/:id/datasource_permissions`: Accepts an array of datasource permissions and upserts them for the group (or just reuses the structure for setting one-by-one, but batch is preferred).
- **INPUT**: Group ID
- **OUTPUT**: JSON array of Data Source ID + `can_read/write/approve`.
- **VERIFY**: Direct curl/Postman request to the endpoint returns 200 OK with correct JSON payload.

### Task 2: Frontend API Client

- **Agent**: `frontend-specialist`
- **Skill**: `api-patterns`
- **Action**: Add `getGroupDataSourcePermissions` and `setGroupDataSourcePermission` to `web/src/lib/api-client.ts`. Add corresponding Typescript interfaces mapping to the Go backend response.
- **INPUT**: `groupId` and `dataSourceId` or bulk data payload.
- **OUTPUT**: Methods returning `Promise<GroupDataSourcePermission[]>`.
- **VERIFY**: TypeScript compiles successfully.

### Task 3: Create GroupDataSourcesTab UI

- **Agent**: `frontend-specialist`
- **Skill**: `frontend-design`
- **Action**: Build `GroupDataSourcesTab.tsx`. It fetches all `dataSources` (to show the list) and `getGroupDataSourcePermissions(groupId)`. It displays a table of data sources with checkboxes for "Read Access", "Write Access", and "Approve Access", matching the exact styling of `GroupPoliciesTab.tsx`.
- **INPUT**: `groupId`.
- **OUTPUT**: React component rendered inside the modal.
- **VERIFY**: Toggle state sends accurate PUT requests and updates optimistically.

### Task 4: Integrate Tab into GroupManager

- **Agent**: `frontend-specialist`
- **Skill**: `frontend-design`
- **Action**: In `GroupManager.tsx`, add a 4th tab button: `Data Sources` (between Details/Members and Policies). Render `GroupDataSourcesTab` when active.
- **INPUT**: `activeTab` state.
- **OUTPUT**: Interactive tabbed modal.
- **VERIFY**: Admin clicks "Edit Group", clicks "Data Sources", and can grant access without any console errors.

## Phase X: Verification

- [ ] Run backend tests if available.
- [ ] Spin up the `npm run dev` and `go run cmd/api/main.go` environment.
- [ ] Log in as Admin, create a test group, assign it a data source.
- [ ] Log in as a User in that group, verify the data source appears in the query editor correctly.
- [ ] Run `npm run lint` and `npx tsc --noEmit` to verify type safety.

## 🛑 Socratic Gate (Please Answer Before Continuing)

1. **API Batch Preference**: Do you prefer I build a true "Batch Update" PUT endpoint in Go, or just rapid-fire `PUT` requests to the Go backend when the admin clicks a checkbox (like `GroupPoliciesTab` does)? _(Default: Rapid-fire individual PUT requests to keep it simple and match the Policies tab pattern)._
2. **Approval**: Does this task breakdown correctly reflect your intent to align the UI with the existing Group-centric frontend architecture?

_(Reply to confirm or if you have any changes to the Socratic questions!)_
