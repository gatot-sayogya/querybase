#!/bin/bash

# Comprehensive Backend API Test Suite
# Tests all major endpoints including the new password reset feature

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}QueryBase Backend API Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Helper function to test endpoint
test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local headers="$4"
    local data="$5"
    local expected_status="$6"
    
    ((TOTAL++))
    echo -n "Testing: $name... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" $headers -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" $headers 2>/dev/null)
    fi
    
    status=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $status)"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_status, got $status)"
        echo "Response: $body"
        ((FAILED++))
        return 1
    fi
}

# Wait for backend to start
echo "Waiting for backend to start..."
for i in {1..10}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}Backend is ready!${NC}"
        echo ""
        break
    fi
    sleep 1
done

# 1. Authentication Tests
echo -e "${YELLOW}=== Authentication Tests ===${NC}"

# Login as admin
LOGIN_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin@querybase.local", "password": "admin123"}')

ADMIN_TOKEN=$(echo $LOGIN_RESP | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$ADMIN_TOKEN" ]; then
    echo -e "${GREEN}✓ Admin login successful${NC}"
    ((PASSED++))
    ((TOTAL++))
else
    echo -e "${RED}✗ Admin login failed${NC}"
    ((FAILED++))
    ((TOTAL++))
fi

# Login as regular user
test_endpoint "User login" "POST" "http://localhost:8080/api/v1/auth/login" \
    "-H 'Content-Type: application/json'" \
    '{"username": "user@querybase.local", "password": "user123"}' \
    "200"

# Invalid credentials
test_endpoint "Invalid login" "POST" "http://localhost:8080/api/v1/auth/login" \
    "-H 'Content-Type: application/json'" \
    '{"username": "admin@querybase.local", "password": "wrongpass"}' \
    "401"

# Get current user
test_endpoint "Get current user" "GET" "http://localhost:8080/api/v1/auth/me" \
    "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
    "" \
    "200"

echo ""

# 2. User Management Tests
echo -e "${YELLOW}=== User Management Tests ===${NC}"

# List users
test_endpoint "List users" "GET" "http://localhost:8080/api/v1/auth/users" \
    "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
    "" \
    "200"

# Get users list for later tests
USERS=$(curl -s http://localhost:8080/api/v1/auth/users \
  -H "Authorization: Bearer $ADMIN_TOKEN")

VIEWER_ID=$(echo $USERS | grep -o '"id":"[a-f0-9-]*","email":"viewer@querybase.local"' | grep -o '"id":"[a-f0-9-]*"' | cut -d'"' -f4)

# Get specific user
if [ -n "$VIEWER_ID" ]; then
    test_endpoint "Get user by ID" "GET" "http://localhost:8080/api/v1/auth/users/$VIEWER_ID" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
fi

# Create user
NEW_USER_EMAIL="testuser_$(date +%s)@example.com"
CREATE_USER_RESP=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$NEW_USER_EMAIL\", \"username\": \"testuser\", \"password\": \"testpass123\", \"full_name\": \"Test User\", \"role\": \"user\"}")

CREATE_STATUS=$(echo "$CREATE_USER_RESP" | tail -n1)
if [ "$CREATE_STATUS" = "201" ]; then
    echo -e "Testing: Create user... ${GREEN}✓ PASS${NC} (HTTP 201)"
    ((PASSED++))
    NEW_USER_ID=$(echo "$CREATE_USER_RESP" | sed '$d' | grep -o '"id":"[^"]*' | cut -d'"' -f4)
else
    echo -e "Testing: Create user... ${RED}✗ FAIL${NC} (Expected 201, got $CREATE_STATUS)"
    ((FAILED++))
fi
((TOTAL++))

echo ""

# 3. Password Reset Tests (NEW)
echo -e "${YELLOW}=== Password Reset Tests ===${NC}"

if [ -n "$VIEWER_ID" ]; then
    # Reset viewer password
    test_endpoint "Admin reset user password" "POST" "http://localhost:8080/api/v1/auth/users/$VIEWER_ID/reset-password" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN' -H 'Content-Type: application/json'" \
        '{"new_password": "newpassword123"}' \
        "200"
    
    # Verify login with new password
    test_endpoint "Login with new password" "POST" "http://localhost:8080/api/v1/auth/login" \
        "-H 'Content-Type: application/json'" \
        '{"username": "viewer@querybase.local", "password": "newpassword123"}' \
        "200"
    
    # Reset back to original
    curl -s -X POST "http://localhost:8080/api/v1/auth/users/$VIEWER_ID/reset-password" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"new_password": "viewer123"}' > /dev/null
    
    # Test password validation (too short)
    test_endpoint "Reset with short password" "POST" "http://localhost:8080/api/v1/auth/users/$VIEWER_ID/reset-password" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN' -H 'Content-Type: application/json'" \
        '{"new_password": "short"}' \
        "400"
fi

# Change password (self-service)
USER_LOGIN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user@querybase.local", "password": "user123"}')
USER_TOKEN=$(echo $USER_LOGIN | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$USER_TOKEN" ]; then
    test_endpoint "Change own password" "POST" "http://localhost:8080/api/v1/auth/change-password" \
        "-H 'Authorization: Bearer $USER_TOKEN' -H 'Content-Type: application/json'" \
        '{"current_password": "user123", "new_password": "newuserpass123"}' \
        "200"
    
    # Reset back
    NEW_USER_LOGIN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
      -H "Content-Type: application/json" \
      -d '{"username": "user@querybase.local", "password": "newuserpass123"}')
    NEW_USER_TOKEN=$(echo $NEW_USER_LOGIN | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    curl -s -X POST http://localhost:8080/api/v1/auth/change-password \
      -H "Authorization: Bearer $NEW_USER_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"current_password": "newuserpass123", "new_password": "user123"}' > /dev/null
fi

echo ""

# 4. Datasource Tests
echo -e "${YELLOW}=== Datasource Tests ===${NC}"

# List datasources
test_endpoint "List datasources" "GET" "http://localhost:8080/api/v1/datasources" \
    "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
    "" \
    "200"

# Get datasources for testing
DATASOURCES=$(curl -s http://localhost:8080/api/v1/datasources \
  -H "Authorization: Bearer $ADMIN_TOKEN")

PG_DS_ID=$(echo $DATASOURCES | grep -o '"id":"[a-f0-9-]*","name":"Test PostgreSQL"' | grep -o '"id":"[a-f0-9-]*"' | cut -d'"' -f4)

if [ -n "$PG_DS_ID" ]; then
    # Get datasource
    test_endpoint "Get datasource" "GET" "http://localhost:8080/api/v1/datasources/$PG_DS_ID" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
    
    # Test connection
    test_endpoint "Test datasource connection" "POST" "http://localhost:8080/api/v1/datasources/$PG_DS_ID/test" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
    
    # Get schema
    test_endpoint "Get datasource schema" "GET" "http://localhost:8080/api/v1/datasources/$PG_DS_ID/schema" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
fi

echo ""

# 5. Query Tests
echo -e "${YELLOW}=== Query Execution Tests ===${NC}"

if [ -n "$PG_DS_ID" ]; then
    # Execute query
    test_endpoint "Execute SELECT query" "POST" "http://localhost:8080/api/v1/queries" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN' -H 'Content-Type: application/json'" \
        "{\"datasource_id\": \"$PG_DS_ID\", \"query\": \"SELECT * FROM users LIMIT 3\"}" \
        "200"
    
    # Query history
    test_endpoint "Get query history" "GET" "http://localhost:8080/api/v1/queries/history" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
fi

echo ""

# 6. Group Tests
echo -e "${YELLOW}=== Group Management Tests ===${NC}"

# List groups
test_endpoint "List groups" "GET" "http://localhost:8080/api/v1/groups" \
    "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
    "" \
    "200"

# Get groups for testing
GROUPS=$(curl -s http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $ADMIN_TOKEN")

GROUP_ID=$(echo $GROUPS | grep -o '"id":"[a-f0-9-]*"' | head -1 | cut -d'"' -f4)

if [ -n "$GROUP_ID" ]; then
    # Get group
    test_endpoint "Get group by ID" "GET" "http://localhost:8080/api/v1/groups/$GROUP_ID" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
    
    # List group users
    test_endpoint "List group users" "GET" "http://localhost:8080/api/v1/groups/$GROUP_ID/users" \
        "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "" \
        "200"
fi

echo ""

# 7. Approval Tests
echo -e "${YELLOW}=== Approval System Tests ===${NC}"

# List approvals
test_endpoint "List approvals" "GET" "http://localhost:8080/api/v1/approvals" \
    "-H 'Authorization: Bearer $ADMIN_TOKEN'" \
    "" \
    "200"

echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed${NC}"
    exit 1
fi
