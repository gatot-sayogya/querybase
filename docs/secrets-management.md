# Secret Key Management

This directory contains utilities for managing QueryBase secret keys.

## Overview

QueryBase uses two critical 32-character keys:

1. **JWT_SECRET** - For signing JWT authentication tokens
2. **ENCRYPTION_KEY** - For encrypting data source passwords

**⚠️ IMPORTANT:** If you lose or change these keys, all encrypted data source passwords will become unreadable!

## Available Scripts

### 1. Generate New Keys

```bash
./scripts/manage-secrets.sh
```

This will:
- Generate new secure 32-character keys
- Validate key length (must be exactly 32 chars for AES-256)
- Save to `.secret_keys` file
- Update your `.env` file
- Backup your old `.env`

### 2. Backup Current Keys

```bash
./scripts/backup-secrets.sh backup
```

Creates a timestamped backup in `.secret_backups/`.
Keeps only the last 10 backups.

### 3. Restore Keys from Backup

```bash
# List available backups
./scripts/backup-secrets.sh list

# Restore a specific backup
./scripts/backup-secrets.sh restore keys_20240115_120000.env
```

### 4. Automatic Backup on Startup

The `start-dev.sh` script now:
1. Validates key length on startup
2. Shows key fingerprints (first/last 8 chars)
3. Fails early if keys are invalid

## Best Practices

### For Development

1. **Use the `.env` file** - It's already configured and loaded by `start-dev.sh`
2. **Don't commit `.env`** - It's in `.gitignore` for security
3. **Backup before changes** - Run `./scripts/backup-secrets.sh` before modifying keys

### For Production

1. **Use environment variables** - Set `JWT_SECRET` and `ENCRYPTION_KEY` in your deployment platform
2. **Use a secrets manager** - Consider AWS Secrets Manager, HashiCorp Vault, or similar
3. **Rotate carefully** - Changing keys requires re-encrypting all data source passwords

### Key Rotation

If you must rotate keys:

1. Backup current keys: `./scripts/backup-secrets.sh`
2. Export all data sources (note their passwords)
3. Generate new keys: `./scripts/manage-secrets.sh`
4. Update all data sources with their passwords
5. Test connections

## Troubleshooting

### "cipher: message authentication failed"

This means the encrypted password can't be decrypted with the current key.

**Solution:**
1. Restore your previous keys from backup
2. Or delete and recreate the data source with the new key

### "JWT_SECRET must be exactly 32 characters"

The key must be exactly 32 bytes for AES-256 encryption.

**Solution:**
Run `./scripts/manage-secrets.sh` to generate a valid key.

### Lost Your Keys?

If you don't have a backup:
1. Data source passwords are unrecoverable
2. You'll need to delete and recreate all data sources
3. Update the connection passwords manually

**Prevention:** Always backup keys before making changes!

## Storage Locations

| Location | Purpose | Git Tracked |
|----------|---------|-------------|
| `.env` | Current active keys | ❌ No |
| `.env.example` | Template with placeholders | ✅ Yes |
| `.secret_keys` | Generated keys backup | ❌ No |
| `.secret_backups/` | Timestamped backups | ❌ No |
| `config/config.yaml` | Default/fallback values | ✅ Yes |

## Security Notes

- Never share your `.env` file
- Never commit real keys to Git
- Restrict file permissions: `chmod 600 .env`
- Use different keys for different environments
- Consider using Docker secrets in production
