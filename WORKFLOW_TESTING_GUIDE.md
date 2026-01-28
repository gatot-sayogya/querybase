# QueryBase Complete Workflow Testing Guide

**Date:** January 28, 2025
**Status:** Ready to Test
**Servers:**
- ✅ Backend API: http://localhost:8080
- ✅ Frontend: http://localhost:3000

---

## Prerequisites Verification

### 1. Backend Health Check

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "message": "QueryBase API is running",
  "status": "ok"
}
```

### 2. Frontend Access

Open browser: http://localhost:3000

**Expected:** Welcome page with "Login to Get Started" button

---

## Database Schema Reference

This section provides an overview of the QueryBase database tables and their columns to help you write effective test queries.

### Quick Schema Overview

```sql
-- List all tables
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

### Core Tables

#### 1. `users` - User Accounts
Stores user authentication and profile information.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) | User email (unique) |
| username | VARCHAR(100) | Username (unique) |
| password_hash | VARCHAR(255) | Bcrypt password hash |
| full_name | VARCHAR(255) | User's full name |
| role | VARCHAR(20) | admin, user, or viewer |
| avatar_url | TEXT | Profile avatar URL |
| is_active | BOOLEAN | Account status |
| created_at | TIMESTAMPTZ | Account creation time |
| updated_at | TIMESTAMPTZ | Last update time |
| deleted_at | TIMESTAMPTZ | Soft delete timestamp |

**Sample Queries:**
```sql
-- List all users with their roles
SELECT id, email, username, full_name, role, is_active
FROM users
ORDER BY created_at DESC;

-- Count users by role
SELECT role, COUNT(*) as user_count
FROM users
GROUP BY role;

-- Find active admins
SELECT username, full_name, email
FROM users
WHERE role = 'admin' AND is_active = true;
```

#### 2. `groups` - User Groups
Groups for organizing users and assigning permissions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| name | VARCHAR(100) | Group name (unique) |
| description | TEXT | Group description |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Last update time |

**Sample Queries:**
```sql
-- List all groups
SELECT id, name, description
FROM groups
ORDER BY name;

-- Count users in each group
SELECT
    g.name,
    COUNT(ug.user_id) as user_count
FROM groups g
LEFT JOIN user_groups ug ON g.id = ug.group_id
GROUP BY g.id, g.name
ORDER BY user_count DESC;
```

#### 3. `user_groups` - Group Memberships
Many-to-many relationship between users and groups.

| Column | Type | Description |
|--------|------|-------------|
| user_id | UUID | Foreign key to users |
| group_id | UUID | Foreign key to groups |
| created_at | TIMESTAMPTZ | When user was added to group |

**Sample Queries:**
```sql
-- Find all groups for a user
SELECT g.name, ugm.created_at as joined_at
FROM user_groups ug
JOIN user_groups ugm ON ugm.user_id = ug.id
JOIN groups g ON ugm.group_id = g.id
WHERE ug.username = 'admin';

-- List all members of a group
SELECT u.username, u.email, u.full_name
FROM user_groups ug
JOIN users u ON ug.user_id = u.id
WHERE ug.group_id = (SELECT id FROM groups WHERE name = 'Admins' LIMIT 1);
```

#### 4. `data_sources` - Database Connections
Configured database connections.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| name | VARCHAR(100) | Data source name |
| type | VARCHAR(20) | postgresql or mysql |
| host | VARCHAR(255) | Database host |
| port | INTEGER | Port number |
| database_name | VARCHAR(100) | Database name |
| username | VARCHAR(100) | Database username |
| password | TEXT | Encrypted password |
| is_active | BOOLEAN | Connection status |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Last update time |

**Sample Queries:**
```sql
-- List all active data sources
SELECT id, name, type, host, port, database_name, is_active
FROM data_sources
WHERE is_active = true
ORDER BY name;

-- Count data sources by type
SELECT type, COUNT(*) as count
FROM data_sources
WHERE is_active = true
GROUP BY type;
```

#### 5. `queries` - Saved and Executed Queries
Query metadata and execution history.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| data_source_id | UUID | Foreign key to data_sources |
| user_id | UUID | Foreign key to users |
| query_text | TEXT | SQL query text |
| operation_type | VARCHAR(20) | select, insert, update, delete, create_table, drop_table, alter_table |
| name | VARCHAR(500) | Optional query name |
| description | TEXT | Optional query description |
| status | VARCHAR(20) | pending, running, completed, failed |
| row_count | INTEGER | Number of rows affected/returned |
| execution_time_ms | INTEGER | Execution time in milliseconds |
| error_message | TEXT | Error message if failed |
| requires_approval | BOOLEAN | Whether approval was needed |
| created_at | TIMESTAMPTZ | Query creation time |
| updated_at | TIMESTAMPTZ | Last update time |
| deleted_at | TIMESTAMPTZ | Soft delete timestamp |

**Sample Queries:**
```sql
-- Recent queries with status
SELECT
    q.id,
    q.name,
    q.operation_type,
    q.status,
    q.row_count,
    q.created_at,
    u.username as created_by,
    ds.name as data_source
FROM queries q
JOIN users u ON q.user_id = u.id
JOIN data_sources ds ON q.data_source_id = ds.id
WHERE q.deleted_at IS NULL
ORDER BY q.created_at DESC
LIMIT 20;

-- Query execution statistics
SELECT
    operation_type,
    status,
    COUNT(*) as query_count,
    AVG(execution_time_ms) as avg_time_ms,
    AVG(row_count) as avg_rows
FROM queries
GROUP BY operation_type, status
ORDER BY operation_type, status;

-- Failed queries
SELECT
    q.query_text,
    q.error_message,
    q.created_at,
    u.username
FROM queries q
JOIN users u ON q.user_id = u.id
WHERE q.status = 'failed'
ORDER BY q.created_at DESC
LIMIT 10;
```

#### 6. `query_results` - Cached Query Results
Cached results from SELECT queries.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| query_id | UUID | Foreign key to queries |
| data | JSONB | Query result data |
| column_names | JSONB | Array of column names |
| column_types | JSONB | Array of column types |
| row_count | INTEGER | Number of rows |
| cached_at | TIMESTAMPTZ | When result was cached |
| expires_at | TIMESTAMPTZ | Cache expiration time |
| size_bytes | INTEGER | Size in bytes |

**Sample Queries:**
```sql
-- Find largest cached results
SELECT
    q.name,
    qr.row_count,
    qr.size_bytes,
    qr.cached_at
FROM query_results qr
JOIN queries q ON qr.query_id = q.id
ORDER BY qr.size_bytes DESC
LIMIT 10;

-- Cache hit rate (queries with cached results vs total)
SELECT
    COUNT(DISTINCT qr.query_id) as cached_queries,
    (SELECT COUNT(*) FROM queries WHERE operation_type = 'select') as total_select_queries,
    ROUND(100.0 * COUNT(DISTINCT qr.query_id) / NULLIF((SELECT COUNT(*) FROM queries WHERE operation_type = 'select'), 0), 2) as cache_hit_percentage
FROM query_results qr;
```

#### 7. `query_history` - Query Execution Log
Complete log of all query executions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| query_id | UUID | Foreign key to queries (nullable) |
| user_id | UUID | Foreign key to users |
| data_source_id | UUID | Foreign key to data_sources |
| query_text | TEXT | SQL query executed |
| operation_type | VARCHAR(20) | Operation type |
| status | VARCHAR(20) | Execution status |
| row_count | INTEGER | Rows affected/returned |
| execution_time_ms | INTEGER | Execution time |
| error_message | TEXT | Error if failed |
| executed_at | TIMESTAMPTZ | Execution timestamp |

**Sample Queries:**
```sql
-- Recent query history
SELECT
    qh.id,
    qh.query_text,
    qh.operation_type,
    qh.status,
    qh.row_count,
    qh.executed_at,
    u.username as executed_by,
    ds.name as data_source
FROM query_history qh
JOIN users u ON qh.user_id = u.id
JOIN data_sources ds ON qh.data_source_id = ds.id
ORDER BY qh.executed_at DESC
LIMIT 50;

-- Query history for specific data source
SELECT
    operation_type,
    status,
    COUNT(*) as execution_count,
    SUM(execution_time_ms) as total_time_ms
FROM query_history
WHERE data_source_id = (SELECT id FROM data_sources WHERE name = 'Updated Test Database' LIMIT 1)
GROUP BY operation_type, status;

-- User activity summary
SELECT
    u.username,
    COUNT(*) as total_queries,
    SUM(CASE WHEN qh.status = 'completed' THEN 1 ELSE 0 END) as successful_queries,
    SUM(CASE WHEN qh.status = 'failed' THEN 1 ELSE 0 END) as failed_queries,
    AVG(qh.execution_time_ms) as avg_time_ms
FROM query_history qh
JOIN users u ON qh.user_id = u.id
GROUP BY u.id, u.username
ORDER BY total_queries DESC;
```

#### 8. `approval_requests` - Approval Queue
Pending and processed approval requests.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| query_id | UUID | Foreign key to queries (nullable) |
| data_source_id | UUID | Foreign key to data_sources |
| query_text | TEXT | SQL query to approve |
| operation_type | VARCHAR(20) | Operation type |
| requested_by | UUID | Foreign key to users |
| status | VARCHAR(20) | pending, approved, rejected |
| rejection_reason | TEXT | Reason for rejection |
| created_at | TIMESTAMPTZ | Request time |
| updated_at | TIMESTAMPTZ | Last update time |

**Sample Queries:**
```sql
-- Pending approval requests
SELECT
    ar.id,
    ar.query_text,
    ar.operation_type,
    ar.created_at,
    u.username as requester,
    u.full_name as requester_name,
    ds.name as data_source
FROM approval_requests ar
JOIN users u ON ar.requested_by = u.id
JOIN data_sources ds ON ar.data_source_id = ds.id
WHERE ar.status = 'pending'
ORDER BY ar.created_at ASC;

-- Approval statistics
SELECT
    status,
    operation_type,
    COUNT(*) as request_count
FROM approval_requests
GROUP BY status, operation_type
ORDER BY status, operation_type;

-- Approval history
SELECT
    ar.operation_type,
    ar.status,
    ar.created_at,
    u_requester.username as requester,
    COUNT(ar.id) OVER () as total_approvals
FROM approval_requests ar
JOIN users u_requester ON ar.requested_by = u_requester.id
ORDER BY ar.created_at DESC
LIMIT 20;
```

#### 9. `approval_reviews` - Approval Decisions
Individual approval reviews and decisions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| approval_request_id | UUID | Foreign key to approval_requests |
| reviewed_by | UUID | Foreign key to users |
| decision | VARCHAR(20) | approved or rejected |
| comments | TEXT | Review comments |
| reviewed_at | TIMESTAMPTZ | Review timestamp |

**Sample Queries:**
```sql
-- Recent approval reviews
SELECT
    arv.decision,
    arv.comments,
    arv.reviewed_at,
    u_reviewer.username as reviewer,
    u_requester.username as requester,
    ar.operation_type
FROM approval_reviews arv
JOIN approval_requests ar ON arv.approval_request_id = ar.id
JOIN users u_reviewer ON arv.reviewed_by = u_reviewer.id
JOIN users u_requester ON ar.requested_by = u_requester.id
ORDER BY arv.reviewed_at DESC
LIMIT 20;

-- Approver activity
SELECT
    u.username,
    COUNT(*) as reviews_count,
    SUM(CASE WHEN arv.decision = 'approved' THEN 1 ELSE 0 END) as approved_count,
    SUM(CASE WHEN arv.decision = 'rejected' THEN 1 ELSE 0 END) as rejected_count
FROM approval_reviews arv
JOIN users u ON arv.reviewed_by = u.id
GROUP BY u.id, u.username
ORDER BY reviews_count DESC;
```

#### 10. `data_source_permissions` - Group Permissions
Permissions for groups to access data sources.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| data_source_id | UUID | Foreign key to data_sources |
| group_id | UUID | Foreign key to groups |
| can_read | BOOLEAN | Can execute SELECT queries |
| can_write | BOOLEAN | Can submit write requests |
| can_approve | BOOLEAN | Can approve write operations |
| created_at | TIMESTAMPTZ | Permission creation time |

**Sample Queries:**
```sql
-- List all permissions
SELECT
    ds.name as data_source,
    g.name as group_name,
    dsp.can_read,
    dsp.can_write,
    dsp.can_approve
FROM data_source_permissions dsp
JOIN data_sources ds ON dsp.data_source_id = ds.id
JOIN groups g ON dsp.group_id = g.id
ORDER BY ds.name, g.name;

-- Groups that can approve for a data source
SELECT
    g.name as group_name,
    COUNT(ug.user_id) as member_count
FROM data_source_permissions dsp
JOIN groups g ON dsp.group_id = g.id
LEFT JOIN user_groups ug ON g.id = ug.group_id
WHERE dsp.data_source_id = (SELECT id FROM data_sources WHERE name = 'Updated Test Database' LIMIT 1)
  AND dsp.can_approve = true
GROUP BY g.id, g.name;
```

---

### Useful Test Queries by Category

#### Simple SELECT Queries
```sql
-- Basic user lookup
SELECT * FROM users LIMIT 10;

-- Count all records
SELECT
    (SELECT COUNT(*) FROM users) as user_count,
    (SELECT COUNT(*) FROM groups) as group_count,
    (SELECT COUNT(*) FROM data_sources) as datasource_count,
    (SELECT COUNT(*) FROM queries) as query_count;

-- Recent activity
SELECT * FROM query_history
ORDER BY executed_at DESC
LIMIT 20;
```

#### JOIN Queries
```sql
-- Users with their groups
SELECT
    u.username,
    u.email,
    COUNT(ug.group_id) as group_count
FROM users u
LEFT JOIN user_groups ug ON u.id = ug.user_id
GROUP BY u.id, u.username, u.email
ORDER BY group_count DESC;

-- Queries with data sources and users
SELECT
    q.operation_type,
    q.status,
    u.username as created_by,
    ds.name as data_source
FROM queries q
JOIN users u ON q.user_id = u.id
JOIN data_sources ds ON q.data_source_id = ds.id
ORDER BY q.created_at DESC
LIMIT 15;
```

#### Aggregation Queries
```sql
-- Query statistics by operation type
SELECT
    operation_type,
    COUNT(*) as total_queries,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
    ROUND(AVG(execution_time_ms)::numeric, 2) as avg_time_ms
FROM queries
GROUP BY operation_type
ORDER BY total_queries DESC;

-- Daily query volume
SELECT
    DATE(executed_at) as execution_date,
    COUNT(*) as query_count,
    COUNT(DISTINCT user_id) as unique_users
FROM query_history
WHERE executed_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(executed_at)
ORDER BY execution_date DESC;
```

#### Performance Analysis Queries
```sql
-- Slowest queries
SELECT
    qh.query_text,
    qh.execution_time_ms,
    qh.row_count,
    qh.executed_at,
    u.username as executed_by
FROM query_history qh
JOIN users u ON qh.user_id = u.id
WHERE qh.execution_time_ms IS NOT NULL
ORDER BY qh.execution_time_ms DESC
LIMIT 10;

-- Largest result sets
SELECT
    q.query_text,
    q.row_count,
    q.execution_time_ms,
    q.created_at,
    ds.name as data_source
FROM queries q
JOIN data_sources ds ON q.data_source_id = ds.id
WHERE q.row_count IS NOT NULL
ORDER BY q.row_count DESC
LIMIT 10;
```

---

## Complete Workflow Test

### Step 1: Login (Authentication)

**URL:** http://localhost:3000/login

**Actions:**
1. Click "Login to Get Started"
2. Enter credentials:
   - Username: `admin`
   - Password: `admin123`
3. Click "Sign in"

**Expected Result:**
- ✅ Redirected to `/dashboard`
- ✅ See Query Editor page
- ✅ Data source selector shows available sources
- ✅ Monaco SQL Editor is visible
- ✅ User info shows "admin (admin)" in navigation

**Verify:**
- [ ] Login successful
- [ ] Redirected to dashboard
- [ ] No error messages
- [ ] JWT token stored in localStorage

---

### Step 2: Execute SELECT Query (Direct Execution)

**URL:** http://localhost:3000/dashboard

**Actions:**
1. Select a data source from dropdown
2. In Monaco Editor, type:
   ```sql
   SELECT * FROM users LIMIT 5;
   ```
3. Click "Run Query" button

**Expected Result:**
- ✅ Loading spinner appears
- ✅ Results table appears below editor
- ✅ Column headers with types (id, email, username, etc.)
- ✅ Data rows displayed (up to 5)
- ✅ Row count shows "X rows returned"
- ✅ Export buttons visible (CSV, JSON)
- ✅ Pagination controls (if more than 50 rows)

**Verify:**
- [ ] Query executes successfully
- [ ] Results displayed correctly
- [ ] Column names and types shown
- [ ] Data is accurate
- [ ] No errors

**Advanced Tests:**
- Try pagination: Click "Next" button
- Try export: Click "Export CSV" (downloads file)
- Try sorting: Click column header (if implemented)
- Try save: Click "Save Query" button

---

### Step 3: Execute Write Query (Approval Workflow)

**URL:** http://localhost:3000/dashboard

**Actions:**
1. Clear editor or type new query:
   ```sql
   DELETE FROM query_history WHERE id = (SELECT MIN(id) FROM query_history WHERE created_at < NOW() - INTERVAL '30 days');
   ```
2. Click "Run Query" button

**Expected Result:**
- ✅ Query submission (may show approval required)
- ✅ Approval request created
- ✅ Message indicates approval required
- ✅ Query appears in approval queue

**Alternative Test (Safer):**
```sql
UPDATE users SET full_name = 'Test User' WHERE username = 'admin';
```

**Verify:**
- [ ] Query submitted successfully
- [ ] Approval workflow triggered
- [ ] No immediate execution (as expected)

---

### Step 4: Navigate to Approvals Dashboard

**URL:** http://localhost:3000/dashboard/approvals

**Navigation:**
1. Click "✅ Approvals" in top navigation
   OR
2. Direct URL: http://localhost:3000/dashboard/approvals

**Expected Result:**
- ✅ Approvals dashboard loads
- ✅ Two-column layout (list + detail)
- ✅ Approval list shows pending requests
- ✅ Filter tabs visible (All, Pending, Approved, Rejected)
- ✅ Pending count badge shows number

**Verify:**
- [ ] Page loads successfully
- [ ] List is visible
- [ ] Pending filter is selected by default
- [ ] Approval requests from Step 3 are visible

---

### Step 5: Review Approval Request

**URL:** http://localhost:3000/dashboard/approvals

**Actions:**
1. Look for the DELETE query approval
2. Click on the approval request in the list

**Expected Result:**
- ✅ Right panel shows approval details
- ✅ SQL query displayed with syntax highlighting
- ✅ Operation type badge (DELETE)
- ✅ Status badge (Pending)
- ✅ Created/Updated timestamps
- ✅ Comment textarea visible
- ✅ Approve button (green) visible
- ✅ Reject button (red) visible

**Verify:**
- [ ] Approval details loaded
- [ ] SQL query is correct
- [ ] All information displayed
- [ ] Action buttons are enabled

---

### Step 6: Add Review Comment

**URL:** http://localhost:3000/dashboard/approvals

**Actions:**
1. In the "Review Comment" textarea, type:
   ```
   This query deletes old query history. Looks safe to proceed.
   ```
2. Leave comment empty for now (or add one)

**Expected Result:**
- ✅ Textarea accepts input
- ✅ Comment is displayed
- ✅ Approve/Reject buttons still enabled

**Verify:**
- [ ] Comment input works
- [ ] Text is visible

---

### Step 7: Reject Approval (Testing)

**URL:** http://localhost:3000/dashboard/approvals

**Actions:**
1. Click the "Reject" button (red)
2. Wait for processing

**Expected Result:**
- ✅ Button shows "Processing..."
- ✅ Success message appears
- ✅ Status updates to "Rejected"
- ✅ Buttons become disabled
- ✅ Info message: "This approval has been rejected. No further actions can be taken."

**Verify:**
- [ ] Rejection successful
- [ ] Status changed from Pending to Rejected
- [ ] Comment saved (if added)
- [ ] Approval list refreshes automatically

---

### Step 8: Test Another Approval (Optional)

**URL:** http://localhost:3000/dashboard

**Actions:**
1. Go back to Query Editor
2. Execute another write query:
   ```sql
   INSERT INTO query_history (user_id, query_text) VALUES (1, 'Test query');
   ```
3. Navigate to Approvals
4. Select the new approval
5. This time, click "Approve"

**Expected Result:**
- ✅ Approval status changes to "Approved"
- ✅ Query executes in background
- ✅ Success message shown

**Verify:**
- [ ] Approval successful
- [ ] Query executed
- [ ] Status updated correctly

---

### Step 9: Query History (Optional)

**URL:** http://localhost:3000/dashboard/history

**Expected Result:**
- ✅ Query history page loads
- ✅ Shows previous queries
- ✅ Can re-run queries
- ✅ Shows execution status

**Note:** This feature may not be fully implemented yet.

---

## Feature Checklist

### Authentication
- [ ] Login with correct credentials
- [ ] Login with incorrect credentials (should fail)
- [ ] Redirect to login if not authenticated
- [ ] Logout functionality
- [ ] JWT token persistence

### Query Editor
- [ ] Monaco Editor loads
- [ ] SQL syntax highlighting works
- [ ] Data source selector populates
- [ ] Run query button works
- [ ] Save query button works
- [ ] Loading states display

### Query Results
- [ ] Results table displays
- [ ] Column headers with types
- [ ] Pagination works (if >50 rows)
- [ ] NULL values formatted correctly
- [ ] Export to CSV works
- [ ] Export to JSON works
- [ ] Empty state shows when no results

### Approval Dashboard
- [ ] Approval list loads
- [ ] Filters work (All/Pending/Approved/Rejected)
- [ ] Status badges display correctly
- [ ] Operation badges display correctly
- [ ] Click to select approval
- [ ] Detail view loads
- [ ] SQL query displays
- [ ] Comment input works
- [ ] Approve button works
- [ ] Reject button works
- [ ] Status updates after action
- [ ] List refreshes automatically

### Navigation
- [ ] Logo links to home
- [ ] Query Editor link active
- [ ] Approvals link active
- [ ] Admin menu visible (for admin users)
- [ ] Mobile menu works (responsive)
- [ ] User info displays correctly
- [ ] Logout works

---

## Error Handling Tests

### Test 1: Invalid SQL

**Query:** `SELECT * FROM nonexistent_table;`

**Expected:**
- ✅ Error message displayed
- ✅ No crash
- ✅ Clear indication of what went wrong

### Test 2: Empty Query

**Action:** Click "Run Query" with empty editor

**Expected:**
- ✅ Validation error
- ✅ "Please enter a SQL query" message

### Test 3: No Data Source

**Action:** Try to run query without selecting data source

**Expected:**
- ✅ "Please select a data source" error

---

## Performance Tests

### Test 1: Large Result Set

**Query:** `SELECT * FROM query_history LIMIT 100;`

**Expected:**
- ✅ Query completes in reasonable time
- ✅ Results load smoothly
- [ ] Pagination works efficiently
- [ ] No browser freeze

### Test 2: Multiple Rapid Queries

**Actions:**
1. Execute SELECT query
2. Immediately execute another SELECT
3. Execute third query

**Expected:**
- ✅ No race conditions
- ✅ Results display correctly for each
- ✅ No memory leaks in browser console

---

## Browser Console Checks

Open browser DevTools (F12) and check:

### Console Tab
- [ ] No JavaScript errors
- [ ] No warnings (except ESLint hook warnings)
- [ ] Network requests shown

### Network Tab
- [ ] API calls visible
- [ ] Request/response formats correct
- [ ] Status codes: 200, 201, etc.

### Application Tab
- [ ] LocalStorage has JWT token
- [ ] Token format: `<JWT_TOKEN_PLACEHOLDER>` (should be a long base64 string)

---

## Success Criteria

### Must Pass
- ✅ User can login
- ✅ User can execute SELECT query
- ✅ Results display correctly
- ✅ User can execute write query
- ✅ Approval request created
- ✅ Approver can view approvals
- ✅ Approver can approve/reject
- ✅ Status updates correctly

### Nice to Have
- [ ] Export functionality works
- [ ] Save query works
- [ ] Pagination smooth
- [ ] Responsive design works on mobile

---

## Troubleshooting

### Issue: "Connection Refused"

**Problem:** Cannot connect to backend

**Solution:**
```bash
# Check if backend is running
curl http://localhost:8080/health

# If not running, start it:
make run-api
```

### Issue: "Login Failed"

**Problem:** Cannot authenticate

**Solution:**
```bash
# Check admin user exists
make db-shell

# In psql:
SELECT email, username FROM users WHERE role = 'admin';
```

### Issue: "No Data Sources"

**Problem:** Data source selector empty

**Solution:**
```bash
# Create a data source first
curl -X POST http://localhost:8080/api/v1/datasources \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test PostgreSQL",
    "type": "postgresql",
    "host": "localhost",
    "port": 5432,
    "database_name": "querybase",
    "username": "querybase",
    "password": "querybase"
  }'
```

### Issue: "Approvals Not Showing"

**Problem:** Approval list empty

**Solution:**
- Make sure you've executed a write query first
- Check that query required approval (INSERT/UPDATE/DELETE)
- Try switching filter to "All"

---

## Test Results Summary

After completing all tests, fill out:

| Test | Status | Notes |
|------|--------|-------|
| Login | ⬜ Pass / Fail | |
| SELECT Query | ⬜ Pass / Fail | |
| Results Display | ⬜ Pass / Fail | |
| Pagination | ⬜ Pass / Fail | |
| Export CSV | ⬜ Pass / Fail | |
| Export JSON | ⬜ Pass / Fail | |
| Write Query Submission | ⬜ Pass / Fail | |
| Approval List | ⬜ Pass / Fail | |
| Approval Detail | ⬜ Pass / Fail | |
| Approve Action | ⬜ Pass / Fail | |
| Reject Action | ⬜ Pass / Fail | |
| Status Update | ⬜ Pass / Fail | |

**Overall Result:** _____ / 11 tests passed

---

## Next Steps After Testing

### If All Tests Pass:
- ✅ Commit Phase 3 to GitHub
- ✅ Update documentation
- ✅ Consider Phase 4 (Admin Features)

### If Tests Fail:
1. Note which tests failed
2. Check browser console for errors
3. Check backend logs
4. Verify API endpoints are working
5. Report issues for fixing

---

## API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/auth/login` | POST | Login |
| `/api/v1/auth/me` | GET | Get current user |
| `/api/v1/datasources` | GET | List data sources |
| `/api/v1/queries` | POST | Execute query |
| `/api/v1/queries/:id` | GET | Get query with results |
| `/api/v1/approvals` | GET | List approvals |
| `/api/v1/approvals/:id` | GET | Get approval details |
| `/api/v1/approvals/:id/review` | POST | Approve/reject |

---

## Ready-to-Use Test Queries

Copy and paste these queries directly into the Query Editor for testing:

### Beginner Queries (Safe to Run)

#### 1. Count Records
```sql
-- Count records in various tables
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'groups', COUNT(*) FROM groups
UNION ALL
SELECT 'data_sources', COUNT(*) FROM data_sources
UNION ALL
SELECT 'queries', COUNT(*) FROM queries;
```

#### 2. List Active Users
```sql
SELECT
    id,
    username,
    email,
    full_name,
    role,
    is_active,
    created_at
FROM users
WHERE is_active = true
ORDER BY created_at DESC;
```

#### 3. Recent Query History
```sql
SELECT
    id,
    query_text,
    operation_type,
    status,
    executed_at
FROM query_history
ORDER BY executed_at DESC
LIMIT 10;
```

#### 4. Data Sources Overview
```sql
SELECT
    name,
    type,
    host,
    port,
    database_name,
    is_active
FROM data_sources
ORDER BY name;
```

### Intermediate Queries

#### 5. User Group Membership
```sql
SELECT
    u.username,
    u.email,
    STRING_AGG(g.name, ', ') as groups
FROM users u
LEFT JOIN user_groups ug ON u.id = ug.user_id
LEFT JOIN groups g ON ug.group_id = g.id
GROUP BY u.id, u.username, u.email
ORDER BY u.username;
```

#### 6. Query Performance Summary
```sql
SELECT
    operation_type,
    COUNT(*) as total_queries,
    ROUND(AVG(execution_time_ms)::numeric, 2) as avg_time_ms,
    MIN(execution_time_ms) as min_time_ms,
    MAX(execution_time_ms) as max_time_ms
FROM query_history
WHERE execution_time_ms IS NOT NULL
GROUP BY operation_type
ORDER BY total_queries DESC;
```

#### 7. Approval Status Summary
```sql
SELECT
    status,
    operation_type,
    COUNT(*) as count
FROM approval_requests
GROUP BY status, operation_type
ORDER BY status, operation_type;
```

#### 8. Recent Failed Queries
```sql
SELECT
    qh.query_text,
    qh.error_message,
    qh.executed_at,
    u.username as executed_by
FROM query_history qh
JOIN users u ON qh.user_id = u.id
WHERE qh.status = 'failed'
ORDER BY qh.executed_at DESC
LIMIT 10;
```

### Advanced Queries

#### 9. Complete User Activity Report
```sql
SELECT
    u.username,
    u.full_name,
    u.email,
    COUNT(DISTINCT qh.id) as total_queries,
    COUNT(DISTINCT CASE WHEN qh.status = 'completed' THEN qh.id END) as successful,
    COUNT(DISTINCT CASE WHEN qh.status = 'failed' THEN qh.id END) as failed,
    ROUND(AVG(qh.execution_time_ms)::numeric, 2) as avg_time_ms,
    MIN(qh.executed_at) as first_query,
    MAX(qh.executed_at) as last_query
FROM users u
LEFT JOIN query_history qh ON u.id = qh.user_id
GROUP BY u.id, u.username, u.full_name, u.email
ORDER BY total_queries DESC;
```

#### 10. Data Source Usage Statistics
```sql
SELECT
    ds.name as data_source,
    ds.type,
    COUNT(DISTINCT qh.id) as query_count,
    COUNT(DISTINCT qh.user_id) as unique_users,
    ROUND(AVG(qh.execution_time_ms)::numeric, 2) as avg_time_ms,
    SUM(CASE WHEN qh.status = 'completed' THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN qh.status = 'failed' THEN 1 ELSE 0 END) as failed
FROM data_sources ds
LEFT JOIN query_history qh ON ds.id = qh.data_source_id
WHERE ds.is_active = true
GROUP BY ds.id, ds.name, ds.type
ORDER BY query_count DESC;
```

#### 11. Pending Approvals Detail
```sql
SELECT
    ar.id as approval_id,
    ar.query_text,
    ar.operation_type,
    ar.created_at as requested_at,
    u.username as requester,
    u.full_name as requester_name,
    u.email as requester_email,
    ds.name as data_source
FROM approval_requests ar
JOIN users u ON ar.requested_by = u.id
JOIN data_sources ds ON ar.data_source_id = ds.id
WHERE ar.status = 'pending'
ORDER BY ar.created_at ASC;
```

#### 12. Query Execution Trends (Last 7 Days)
```sql
SELECT
    DATE(executed_at) as date,
    COUNT(*) as total_queries,
    COUNT(DISTINCT user_id) as unique_users,
    SUM(CASE WHEN operation_type = 'select' THEN 1 ELSE 0 END) as select_queries,
    SUM(CASE WHEN operation_type IN ('insert', 'update', 'delete') THEN 1 ELSE 0 END) as write_queries,
    ROUND(AVG(execution_time_ms)::numeric, 2) as avg_time_ms
FROM query_history
WHERE executed_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(executed_at)
ORDER BY date DESC;
```

### Data Quality Queries

#### 13. Find Orphaned Records
```sql
-- Query history without valid user
SELECT
    qh.id,
    qh.query_text,
    qh.user_id
FROM query_history qh
LEFT JOIN users u ON qh.user_id = u.id
WHERE u.id IS NULL;

-- Query history without valid data source
SELECT
    qh.id,
    qh.query_text,
    qh.data_source_id
FROM query_history qh
LEFT JOIN data_sources ds ON qh.data_source_id = ds.id
WHERE ds.id IS NULL;
```

#### 14. Duplicate Detection
```sql
-- Find potential duplicate usernames (if any)
SELECT
    username,
    COUNT(*) as count
FROM users
GROUP BY username
HAVING COUNT(*) > 1;

-- Find queries with same text
SELECT
    query_text,
    COUNT(*) as execution_count
FROM query_history
GROUP BY query_text
HAVING COUNT(*) > 1
ORDER BY execution_count DESC
LIMIT 10;
```

---

## Query Templates by Use Case

### For Testing Data Source Connections
```sql
-- Simple connection test
SELECT 1 as test_value, NOW() as current_time;

-- Table existence check
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
LIMIT 10;
```

### For Performance Testing
```sql
-- Small result set (fast)
SELECT * FROM users LIMIT 10;

-- Medium result set
SELECT * FROM query_history LIMIT 100;

-- Join test
SELECT
    u.username,
    qh.query_text,
    qh.executed_at
FROM users u
JOIN query_history qh ON u.id = qh.user_id
LIMIT 50;
```

### For Testing Approval Workflow
```sql
-- Safe UPDATE query (requires approval)
UPDATE users
SET full_name = 'Test Update'
WHERE username = 'admin';

-- Safe INSERT query (requires approval)
-- NOTE: This may fail due to constraints, but will trigger approval
-- INSERT INTO query_history (user_id, query_text, operation_type, status)
-- VALUES (1, 'Test query', 'select', 'completed');
```

### For Testing Export Functionality
```sql
-- Query with various data types for export testing
SELECT
    u.id,
    u.username,
    u.email,
    u.full_name,
    u.role,
    u.is_active,
    u.created_at,
    u.updated_at
FROM users u
WHERE u.is_active = true
LIMIT 50;
```

---

**Testing Date:** _____________
**Tester:** _____________
**Backend Version:** 0.1.0
**Frontend Version:** 0.1.0
**Environment:** Development
