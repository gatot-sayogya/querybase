# Admin Panel Guide

Learn how to manage users, groups, and data sources in QueryBase.

## Overview

The Admin Panel provides administrative functions for managing:

- **Users**: Create, edit, delete users
- **Groups**: Organize users into teams
- **Data Sources**: Configure database connections
- **Permissions**: Control access to data sources

## Access Requirements

- **Role**: Admin only
- **Navigation**: Dashboard → Admin
- **Permissions**: Full administrative access

---

## User Management

### View Users

1. Navigate to **Admin → Users**
2. See list of all users
3. User information:
   - Email
   - Username
   - Role (Admin, User, Viewer)
   - Created date
   - Last login

### Create User

1. Click **"Add User"** button
2. Fill in user details:
   - **Email**: Required, unique
   - **Username**: Required, unique
   - **Password**: Required, min 8 characters
   - **Role**: Select from dropdown (Admin, User, Viewer)
   - **Active**: Toggle on/off
3. Click **"Create User"**
4. User receives welcome email (optional, to be implemented)

### Edit User

1. Click on user in list
2. Edit user details:
   - **Email**: Can be changed
   - **Username**: Can be changed
   - **Password**: Option to reset
   - **Role**: Update role
   - **Active**: Enable/disable account
3. Click **"Save Changes"**

### Delete User

1. Click on user in list
2. Click **"Delete User"** button
3. Confirm deletion
4. **Warning**: This cannot be undone!

### User Roles

| Role | Permissions |
|------|-------------|
| **Admin** | Full access to all features and admin panel |
| **User** | Can execute queries, submit approvals, manage own data sources |
| **Viewer** | Read-only access, can execute SELECT queries only |

---

## Group Management

### View Groups

1. Navigate to **Admin → Groups**
2. See list of all groups
3. Group information:
   - Name
   - Description
   - Member count
   - Created date

### Create Group

1. Click **"Add Group"** button
2. Fill in group details:
   - **Name**: Required, unique
   - **Description**: Optional
3. Click **"Create Group"**
4. Add users to group (see below)

### Edit Group

1. Click on group in list
2. Edit group details:
   - **Name**: Can be changed
   - **Description**: Update description
3. Click **"Save Changes"**

### Delete Group

1. Click on group in list
2. Click **"Delete Group"** button
3. Confirm deletion
4. **Warning**: This cannot be undone!
5. **Note**: Permissions for this group are also deleted

### Manage Group Members

1. Click on group in list
2. Go to **"Members"** tab
3. **Add Members**:
   - Click **"Add Members"** button
   - Select users from dropdown
   - Click **"Add"**
4. **Remove Members**:
   - Find user in member list
   - Click **"Remove"** next to user
   - Confirm removal

---

## Data Source Management

### View Data Sources

1. Navigate to **Admin → Data Sources**
2. See list of all data sources
3. Data source information:
   - Name
   - Type (PostgreSQL, MySQL)
   - Host
   - Port
   - Database name
   - Status (Connected, Disconnected)
   - Health check time

### Create Data Source

1. Click **"Add Data Source"** button
2. Fill in connection details:
   - **Name**: Required, unique
   - **Type**: PostgreSQL or MySQL
   - **Host**: Server address
   - **Port**: Default 5432 (PostgreSQL) or 3306 (MySQL)
   - **Database Name**: Database to connect to
   - **Username**: Database user
   - **Password**: Encrypted automatically
   - **SSL Mode**: Optional (disable, require, verify-ca, verify-full)
3. Click **"Test Connection"** to verify
4. Click **"Create Data Source"**
5. Set permissions (see below)

### Edit Data Source

1. Click on data source in list
2. Edit connection details
3. Click **"Save Changes"**
4. **Warning**: Connection changes affect all users

### Delete Data Source

1. Click on data source in list
2. Click **"Delete Data Source"** button
3. Confirm deletion
4. **Warning**: This cannot be undone!
5. **Note**: All permissions and query history for this data source are deleted

### Test Connection

1. Click on data source in list
2. Click **"Test Connection"** button
3. View results:
   - ✅ **Connected**: Connection successful
   - ❌ **Failed**: Check error message

### Sync Schema

1. Click on data source in list
2. Click **"Sync Schema"** button
3. Schema refreshes immediately
4. Updated timestamp displays

---

## Permission Management

### Understanding Permissions

Permissions control what users can do with each data source:

| Permission | Description |
|------------|-------------|
| **can_read** | Execute SELECT queries |
| **can_write** | Submit write operation requests (INSERT, UPDATE, DELETE, DDL) |
| **can_approve** | Approve or reject write operation requests |

### View Permissions

1. Click on data source in list
2. Go to **"Permissions"** tab
3. See list of groups with permissions

### Set Permissions

1. Click on data source in list
2. Go to **"Permissions"** tab
3. Click **"Edit Permissions"** button
4. For each group:
   - **can_read**: Check/uncheck
   - **can_write**: Check/uncheck
   - **can_approve**: Check/uncheck
5. Click **"Save Permissions"**

### Permission Best Practices

1. **Principle of Least Privilege**
   - Only grant necessary permissions
   - Start with read-only
   - Add write/approve as needed

2. **Group-Based Permissions**
   - Assign permissions to groups, not individuals
   - Users inherit group permissions
   - Easier to manage

3. **Separation of Duties**
   - `can_write` users should not have `can_approve`
   - Different approvers for different data sources
   - At least 2 approvers per data source

### Example Permission Setups

#### Development Team
- **Developers Group**: can_read, can_write
- **Team Lead Group**: can_read, can_write, can_approve
- **Data Analysts**: can_read only

#### Production Database
- **DBA Group**: can_read, can_write, can_approve
- **Developers**: can_read only (no write access)
- **Analytics**: can_read only

#### Staging Database
- **Developers**: can_read, can_write
- **QA Team**: can_read, can_write, can_approve
- **Stakeholders**: can_read only

---

## Best Practices

### User Management

1. **Use Strong Passwords**
   - Minimum 8 characters
   - Mix of letters, numbers, symbols
   - Require password change on first login

2. **Regular User Audits**
   - Review user list monthly
   - Remove inactive users
   - Update roles as needed

3. **Disable Instead of Delete**
   - Disable accounts temporarily
   - Delete only after 30+ days inactive

### Group Management

1. **Organize by Team/Function**
   - Development team
   - QA team
   - Data analysts
   - DBAs

2. **Clear Group Descriptions**
   - Document group purpose
   - List typical members
   - Note permission level

### Data Source Management

1. **Use Service Accounts**
   - Create dedicated database users for QueryBase
   - Don't use personal database accounts
   - Limit service account permissions

2. **Enable SSL for Production**
   - Use `require` or `verify-ca` SSL mode
   - Protect credentials in transit
   - Encrypt at rest

3. **Regular Health Checks**
   - Monitor connection status
   - Set up alerts for failures
   - Test connections weekly

### Permission Management

1. **Document Permission Matrix**
   - Create spreadsheet of groups → data sources → permissions
   - Review quarterly
   - Update as teams change

2. **Require Multiple Approvers**
   - At least 2 people with `can_approve` per data source
   - Prevents single point of failure
   - Provides oversight

---

## Troubleshooting

### Cannot Create User

**Possible causes:**
- Email/username already exists
- Invalid email format
- Password too weak

**Solutions:**
- Check for duplicates
- Verify email format
- Use stronger password

### Cannot Delete User

**Possible causes:**
- User owns approval requests
- User is only approver for data source

**Solutions:**
- Reassign approval requests
- Add another approver first

### Data Source Connection Fails

**Possible causes:**
- Incorrect host/port
- Firewall blocking connection
- Invalid credentials
- Database not running

**Solutions:**
- Verify connection details
- Check network connectivity
- Test credentials with database client
- Ensure database is running

### Permissions Not Working

**Possible causes:**
- User not in group
- Group permissions not set
- Data source permissions not saved

**Solutions:**
- Verify user group membership
- Check group permissions for data source
- Re-save permissions

---

## Related Features

- **[User Guide](../user-guide/)** - General usage documentation
- **[Approval Workflow](approval-workflow.md)** - How permissions affect approvals
- **[Schema Browser](schema-browser.md)** - Managing data source schemas

---

**Need help?** Check the [main documentation](../README.md) or [contact support](../README.md).
