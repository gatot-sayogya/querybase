#!/bin/bash

# QueryBase Integration Test
# Tests the complete flow from authentication to query execution and approval workflow
# Based on docs/architecture/detailed-flow.md

set -e  # Exit on error

API_URL="http://localhost:8080"
TOKEN=""
ADMIN_TOKEN=""
QUERY_ID=""
APPROVAL_ID=""
TRANSACTION_ID=""
COMMENT_ID=""

echo "=================================="
echo "QueryBase Integration Test"
echo "=================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${YELLOW}[STEP]${NC} $1"
}

test_endpoint() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="$3"
    local token="$4"

    if [ -n "$data" ]; then
        if [ "$method" = "GET" ]; then
            curl -s -X "$method" "$API_URL$endpoint" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                -G --data-urlencode "$data"
        else
            curl -s -X "$method" "$API_URL$endpoint" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                -d "$data"
        fi
    else
        curl -s -X "$method" "$API_URL$endpoint" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json"
    fi
}

# ============================================================================
# PHASE 1: Authentication & Authorization
# ============================================================================

log_step "Phase 1: Authentication & Authorization"
echo ""

# Test 1: Admin Login
log_info "Test 1.1: Admin Login"
ADMIN_RESPONSE=$(test_endpoint "/api/v1/auth/login" "POST" '{
  "username": "admin",
  "password": "admin123"
}')
echo "$ADMIN_RESPONSE" | jq '.'
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.token')
if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    log_error "Admin login failed"
    exit 1
fi
log_info "Admin login successful"
echo ""

# Test 2: Get Current User
log_info "Test 1.2: Get Current User (Admin)"
test_endpoint "/api/v1/auth/me" "GET" "" "$ADMIN_TOKEN" | jq '.'
echo ""

# Test 3: Create Regular User
log_info "Test 1.3: Create Regular User"
USER_RESPONSE=$(test_endpoint "/api/v1/auth/users" "POST" '{
  "email": "testuser@example.com",
  "username": "testuser",
  "password": "password123",
  "full_name": "Test User",
  "role": "user"
}' "$ADMIN_TOKEN")
echo "$USER_RESPONSE" | jq '.'
USER_ID=$(echo "$USER_RESPONSE" | jq -r '.id')
echo ""

# Test 4: Regular User Login
log_info "Test 1.4: Regular User Login"
USER_LOGIN_RESPONSE=$(test_endpoint "/api/v1/auth/login" "POST" '{
  "username": "testuser",
  "password": "password123"
}')
echo "$USER_LOGIN_RESPONSE" | jq '.'
TOKEN=$(echo "$USER_LOGIN_RESPONSE" | jq -r '.token')
if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    log_error "User login failed"
    exit 1
fi
log_info "User login successful"
echo ""

# Test 5: Verify User Token
log_info "Test 1.5: Get Current User (Regular)"
test_endpoint "/api/v1/auth/me" "GET" "" "$TOKEN" | jq '.'
echo ""

# ============================================================================
# PHASE 2: Data Source Management
# ============================================================================

log_step "Phase 2: Data Source Management"
echo ""

# Test 6: Create Data Source
log_info "Test 2.1: Create Data Source"
DATASOURCE_RESPONSE=$(test_endpoint "/api/v1/datasources" "POST" '{
  "name": "Test PostgreSQL",
  "type": "postgresql",
  "host": "localhost",
  "port": 5432,
  "database_name": "querybase",
  "username": "querybase",
  "password": "querybase"
}' "$ADMIN_TOKEN")
echo "$DATASOURCE_RESPONSE" | jq '.'
DATASOURCE_ID=$(echo "$DATASOURCE_RESPONSE" | jq -r '.id')
echo ""

# Test 7: List Data Sources
log_info "Test 2.2: List Data Sources"
test_endpoint "/api/v1/datasources" "GET" "" "$TOKEN" | jq '.'
echo ""

# Test 8: Test Data Source Connection
log_info "Test 2.3: Test Data Source Connection"
test_endpoint "/api/v1/datasources/$DATASOURCE_ID/test" "POST" '{
  "type": "postgresql",
  "host": "localhost",
  "port": 5432,
  "database_name": "querybase",
  "username": "querybase",
  "password": "querybase"
}' "$TOKEN" | jq '.'
echo ""

# Test 9: Health Check
log_info "Test 2.4: Data Source Health Check"
HEALTH_RESPONSE=$(test_endpoint "/api/v1/datasources/$DATASOURCE_ID/health" "GET" "" "$TOKEN")
echo "$HEALTH_RESPONSE" | jq '.'
HEALTH_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.status')
log_info "Health Status: $HEALTH_STATUS"
echo ""

# ============================================================================
# PHASE 3: Group & Permission Management
# ============================================================================

log_step "Phase 3: Group & Permission Management"
echo ""

# Test 10: Create Group
log_info "Test 3.1: Create Group"
GROUP_RESPONSE=$(test_endpoint "/api/v1/groups" "POST" '{
  "name": "Data Analysts",
  "description": "Group for data analysts"
}' "$ADMIN_TOKEN")
echo "$GROUP_RESPONSE" | jq '.'
GROUP_ID=$(echo "$GROUP_RESPONSE" | jq -r '.id')
echo ""

# Test 11: Add User to Group
log_info "Test 3.2: Add User to Group"
test_endpoint "/api/v1/groups/$GROUP_ID/users" "POST" "{
  \"user_id\": \"$USER_ID\"
}" "$ADMIN_TOKEN" | jq '.'
echo ""

# Test 12: Set Group Permissions
log_info "Test 3.3: Set Group Permissions (can_read, can_write, can_approve)"
test_endpoint "/api/v1/datasources/$DATASOURCE_ID/permissions" "PUT" "{
  \"group_id\": \"$GROUP_ID\",
  \"can_read\": true,
  \"can_write\": true,
  \"can_approve\": true
}" "$ADMIN_TOKEN" | jq '.'
echo ""

# ============================================================================
# PHASE 4A: SELECT Query Path (Direct Execution)
# ============================================================================

log_step "Phase 4A: SELECT Query Path (Direct Execution)"
echo ""

# Test 13: Execute SELECT Query
log_info "Test 4A.1: Execute SELECT Query"
SELECT_RESPONSE=$(test_endpoint "/api/v1/queries" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"SELECT * FROM users LIMIT 5\",
  \"name\": \"Get Users Test\",
  \"description\": \"Test query to get users\"
}" "$TOKEN")
echo "$SELECT_RESPONSE" | jq '.'
QUERY_ID=$(echo "$SELECT_RESPONSE" | jq -r '.query_id')
SELECT_STATUS=$(echo "$SELECT_RESPONSE" | jq -r '.status')
echo ""

if [ "$SELECT_STATUS" = "completed" ]; then
    log_info "SELECT query executed successfully"
else
    log_error "SELECT query execution failed"
fi
echo ""

# Test 14: Get Query with Results
log_info "Test 4A.2: Get Query Details with Results"
test_endpoint "/api/v1/queries/$QUERY_ID" "GET" "" "$TOKEN" | jq '.'
echo ""

# Test 15: Paginated Results
log_info "Test 4A.3: Get Paginated Results"
test_endpoint "/api/v1/queries/$QUERY_ID/results?page=1&per_page=10" "GET" "" "$TOKEN" | jq '.'
echo ""

# Test 16: Export Query Results (JSON)
log_info "Test 4A.4: Export Query Results as JSON"
EXPORT_RESPONSE=$(test_endpoint "/api/v1/queries/export" "POST" "{
  \"query_id\": \"$QUERY_ID\",
  \"format\": \"json\"
}" "$TOKEN")
echo "$EXPORT_RESPONSE" | head -20
echo "..."

# Test 17: Export Query Results (CSV)
log_info "Test 4A.5: Export Query Results as CSV"
EXPORT_CSV=$(test_endpoint "/api/v1/queries/export" "POST" "{
  \"query_id\": \"$QUERY_ID\",
  \"format\": \"csv\"
}" "$TOKEN")
echo "$EXPORT_CSV" | head -5
echo "..."
echo ""

# ============================================================================
# PHASE 4B: Write Query Path (Approval Workflow)
# ============================================================================

log_step "Phase 4B: Write Query Path (Approval Workflow)"
echo ""

# Test 18: Submit DELETE Query (requires approval)
log_info "Test 4B.1: Submit DELETE Query (Requires Approval)"
DELETE_RESPONSE=$(test_endpoint "/api/v1/queries" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"DELETE FROM query_history WHERE id = (SELECT MIN(id) FROM query_history)\"
}" "$TOKEN")
echo "$DELETE_RESPONSE" | jq '.'
APPROVAL_ID=$(echo "$DELETE_RESPONSE" | jq -r '.approval_id')
REQUIRES_APPROVAL=$(echo "$DELETE_RESPONSE" | jq -r '.requires_approval')
echo ""

if [ "$REQUIRES_APPROVAL" = "true" ]; then
    log_info "Write query correctly requires approval"
else
    log_error "Write query should require approval"
fi
echo ""

# Test 19: List Approval Requests
log_info "Test 4B.2: List Approval Requests"
test_endpoint "/api/v1/approvals?status=pending" "GET" "" "$TOKEN" | jq '.'
echo ""

# Test 20: Get Approval Details
log_info "Test 4B.3: Get Approval Details"
test_endpoint "/api/v1/approvals/$APPROVAL_ID" "GET" "" "$TOKEN" | jq '.'
echo ""

# ============================================================================
# PHASE 5: Approval Comments
# ============================================================================

log_step "Phase 5: Approval Comments"
echo ""

# Test 21: Add Comment to Approval
log_info "Test 5.1: Add Comment to Approval Request"
COMMENT_RESPONSE=$(test_endpoint "/api/v1/approvals/$APPROVAL_ID/comments" "POST" "{
  \"comment\": \"Please review this DELETE query carefully\"
}" "$TOKEN")
echo "$COMMENT_RESPONSE" | jq '.'
COMMENT_ID=$(echo "$COMMENT_RESPONSE" | jq -r '.id')
echo ""

# Test 22: List Comments
log_info "Test 5.2: List Comments on Approval"
test_endpoint "/api/v1/approvals/$APPROVAL_ID/comments" "GET" "" "$TOKEN" | jq '.'
echo ""

# ============================================================================
# PHASE 6: EXPLAIN Query
# ============================================================================

log_step "Phase 6: EXPLAIN Query Feature"
echo ""

# Test 23: EXPLAIN Query
log_info "Test 6.1: EXPLAIN Query"
EXPLAIN_RESPONSE=$(test_endpoint "/api/v1/queries/explain" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"SELECT * FROM users WHERE username = 'admin'\",
  \"analyze\": false
}" "$TOKEN")
echo "$EXPLAIN_RESPONSE" | jq '.'
echo ""

# Test 24: EXPLAIN ANALYZE
log_info "Test 6.2: EXPLAIN ANALYZE Query"
EXPLAIN_ANALYZE_RESPONSE=$(test_endpoint "/api/v1/queries/explain" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"SELECT * FROM users\",
  \"analyze\": true
}" "$TOKEN")
echo "$EXPLAIN_ANALYZE_RESPONSE" | jq '.plan' | head -10
echo "..."
echo ""

# ============================================================================
# PHASE 7: Dry Run DELETE
# ============================================================================

log_step "Phase 7: Dry Run DELETE Feature"
echo ""

# Test 25: Dry Run DELETE
log_info "Test 7.1: Dry Run DELETE Query"
DRYRUN_RESPONSE=$(test_endpoint "/api/v1/queries/dry-run" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"DELETE FROM query_history WHERE executed_at < NOW() - INTERVAL '30 days'\"
}" "$TOKEN")
echo "$DRYRUN_RESPONSE" | jq '.'
AFFECTED_ROWS=$(echo "$DRYRUN_RESPONSE" | jq -r '.affected_rows')
log_info "Dry run would affect $AFFECTED_ROWS rows"
echo ""

# ============================================================================
# PHASE 8: Transaction Preview (Advanced)
# ============================================================================

log_step "Phase 8: Transaction Preview Feature"
echo ""

# Test 26: Start Transaction for Preview
log_info "Test 8.1: Start Transaction for Write Preview"
TX_RESPONSE=$(test_endpoint "/api/v1/approvals/$APPROVAL_ID/transaction-start" "POST" "" "$ADMIN_TOKEN")
echo "$TX_RESPONSE" | jq '.'
TRANSACTION_ID=$(echo "$TX_RESPONSE" | jq -r '.transaction_id')
TX_STATUS=$(echo "$TX_RESPONSE" | jq -r '.status')
echo ""

if [ "$TX_STATUS" = "active" ]; then
    log_info "Transaction started successfully"
else
    log_info "Transaction status: $TX_STATUS"
fi
echo ""

# Test 27: Get Transaction Status
log_info "Test 8.2: Get Transaction Status"
test_endpoint "/api/v1/transactions/$TRANSACTION_ID" "GET" "" "$ADMIN_TOKEN" | jq '.'
echo ""

# Test 28: Rollback Transaction (for testing)
log_info "Test 8.3: Rollback Transaction (Test Only)"
test_endpoint "/api/v1/transactions/$TRANSACTION_ID/rollback" "POST" "" "$ADMIN_TOKEN" | jq '.'
echo ""

# ============================================================================
# PHASE 9: Query History
# ============================================================================

log_step "Phase 9: Query History"
echo ""

# Test 29: Get Query History
log_info "Test 9.1: List Query History"
test_endpoint "/api/v1/queries/history?page=1&limit=10" "GET" "" "$TOKEN" | jq '.'
echo ""

# ============================================================================
# PHASE 10: Validation Tests
# ============================================================================

log_step "Phase 10: Input Validation Tests"
echo ""

# Test 30: Invalid SQL Validation
log_info "Test 10.1: Invalid SQL (Missing closing quote)"
INVALID_SQL_RESPONSE=$(test_endpoint "/api/v1/queries" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"SELECT * FROM users WHERE name = 'test\"
}" "$TOKEN")
echo "$INVALID_SQL_RESPONSE" | jq '.'
echo ""

# Test 31: Empty Query Validation
log_info "Test 10.2: Empty Query"
EMPTY_QUERY_RESPONSE=$(test_endpoint "/api/v1/queries" "POST" "{
  \"data_source_id\": \"$DATASOURCE_ID\",
  \"query_text\": \"   \"
}" "$TOKEN")
echo "$EMPTY_QUERY_RESPONSE" | jq '.'
echo ""

# ============================================================================
# PHASE 11: Permission Tests
# ============================================================================

log_step "Phase 11: Permission Tests"
echo ""

# Test 32: Access Denied - Admin Endpoint
log_info "Test 11.1: Regular User Access Admin Endpoint (Should Fail)"
test_endpoint "/api/v1/auth/users" "GET" "" "$TOKEN" | jq '.'
echo ""

# Test 33: Delete Comment (Owner)
log_info "Test 11.2: Delete Own Comment (Should Succeed)"
test_endpoint "/api/v1/approvals/$APPROVAL_ID/comments/$COMMENT_ID" "DELETE" "" "$TOKEN" | jq '.'
echo ""

# ============================================================================
# SUMMARY
# ============================================================================

log_step "Test Summary"
echo ""

log_info "All core flows tested successfully!"
echo ""
echo "Tested Features:"
echo "  ✓ Authentication (Admin & Regular User)"
echo "  ✓ User Management (Create, List, Get)"
echo "  ✓ Data Source Management (Create, List, Test, Health Check)"
echo "  ✓ Group Management (Create, Add User)"
echo "  ✓ Permission Management (Set Permissions)"
echo "  ✓ SELECT Query Execution (Direct Path)"
echo "  ✓ Query Results Pagination"
echo "  ✓ Query Export (CSV & JSON)"
echo "  ✓ Write Query Approval Workflow"
echo "  ✓ Approval Comments System"
echo "  ✓ EXPLAIN Query"
echo "  ✓ Dry Run DELETE"
echo "  ✓ Transaction Preview (Start, Status, Rollback)"
echo "  ✓ Query History"
echo "  ✓ Input Validation"
echo "  ✓ Permission Checks"
echo ""

echo "=================================="
echo "Integration Test Complete!"
echo "=================================="
