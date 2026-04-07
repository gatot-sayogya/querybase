#!/bin/bash
# QueryBase Secret Key Manager
# This script generates and manages JWT_SECRET and ENCRYPTION_KEY

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$PROJECT_DIR/.env"
KEY_FILE="$PROJECT_DIR/.secret_keys"

# Function to generate a 32-byte key
generate_key() {
    openssl rand -base64 32 | tr -d '=+/' | cut -c1-32
}

# Function to validate key length (must be exactly 32 chars)
validate_key() {
    local key="$1"
    local name="$2"
    if [ ${#key} -ne 32 ]; then
        echo "❌ Error: $name must be exactly 32 characters (current: ${#key})"
        return 1
    fi
    echo "✅ $name is valid (32 characters)"
    return 0
}

# Check if keys exist
if [ -f "$KEY_FILE" ]; then
    echo "📁 Found existing keys in $KEY_FILE"
    source "$KEY_FILE"
    
    # Validate existing keys
    if validate_key "$JWT_SECRET" "JWT_SECRET" && validate_key "$ENCRYPTION_KEY" "ENCRYPTION_KEY"; then
        echo ""
        echo "✅ Existing keys are valid!"
        echo "JWT_SECRET: ${JWT_SECRET:0:8}...${JWT_SECRET: -8}"
        echo "ENCRYPTION_KEY: ${ENCRYPTION_KEY:0:8}...${ENCRYPTION_KEY: -8}"
        exit 0
    else
        echo ""
        echo "⚠️  Existing keys are invalid. Regenerating..."
    fi
fi

# Generate new keys
echo "🔐 Generating new secure keys..."
JWT_SECRET=$(generate_key)
ENCRYPTION_KEY=$(generate_key)

# Validate
echo ""
validate_key "$JWT_SECRET" "JWT_SECRET"
validate_key "$ENCRYPTION_KEY" "ENCRYPTION_KEY"

# Save to key file
echo "JWT_SECRET=$JWT_SECRET" > "$KEY_FILE"
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY" >> "$KEY_FILE"
chmod 600 "$KEY_FILE"

echo ""
echo "✅ Keys saved to $KEY_FILE"
echo ""

# Update .env file if it exists
if [ -f "$ENV_FILE" ]; then
    echo "📝 Updating $ENV_FILE..."
    
    # Backup original
    cp "$ENV_FILE" "$ENV_FILE.backup.$(date +%Y%m%d_%H%M%S)"
    
    # Update JWT_SECRET
    if grep -q "^JWT_SECRET=" "$ENV_FILE"; then
        sed -i.bak "s/^JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$ENV_FILE"
        rm -f "$ENV_FILE.bak"
    else
        echo "JWT_SECRET=$JWT_SECRET" >> "$ENV_FILE"
    fi
    
    # Update ENCRYPTION_KEY
    if grep -q "^ENCRYPTION_KEY=" "$ENV_FILE"; then
        sed -i.bak "s/^ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" "$ENV_FILE"
        rm -f "$ENV_FILE.bak"
    else
        echo "ENCRYPTION_KEY=$ENCRYPTION_KEY" >> "$ENV_FILE"
    fi
    
    echo "✅ $ENV_FILE updated!"
fi

echo ""
echo "🔑 New Keys Generated:"
echo "JWT_SECRET: ${JWT_SECRET:0:8}...${JWT_SECRET: -8}"
echo "ENCRYPTION_KEY: ${ENCRYPTION_KEY:0:8}...${ENCRYPTION_KEY: -8}"
echo ""
echo "⚠️  IMPORTANT: Keep these keys secret and back them up securely!"
echo "   If you lose these keys, all encrypted data source passwords will be unrecoverable."
