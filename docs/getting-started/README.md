# Getting Started with QueryBase

Quick start guide for setting up and running QueryBase.

## Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Docker Setup](#docker-setup)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [Verifying Installation](#verifying-installation)

## Prerequisites

**Required:**
- Go 1.22 or later
- Docker and Docker Compose
- Make (optional, for convenience commands)

**Optional:**
- PostgreSQL client (psql) - for manual database access
- Redis CLI - for queue management
- golangci-lint - for linting

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourorg/querybase.git
cd querybase
```

### 2. Start Docker Services

```bash
make docker-up
```

This starts:
- PostgreSQL 15 on port 5432
- Redis 7 on port 6379

### 3. Run Database Migrations

```bash
make migrate-up
```

This creates all tables and the default admin user.

### 4. Download Dependencies

```bash
make deps
```

### 5. Run the API Server

```bash
make run-api
```

The API will be available at `http://localhost:8080`

## Docker Setup

### Start Services

```bash
make docker-up
```

### Stop Services

```bash
make docker-down
```

### View Logs

```bash
make docker-logs
```

### Services

| Service | Host Port | Description |
|---------|-----------|-------------|
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Job queue |

## Configuration

Configuration is loaded from `config/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug  # debug, release

database:
  host: localhost
  port: 5432
  user: querybase
  password: querybase
  name: querybase

redis:
  host: localhost
  port: 6379

jwt:
  secret: change-this-secret-in-production
  expire_hours: 24
```

**Environment Variables Override:**
You can override any config value using environment variables:
```bash
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export JWT_SECRET=your-secret-here
```

## Running the Application

### API Server

```bash
make run-api
```

Or directly:
```bash
go run ./cmd/api/main.go
```

### Background Worker

```bash
make run-worker
```

Or directly:
```bash
go run ./cmd/worker/main.go
```

### Build Binaries

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# List built binaries
make list
```

## Verifying Installation

### 1. Check Health Endpoint

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### 2. Login as Admin

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

Expected response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "...",
    "username": "admin",
    "email": "admin@querybase.local",
    "role": "admin"
  }
}
```

### 3. Run Tests

```bash
# Run all tests
make test

# Run short tests only (skip database-dependent tests)
make test-short

# Run with coverage
make test-coverage
```

## Default Credentials

**Admin User:**
- Email: `admin@querybase.local`
- Username: `admin`
- Password: `admin123`

⚠️ **IMPORTANT:** Change the admin password after first login!

## Database Access

### Open PostgreSQL Shell

```bash
make db-shell
```

Or directly:
```bash
docker exec -it querybase-postgres-1 psql -U querybase -d querybase
```

### Common Database Commands

```sql
-- List all tables
\dt

-- View users
SELECT id, username, email, role FROM users;

-- View data sources
SELECT id, name, type, host, port FROM data_sources;

-- View recent queries
SELECT * FROM query_history ORDER BY executed_at DESC LIMIT 10;

-- View pending approvals
SELECT * FROM approval_requests WHERE status = 'pending';
```

## Next Steps

1. **Create Data Sources:** Add your PostgreSQL or MySQL databases
2. **Create Groups:** Set up groups for team-based access
3. **Add Users:** Invite team members and assign to groups
4. **Set Permissions:** Configure access control per data source
5. **Run Queries:** Start exploring your data!

## Troubleshooting

### PostgreSQL Connection Issues

**Problem:** Cannot connect to PostgreSQL

**Solution:**
```bash
# Check if PostgreSQL is running
docker-compose -f docker/docker-compose.yml ps postgres

# View logs
make docker-logs | grep postgres

# Restart if needed
docker-compose -f docker/docker-compose.yml restart postgres
```

### Redis Connection Issues

**Problem:** Cannot connect to Redis

**Solution:**
```bash
# Check if Redis is running
docker-compose -f docker/docker-compose.yml ps redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

### Migration Errors

**Problem:** Migration fails

**Solution:**
```bash
# Rollback migrations
make migrate-down

# Run migrations again
make migrate-up

# Check migration status
make migrate-status
```

### Port Already in Use

**Problem:** Port 8080 already in use

**Solution:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change the port in config/config.yaml
```

## Development Setup

For development instructions, see:
- [Development Guide](../development/README.md)
- [Testing Guide](../development/testing.md)
- [Build Guide](../development/build.md)

## Production Deployment

For production deployment considerations, see [CLAUDE.md](../../CLAUDE.md).

**Security Checklist:**
- [ ] Change `jwt.secret` in production
- [ ] Use strong database passwords
- [ ] Enable SSL for database connections (`sslmode=require`)
- [ ] Set `server.mode` to `release`
- [ ] Configure proper CORS settings
- [ ] Use environment variables for sensitive data
- [ ] Set up monitoring and logging
- [ ] Configure backups

## Support

- **Documentation:** [docs/](../)
- **Issues:** [GitHub Issues](https://github.com/yourorg/querybase/issues)
- **Contributing:** See [CLAUDE.md](../../CLAUDE.md)

---

**Last Updated:** January 27, 2025
