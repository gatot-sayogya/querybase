# Per-User Role-in-Group Permission Architecture (Plan C)

QueryBase implements a granular, multi-tenant permission model called "Plan C". This model allows administrators to define not only which groups have access to which data sources, but also what specific SQL operations a user can perform based on their assigned role within a specific group.

## Core Concepts

### 1. User Roles in Groups

Every member of a group is assigned a specific role that dictates their baseline permission level for any data source assigned to that group.

| Role        | Purpose                    | Operations Typically Allowed   |
| :---------- | :------------------------- | :----------------------------- |
| **Viewer**  | Read-only observation      | SELECT (limited)               |
| **Member**  | Standard data interaction  | SELECT, INSERT                 |
| **Analyst** | Advanced data manipulation | SELECT, INSERT, UPDATE, DELETE |

### 2. Group-DataSource Permissions

Administrators grant a Group access to a DataSource at a high level:

- `can_read`: Can the group see the schema and run read queries?
- `can_write`: Can the group submit write operations for approval?
- `can_approve`: Can members of this group approve requests? (Usually reserved for Admin/Lead groups).

### 3. Role-Based Policy Overrides

The "Plan C" architecture adds a fine-grained layer where `GroupRolePolicies` define the exact SQL verbs allowed for a specific `role_in_group` on a specific `data_source_id`.

Example:

- Group "Engineering" has access to "Production DB".
- User "Alice" is an "Analyst" in "Engineering".
- The `GroupRolePolicy` for `analyst` on "Production DB" allows `SELECT`, `UPDATE`, but blocks `DELETE`.

## Permission Resolution Logic

When a user attempts to execute a query or submit an approval, the system performs the following resolution:

1. **Identify Parent Groups**: Find all groups the user belongs to that have access to the target DataSource.
2. **Collect Roles**: For each of those groups, identify the user's specific `role_in_group`.
3. **Merge Permissions**:
   - `can_read/write/approve` are **OR-merged** across all eligible groups.
   - SQL Verb visibility (`allow_select`, `allow_insert`, etc.) is resolved by checking the `GroupRolePolicies` for the user's specific roles in those groups.
4. **Effective Permission**: If _any_ of the user's group roles permit the operation on that specific datasource, and the group itself has the high-level `can_read/write` flag, the operation is permitted.

## Database Schema

The implementation relies on two primary tables:

### `user_groups` (Join Table)

- `user_id`: Reference to user
- `group_id`: Reference to group
- `role_in_group`: String (viewer, member, analyst)

### `group_role_policies`

- `group_id`: Reference to group
- `data_source_id`: Reference to data source
- `role_in_group`: String (viewer, member, analyst)
- `allow_select`: Boolean
- `allow_insert`: Boolean
- `allow_update`: Boolean
- `allow_delete`: Boolean

## Advantages of Plan C

- **Flexibility**: A user can be an "Analyst" (high permission) in a "Dev" group but only a "Viewer" (low permission) in a "Production" group.
- **Auditability**: Permissions are tied to functional roles within teams rather than individual overrides.
- **Safety**: Sensitive operations like `DELETE` can be restricted at the role level even if a user has `write` access to the datasource.
