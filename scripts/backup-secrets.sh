#!/bin/bash
# Backup and restore QueryBase secret keys

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$PROJECT_DIR/.env"
BACKUP_DIR="$PROJECT_DIR/.secret_backups"

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

backup_keys() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/keys_$timestamp.env"
    
    if [ -f "$ENV_FILE" ]; then
        grep -E "^(JWT_SECRET|ENCRYPTION_KEY)=" "$ENV_FILE" > "$backup_file"
        chmod 600 "$backup_file"
        echo "✅ Keys backed up to: $backup_file"
        
        # Keep only last 10 backups
        ls -t "$BACKUP_DIR"/keys_*.env 2>/dev/null | tail -n +11 | xargs rm -f
        echo "   (Kept last 10 backups)"
    else
        echo "❌ No .env file found at $ENV_FILE"
        exit 1
    fi
}

restore_keys() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        echo "Available backups:"
        ls -lt "$BACKUP_DIR"/keys_*.env 2>/dev/null | head -10
        echo ""
        echo "Usage: $0 restore <backup_file>"
        exit 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        # Try to find in backup directory
        backup_file="$BACKUP_DIR/$backup_file"
        if [ ! -f "$backup_file" ]; then
            echo "❌ Backup file not found: $backup_file"
            exit 1
        fi
    fi
    
    echo "Restoring keys from: $backup_file"
    
    # Backup current keys first
    backup_keys
    
    # Extract keys from backup
    local jwt_secret=$(grep "^JWT_SECRET=" "$backup_file" | cut -d'=' -f2)
    local enc_key=$(grep "^ENCRYPTION_KEY=" "$backup_file" | cut -d'=' -f2)
    
    # Update .env file
    if [ -f "$ENV_FILE" ]; then
        # Update or add JWT_SECRET
        if grep -q "^JWT_SECRET=" "$ENV_FILE"; then
            sed -i.bak "s/^JWT_SECRET=.*/JWT_SECRET=$jwt_secret/" "$ENV_FILE"
            rm -f "$ENV_FILE.bak"
        else
            echo "JWT_SECRET=$jwt_secret" >> "$ENV_FILE"
        fi
        
        # Update or add ENCRYPTION_KEY
        if grep -q "^ENCRYPTION_KEY=" "$ENV_FILE"; then
            sed -i.bak "s/^ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$enc_key/" "$ENV_FILE"
            rm -f "$ENV_FILE.bak"
        else
            echo "ENCRYPTION_KEY=$enc_key" >> "$ENV_FILE"
        fi
        
        echo "✅ Keys restored successfully!"
        echo "JWT_SECRET: ${jwt_secret:0:8}...${jwt_secret: -8}"
        echo "ENCRYPTION_KEY: ${enc_key:0:8}...${enc_key: -8}"
        echo ""
        echo "⚠️  Restart your application to use the restored keys"
    fi
}

case "${1:-backup}" in
    backup)
        backup_keys
        ;;
    restore)
        restore_keys "$2"
        ;;
    list)
        echo "Available backups:"
        ls -lt "$BACKUP_DIR"/keys_*.env 2>/dev/null | head -10
        ;;
    *)
        echo "Usage: $0 [backup|restore|list]"
        echo ""
        echo "Commands:"
        echo "  backup        - Backup current keys (default)"
        echo "  restore <file> - Restore keys from backup"
        echo "  list          - List available backups"
        exit 1
        ;;
esac
