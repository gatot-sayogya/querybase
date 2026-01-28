# Integration Testing Guide

This guide explains how to run the comprehensive integration test suite that validates the complete QueryBase flow.

## Prerequisites

1. **Start Services:**
   ```bash
   make docker-up
   ```

2. **Run Migrations:**
   ```bash
   make migrate-up
   ```

3. **Start API Server:**
   ```bash
   make run-api
   ```

   The API will be available at `http://localhost:8080`

## Integration Test Script

The integration test script is located at: `scripts/integration-test.sh`

### What It Tests

The script tests **11 phases** covering **37 test cases**:

#### Phase 1: Authentication & Authorization (5 tests)
- Admin login
- Get current user (admin)
- Create regular user
- Regular user login
- Get current user (regular)

#### Phase 2: Data Source Management (4 tests)
- Create data source
- List data sources
- Test data source connection
- Health check

#### Phase 3: Group & Permission Management (3 tests)
- Create group
- Add user to group
- Set group permissions

#### Phase 4A: SELECT Query Path (4 tests)
- Execute SELECT query
- Get query with results
- Paginated results
- Export query results (JSON & CSV)

#### Phase 4B: Write Query Path (3 tests)
- Submit DELETE query (requires approval)
- List approval requests
- Get approval details

#### Phase 5: Approval Comments (2 tests)
- Add comment to approval
- List comments

#### Phase 6: EXPLAIN Query (2 tests)
- EXPLAIN query
- EXPLAIN ANALYZE

#### Phase 7: Dry Run DELETE (1 test)
- Dry run DELETE query

#### Phase 8: Transaction Preview (3 tests)
- Start transaction
- Get transaction status
- Rollback transaction

#### Phase 9: Query History (1 test)
- List query history

#### Phase 10: Validation Tests (2 tests)
- Invalid SQL validation
- Empty query validation

#### Phase 11: Permission Tests (2 tests)
- Access denied for admin endpoints
- Delete own comment

## Running the Tests

### Option 1: Full Integration Test

```bash
cd /Users/gatotsayogya/Project/querybase
./scripts/integration-test.sh
```

### Option 2: Run with API Server on Different Port

Edit the script to change the `API_URL` variable:

```bash
API_URL="http://localhost:3000" ./scripts/integration-test.sh
```

## Expected Output

The script will output colored results:
- 🟢 `[INFO]` - Successful operations
- 🟡 `[STEP]` - Phase headers
- 🔴 `[ERROR]` - Failed operations

Each test shows JSON response with formatted output.

## Test Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Authentication                                           │
│    ├─ Admin Login ────────┐                                │
│    └─ User Login ─────────┘                                │
└──────────┬──────────────────┬──────────────────────────────────┘
           │                  │
           ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Data Source Management                                  │
│    └─ Create DataSource ──── Health Check                     │
└──────────┬───────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Group & Permissions                                     │
│    └─ Create Group ──── Add User ──── Set Permissions        │
└──────────┬───────────────────────────────────────────────────┘
           │
           ├───────────────────┬─────────────────┐
           ▼                   ▼                 ▼
┌──────────────┐    ┌──────────────┐   ┌──────────────┐
│ 4A. SELECT   │    │ 4B. Write     │   │ 6. EXPLAIN    │
│     Queries  │    │   Queries    │   │   & Dry Run  │
└──────┬───────┘    └──────┬───────┘   └──────────────┘
       │                   │
       ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Approval Comments                                      │
│    └─ Add Comment ──── List Comments                      │
└──────────┬───────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────┐
│ 8. Transaction Preview                                     │
│    └─ Start TX ──── Get Status ──── Rollback              │
└─────────────────────────────────────────────────────────────┘
```

## Troubleshooting

### Test Fails: "Connection Refused"

**Problem:** API server not running

**Solution:**
```bash
make run-api
```

### Test Fails: "Authentication Failed"

**Problem:** Default admin password changed

**Solution:**
```bash
# Check if admin user exists
make db-shell

# In psql:
SELECT email, username FROM users WHERE role = 'admin';

# If password was changed, reset it:
# (In a separate terminal, run:)
# Update via API or directly in database
```

### Test Fails: "Data Source Not Found"

**Problem:** Data source not created or wrong ID

**Solution:**
```bash
# List all data sources
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/datasources

# Use the correct data source ID from the response
```

### Test Fails: "Permission Denied"

**Problem:** User doesn't have required permissions

**Solution:**
```bash
# Check user permissions
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/datasources/$DATASOURCE_ID/permissions

# Ensure group has can_read and can_write permissions
```

## Cleanup After Tests

To clean up test data:

```bash
# Option 1: Run migrations again (fresh start)
make migrate-down
make migrate-up

# Option 2: Manually delete test users
make db-shell

# In psql:
DELETE FROM users WHERE username IN ('testuser', 'testuser2');
DELETE FROM groups WHERE name = 'Data Analysts';
```

## Adding Custom Tests

To add custom tests, edit `scripts/integration-test.sh`:

```bash
# Add your test after the appropriate phase
log_info "Custom Test: My Feature"
MY_TEST=$(test_endpoint "/api/v1/my-endpoint" "POST" '{
  "param": "value"
}' "$TOKEN")
echo "$MY_TEST" | jq '.'
echo ""
```

## Continuous Integration

For CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Integration Tests
  run: |
    make docker-up
    make migrate-up
    make run-api &
    sleep 10
    ./scripts/integration-test.sh
  env:
    API_URL: http://localhost:8080
```

## Related Documentation

- [Detailed Technical Flow](architecture/detailed-flow.md) - Complete flow documentation
- [API Reference](../CLAUDE.md) - All endpoints with examples
- [Testing Guide](development/testing.md) - Unit and integration testing
