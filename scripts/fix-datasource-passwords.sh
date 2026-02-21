#!/bin/bash

# Datasource Password Re-encryption Script
# 
# This script re-encrypts all datasource passwords when the JWT_SECRET/ENCRYPTION_KEY
# has been changed. This is necessary because passwords are encrypted with the key,
# and changing the key makes old encrypted passwords unreadable.
#
# Usage:
#   ./scripts/fix-datasource-passwords.sh
#
# Prerequisites:
#   - Backend server must be running on localhost:8080
#   - Admin credentials (admin@querybase.local / admin123)
#   - jq installed (for JSON parsing)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=================================${NC}"
echo -e "${YELLOW}Datasource Password Re-encryption${NC}"
echo -e "${YELLOW}=================================${NC}"
echo ""

# Check if backend is running
echo -n "Checking backend health... "
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}FAILED${NC}"
    echo "Backend is not running on localhost:8080"
    echo "Please start the backend first with: go run cmd/api/main.go"
    exit 1
fi
echo -e "${GREEN}OK${NC}"

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. Output will be less formatted.${NC}"
    echo "Install with: brew install jq (macOS) or apt-get install jq (Linux)"
    USE_JQ=false
else
    USE_JQ=true
fi

# Login and get token
echo -n "Authenticating... "
LOGIN_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin@querybase.local", "password": "admin123"}')

if [ "$USE_JQ" = true ]; then
    TOKEN=$(echo "$LOGIN_RESP" | jq -r '.token')
else
    TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
fi

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}FAILED${NC}"
    echo "Could not authenticate. Please check admin credentials."
    exit 1
fi
echo -e "${GREEN}OK${NC}"

# Get all datasources
echo -n "Fetching datasources... "
DATASOURCES=$(curl -s http://localhost:8080/api/v1/datasources \
  -H "Authorization: Bearer $TOKEN")

if [ "$USE_JQ" = true ]; then
    TOTAL=$(echo "$DATASOURCES" | jq -r '.total')
else
    TOTAL=$(echo "$DATASOURCES" | grep -o '"total":[0-9]*' | cut -d':' -f2)
fi
echo -e "${GREEN}Found $TOTAL datasources${NC}"
echo ""

# Function to update a datasource
update_datasource() {
    local ds_id="$1"
    local password="$2"
    local ds_name="$3"
    
    echo -n "Updating: $ds_name... "
    
    # Update the password
    UPDATE_RESP=$(curl -s -X PUT "http://localhost:8080/api/v1/datasources/$ds_id" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"password\": \"$password\"}")
    
    # Check if update was successful
    if echo "$UPDATE_RESP" | grep -q '"id"'; then
        echo -e "${GREEN}OK${NC}"
        
        # Test the connection
        echo -n "  Testing connection... "
        TEST_RESP=$(curl -s -X POST "http://localhost:8080/api/v1/datasources/$ds_id/test" \
          -H "Authorization: Bearer $TOKEN")
        
        if echo "$TEST_RESP" | grep -q '"success":true'; then
            echo -e "${GREEN}OK${NC}"
            return 0
        else
            echo -e "${RED}FAILED${NC}"
            if [ "$USE_JQ" = true ]; then
                ERROR=$(echo "$TEST_RESP" | jq -r '.error // "Unknown error"')
            else
                ERROR=$(echo "$TEST_RESP" | grep -o '"error":"[^"]*' | cut -d'"' -f4)
            fi
            echo -e "  ${RED}Error: $ERROR${NC}"
            return 1
        fi
    else
        echo -e "${RED}FAILED${NC}"
        if [ "$USE_JQ" = true ]; then
            ERROR=$(echo "$UPDATE_RESP" | jq -r '.error // "Unknown error"')
        else
            ERROR=$(echo "$UPDATE_RESP" | grep -o '"error":"[^"]*' | cut -d'"' -f4)
        fi
        echo -e "  ${RED}Error: $ERROR${NC}"
        return 1
    fi
}

# Process each datasource
UPDATED=0
FAILED=0

# Process each datasource — id | password | name
# Add/remove entries here as datasources change
DATASOURCES="
639e9d34-08b4-4fb7-919f-9f602e13e242|querybase|Test PostgreSQL
04e87632-377e-42b6-9297-56f6c07df354|querybase|Test PostgreSQL (2)
9afe0ac9-9f66-4633-b717-b5eeba23551c|querybase|PostgreSQL Test
1e7d3a97-9a9f-4fe6-813b-0c3aae7aefab|querybase|Updated Test Database
1fd19be8-0050-4ecd-b525-a3d31672e65b|querybase|Schema Test
89c59a2a-09c6-4125-8ca6-7c881caed600|querybase|PostgreSQL Sample Database
2d52fae1-bb32-4257-a0af-6904c1ffad52|querybase|MySQL Target CLI
0224565d-df52-4ecb-9a8e-1ce8e23f6a50|root|MySQL Sample Database
2f454a2e-f83b-42ba-9d4e-efc6f9d043c8|root|Goapotik Dev
"

while IFS="|" read -r ds_id ds_pass ds_name; do
    [ -z "$ds_id" ] && continue
    if update_datasource "$ds_id" "$ds_pass" "$ds_name"; then
        ((UPDATED++))
    else
        ((FAILED++))
    fi
    echo ""
done <<< "$DATASOURCES"

# Summary
echo -e "${YELLOW}=================================${NC}"
echo -e "${YELLOW}Summary${NC}"
echo -e "${YELLOW}=================================${NC}"
echo -e "Total datasources: $TOTAL"
echo -e "Successfully updated: ${GREEN}$UPDATED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All datasources updated successfully!${NC}"
    exit 0
else
    echo -e "${YELLOW}Some datasources failed to update. Please check the errors above.${NC}"
    exit 1
fi
