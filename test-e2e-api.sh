#!/bin/bash
set -e

API_URL="http://localhost:8080/api/v1"
ADMIN_USER="admin@querybase.local"
ADMIN_PASS="admin123"

# E2E Test Data
NEW_USER_EMAIL="e2e_api_test_${RANDOM}@example.com"
NEW_USER_PASS="Pass123!"
GROUP_NAME="E2E API Testing Group ${RANDOM}"

echo "=========================================="
echo " Starting API E2E Workflow Test (Plan C)"
echo "=========================================="

# Helper to extract JSON
get_json_val() {
  echo "$1" | grep -o "\"$2\":\"[^\"]*\"" | cut -d'"' -f4
}

# 1. Login as Admin
echo ""
echo "[1/10] Logging in as Admin..."
ADMIN_LOGIN_RESP=$(curl -s -X POST -H 'Content-Type: application/json' -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" "$API_URL/auth/login")
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN_RESP" | jq -r .token)
if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Admin login failed"
    exit 1
fi
echo "✅ Admin logged in successfully"

# Get a valid datasource ID
# Create a valid local PostgreSQL datasource
echo "Creating a local Test Datasource..."
TEST_DB_PAYLOAD=$(cat <<EOF
{
  "name": "E2E Test DB $RANDOM",
  "type": "postgresql",
  "host": "localhost",
  "port": 5432,
  "database_name": "querybase",
  "username": "querybase",
  "password": "querybase"
}
EOF
)
DS_CREATE_RESP=$(curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' -d "$TEST_DB_PAYLOAD" "$API_URL/datasources")
DS_ID=$(echo "$DS_CREATE_RESP" | jq -r .id)

if [ "$DS_ID" == "null" ] || [ -z "$DS_ID" ]; then
    echo "❌ Failed to create Test Datasource: $DS_CREATE_RESP"
    exit 1
fi
echo "✅ Created Test Datasource ID: $DS_ID"

# 2. Create standard user
echo ""
echo "[2/10] Creating new non-admin user ($NEW_USER_EMAIL)..."
USER_CREATE_RESP=$(curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$NEW_USER_EMAIL\",\"username\":\"e2e_tester_${RANDOM}\",\"full_name\":\"API E2E Tester\",\"password\":\"$NEW_USER_PASS\",\"role\":\"user\"}" \
  "$API_URL/auth/users")
NEW_USER_ID=$(echo "$USER_CREATE_RESP" | jq -r .id)
if [ "$NEW_USER_ID" == "null" ] || [ -z "$NEW_USER_ID" ]; then
    echo "❌ Failed to create user. Response: $USER_CREATE_RESP"
    exit 1
fi
echo "✅ User created with ID: $NEW_USER_ID"

# 3. Create Group
echo ""
echo "[3/10] Creating Group ($GROUP_NAME)..."
GROUP_CREATE_RESP=$(curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"name\":\"$GROUP_NAME\",\"description\":\"Created for API E2E test\"}" \
  "$API_URL/groups")
GROUP_ID=$(echo "$GROUP_CREATE_RESP" | jq -r .id)
if [ "$GROUP_ID" == "null" ] || [ -z "$GROUP_ID" ]; then
    echo "❌ Failed to create group. Response: $GROUP_CREATE_RESP"
    exit 1
fi
echo "✅ Group created with ID: $GROUP_ID"

# 4. Add user to group with "analyst" role
echo ""
echo "[4/10] Adding user to group with 'analyst' role..."
ADD_MEMBER_RESP=$(curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$NEW_USER_ID\",\"role_in_group\":\"analyst\"}" \
  "$API_URL/groups/$GROUP_ID/members")
echo "Response: $ADD_MEMBER_RESP"

# 5. Grant Group access to Datasource
echo ""
echo "[5/11] Granting Group access to Datasource..."
DS_PERM_PAYLOAD="{\"group_id\":\"$GROUP_ID\",\"can_read\":true,\"can_write\":true,\"can_approve\":false}"
DS_PERM_RESP=$(curl -s -X PUT -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "$DS_PERM_PAYLOAD" \
  "$API_URL/datasources/$DS_ID/permissions")
echo "✅ Group access granted."

# 6. Configure Policy for 'analyst' on the datasource
echo ""
echo "[6/11] Configuring group policies for Analyst..."
POLICY_PAYLOAD="{\"role_in_group\":\"analyst\",\"data_source_id\":\"$DS_ID\",\"allow_select\":true,\"allow_insert\":true,\"allow_update\":true,\"allow_delete\":false}"
POLICY_RESP=$(curl -s -X PUT -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "$POLICY_PAYLOAD" \
  "$API_URL/groups/$GROUP_ID/policies")
echo "✅ Policy applied."

# 7. Login as new user
echo ""
echo "[7/11] Logging in as the new user..."
USER_LOGIN_RESP=$(curl -s -X POST -H 'Content-Type: application/json' \
  -d "{\"username\":\"$NEW_USER_EMAIL\",\"password\":\"$NEW_USER_PASS\"}" \
  "$API_URL/auth/login")
USER_TOKEN=$(echo "$USER_LOGIN_RESP" | jq -r .token)
if [ "$USER_TOKEN" == "null" ] || [ -z "$USER_TOKEN" ]; then
    echo "❌ User login failed. Response: $USER_LOGIN_RESP"
    exit 1
fi
echo "✅ User logged in successfully"

# 8. Execute SELECT query
echo ""
echo "[8/11] Testing SELECT query execution..."
SELECT_PAYLOAD="{\"data_source_id\":\"$DS_ID\",\"query_text\":\"SELECT 1 as api_e2e_test;\"}"
SELECT_RESP=$(curl -s -w "\\n%{http_code}" -X POST -H "Authorization: Bearer $USER_TOKEN" -H 'Content-Type: application/json' \
  -d "$SELECT_PAYLOAD" \
  "$API_URL/queries")
HTTP_STATUS=$(echo "$SELECT_RESP" | tail -n 1)
BODY=$(echo "$SELECT_RESP" | sed '$d')
if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ SELECT query succeeded:"
    echo "$BODY" | jq '.data[0]'
else
    echo "❌ SELECT query failed (Status: $HTTP_STATUS): $BODY"
    exit 1
fi

# 9. Submit UPDATE query for approval
echo ""
echo "[9/11] Submitting UPDATE query for approval..."
APPROVAL_PAYLOAD="{\"data_source_id\":\"$DS_ID\",\"query_text\":\"UPDATE users SET full_name = 'E2E Updated' WHERE username = 'admin@querybase.local';\"}"
APPROVAL_RESP=$(curl -s -X POST -H "Authorization: Bearer $USER_TOKEN" -H 'Content-Type: application/json' \
  -d "$APPROVAL_PAYLOAD" \
  "$API_URL/queries")
REQ1_ID=$(echo "$APPROVAL_RESP" | jq -r .approval_id)
if [ "$REQ1_ID" == "null" ] || [ -z "$REQ1_ID" ]; then
    echo "❌ Failed to submit approval (or query failed). Response: $APPROVAL_RESP"
    exit 1
fi
echo "✅ Approval request submitted (ID: $REQ1_ID)"

# 10. Submit another UPDATE query for approval
echo ""
echo "[10/11] Submitting second UPDATE query for approval..."
APPROVAL2_PAYLOAD="{\"data_source_id\":\"$DS_ID\",\"query_text\":\"UPDATE users SET full_name = 'E2E Rejected' WHERE username = 'admin@querybase.local';\"}"
APPROVAL2_RESP=$(curl -s -X POST -H "Authorization: Bearer $USER_TOKEN" -H 'Content-Type: application/json' \
  -d "$APPROVAL2_PAYLOAD" \
  "$API_URL/queries")
REQ2_ID=$(echo "$APPROVAL2_RESP" | jq -r .approval_id)
if [ "$REQ2_ID" == "null" ] || [ -z "$REQ2_ID" ]; then
    echo "❌ Failed to submit second approval: $APPROVAL2_RESP"
    exit 1
fi
echo "✅ Second approval request submitted (ID: $REQ2_ID)"

# 11. Admin Approves and Rejects
echo ""
echo "[11/11] Admin reviewing approvals..."
APPROVE_RESP=$(curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"decision\":\"approved\",\"comments\":\"Looks good\"}" \
  "$API_URL/approvals/$REQ1_ID/review")
echo "✅ First request APPROVED"

REJECT_RESP=$(curl -s -X POST -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"decision\":\"rejected\",\"comments\":\"Testing rejection\"}" \
  "$API_URL/approvals/$REQ2_ID/review")
echo "✅ Second request REJECTED"

echo ""
echo "=========================================="
echo " 🎉 ALL TESTS PASSED SUCCESSFULLY! "
echo "=========================================="

