# Datasource Password Encryption Troubleshooting

## Problem

After changing the `JWT_SECRET` or `ENCRYPTION_KEY` in `.env`, all datasource connection tests fail with one of these errors:

```
"failed to decrypt password: crypto/aes: invalid key size X"
"failed to decrypt password: cipher: message authentication failed"
```

## Root Cause

The application uses `JWT_SECRET` as the AES encryption key for datasource passwords. When you change this key:

1. **Invalid key size error**: The new key is not 16, 24, or 32 bytes (required for AES-128/192/256)
2. **Message authentication failed**: The key is valid, but passwords were encrypted with the old key and cannot be decrypted with the new one

## Solution

### Quick Fix (Automated)

Run the provided script to re-encrypt all datasource passwords:

```bash
chmod +x scripts/fix-datasource-passwords.sh
./scripts/fix-datasource-passwords.sh
```

This script will:

- Authenticate with the backend
- Fetch all datasources
- Update each datasource password to re-encrypt it
- Test each connection to verify success

### Manual Fix

If you prefer to fix datasources manually or the script doesn't work:

1. **Start the backend**:

   ```bash
   export $(grep -v '^#' .env | xargs)
   export DATABASE_HOST=localhost REDIS_HOST=localhost
   go run cmd/api/main.go
   ```

2. **Get an auth token**:

   ```bash
   LOGIN_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin@querybase.local", "password": "admin123"}')
   TOKEN=$(echo $LOGIN_RESP | grep -o '"token":"[^"]*' | cut -d'"' -f4)
   ```

3. **List all datasources**:

   ```bash
   curl -s http://localhost:8080/api/v1/datasources \
     -H "Authorization: Bearer $TOKEN" | json_pp
   ```

4. **Update each datasource password**:

   ```bash
   curl -X PUT http://localhost:8080/api/v1/datasources/{DATASOURCE_ID} \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"password": "{ACTUAL_PASSWORD}"}'
   ```

5. **Test the connection**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/datasources/{DATASOURCE_ID}/test \
     -H "Authorization: Bearer $TOKEN"
   ```

## Prevention

### 1. Use a Valid Encryption Key

Ensure your `JWT_SECRET` in `.env` is exactly 16, 24, or 32 bytes:

```bash
# Good: 32 bytes (AES-256)
JWT_SECRET=01234567890123456789012345678901

# Bad: 45 bytes
JWT_SECRET=dev-secret-key-at-least-32-chars-long-for-jwt
```

### 2. Separate JWT and Encryption Keys (Recommended)

To avoid this issue, use separate keys for JWT signing and password encryption:

**Update `.env`**:

```bash
JWT_SECRET=your-jwt-signing-key-here
ENCRYPTION_KEY=01234567890123456789012345678901  # Must be 32 bytes
```

**Update `cmd/api/main.go`**:

```go
// Add ENCRYPTION_KEY to config
encryptionKey := os.Getenv("ENCRYPTION_KEY")
if encryptionKey == "" {
    encryptionKey = cfg.JWT.Secret // Fallback
}

// Pass to services
schemaService := service.NewSchemaService(db, encryptionKey)
dataSourceService := service.NewDataSourceService(db, encryptionKey)
```

This way, you can rotate JWT keys without affecting stored passwords.

### 3. Document Your Datasource Passwords

Keep a secure record of all datasource credentials so you can easily re-encrypt them:

```bash
# scripts/fix-datasource-passwords.sh
declare -A DATASOURCE_PASSWORDS=(
    ["639e9d34-08b4-4fb7-919f-9f602e13e242"]="querybase"  # Test PostgreSQL
    ["2d52fae1-bb32-4257-a0af-6904c1ffad52"]="querybase"  # MySQL Target CLI
)
```

## Common Datasources

For reference, these are the default seeded datasources:

| Name             | Type       | Host      | Port | Database  | Username  | Password  |
| ---------------- | ---------- | --------- | ---- | --------- | --------- | --------- |
| Test PostgreSQL  | PostgreSQL | localhost | 5432 | querybase | querybase | querybase |
| MySQL Target CLI | MySQL      | localhost | 3306 | querybase | querybase | querybase |

## Migration Strategy

If you need to rotate the encryption key in production:

1. **Before changing the key**:
   - Export all datasource configurations (passwords will be encrypted with old key)
   - Store the old key securely

2. **After changing the key**:
   - Run the fix script during a maintenance window
   - Verify all connections work
   - Update monitoring/alerting

3. **Alternative (zero-downtime)**:
   - Add code to try decrypting with both old and new keys
   - Re-encrypt with new key on first successful old-key decryption
   - Remove old key support after migration period

## Related Files

- **Script**: [`scripts/fix-datasource-passwords.sh`](file:///Users/gatotsayogya/Project/querybase/scripts/fix-datasource-passwords.sh)
- **Config**: [`.env`](file:///Users/gatotsayogya/Project/querybase/.env)
- **Service**: [`internal/service/schema.go`](file:///Users/gatotsayogya/Project/querybase/internal/service/schema.go)
- **Service**: [`internal/service/datasource.go`](file:///Users/gatotsayogya/Project/querybase/internal/service/datasource.go)
