# Local Development Setup

Complete guide to setting up QueryBase for local development.

## Prerequisites

### Required Software

- **Go** 1.21+ - [Download](https://go.dev/dl/)
- **Node.js** 18+ and **npm** - [Download](https://nodejs.org/)
- **Docker** and **Docker Compose** - [Download](https://www.docker.com/products/docker-desktop)
- **Make** - Included on macOS/Linux, Windows via [WSL](https://docs.microsoft.com/en-us/windows/wsl/install)

### Verify Installations

```bash
go version      # Should be go1.21+
node --version  # Should be v18+
npm --version   # Should be 9+
docker --version
docker-compose --version
make --version
```

---

## Project Setup

### 1. Clone Repository

```bash
git clone https://github.com/yourorg/querybase.git
cd querybase
```

### 2. Start Infrastructure

Start PostgreSQL and Redis using Docker Compose:

```bash
make docker-up
```

Verify services are running:

```bash
docker ps
```

You should see:
- `querybase-postgres-1` on port 5432
- `querybase-redis-1` on port 6379

### 3. Run Database Migrations

Create the database schema:

```bash
make migrate-up
```

Verify migrations:

```bash
make db-shell
# In PostgreSQL shell:
\dt
# Should show list of tables
\q
```

### 4. Download Go Dependencies

```bash
make deps
```

### 5. Build Applications

```bash
# Build all (API + Worker)
make build

# Or build separately
make build-api
make build-worker
```

### 6. Install Frontend Dependencies

```bash
cd web
npm install
cd ..
```

---

## Development Servers

You'll need three terminal windows/tabs for development.

### Terminal 1: API Server

```bash
make run-api
```

API runs on http://localhost:8080

### Terminal 2: Background Worker

```bash
make run-worker
```

Worker processes background jobs (query execution, schema sync)

### Terminal 3: Frontend Dev Server

```bash
cd web
npm run dev
```

Frontend runs on http://localhost:3000

---

## Verify Setup

### 1. Check API Health

```bash
curl http://localhost:8080/health
```

Should return:
```json
{"status":"ok"}
```

### 2. Login to Application

1. Open http://localhost:3000
2. Default credentials:
   - **Email:** admin@querybase.local
   - **Password:** admin123
3. You should see the Dashboard

### 3. Create a Test Data Source

1. Go to **Admin → Data Sources**
2. Click **"Add Data Source"**
3. Use the QueryBase PostgreSQL database as a test:
   - **Name:** Test Source
   - **Type:** PostgreSQL
   - **Host:** localhost
   - **Port:** 5432
   - **Database:** querybase
   - **Username:** querybase
   - **Password:** querybase
4. Click **"Test Connection"** → Should succeed
5. Click **"Create Data Source"**

---

## Development Workflow

### Making Backend Changes

1. Edit Go files in `internal/`
2. Restart API server:
   ```bash
   # Terminal 1
   # Press Ctrl+C to stop
   make run-api
   ```
3. Changes are reflected immediately

**Tip:** For faster iteration, use `go run` directly:
```bash
go run cmd/api/main.go
```

### Making Frontend Changes

1. Edit React/Next.js files in `web/src/`
2. Next.js hot-reloads automatically
3. Changes appear in browser within seconds

### Adding New Dependencies

**Backend (Go):**
```bash
go get github.com/package/name
make deps
```

**Frontend (Node):**
```bash
cd web
npm install package-name
```

---

## Database Management

### View Database Contents

```bash
make db-shell
```

Common PostgreSQL commands:
```sql
-- List tables
\dt

-- Describe table
\d users

-- Run query
SELECT * FROM users;

-- Exit
\q
```

### Reset Database

```bash
# Drop all tables
make migrate-down

# Recreate tables
make migrate-up
```

### Seed Data

```bash
# Run seeder (creates admin user)
go run cmd/api/main.go --seed
```

---

## Configuration

### Backend Configuration

Edit `config/config.yaml`:

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
  secret: your-secret-key-min-32-chars
  expire_hours: 24h
```

### Frontend Configuration

Create `web/.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Hot Reload Configuration

### Backend (Air - Optional)

Install [Air](https://github.com/cosmtrek/air) for live reloading:

```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/api ./cmd/api"
  bin = "tmp/api"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor"]
  delay = 1000
```

Run with hot reload:

```bash
air
```

### Frontend (Next.js Built-in)

Next.js includes fast refresh by default. No additional setup needed.

---

## Debugging

### Backend Debugging

**VS Code:**

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch API",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/api",
      "env": {
        "SERVER_MODE": "debug"
      }
    }
  ]
}
```

**Delve (CLI):**

```bash
dlv debug cmd/api/main.go
```

### Frontend Debugging

- **Browser DevTools:** F12 or Cmd+Option+I
- **React DevTools:** [Chrome Extension](https://chrome.google.com/webstore/detail/react-developer-tools/)
- **Next.js Debugging:** Automatically supported in VS Code

---

## Testing

### Backend Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test ./internal/auth/
```

### Frontend Tests

```bash
cd web

# Run unit tests
npm test

# Run with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e
```

---

## Troubleshooting

### Port Already in Use

**Error:** `bind: address already in use`

**Solution:**
```bash
# Find process using port
lsof -i :8080  # API
lsof -i :3000  # Frontend

# Kill process
kill -9 <PID>
```

### Docker Issues

**Error:** `Cannot connect to Docker daemon`

**Solution:**
```bash
# Start Docker Desktop
open -a Docker

# Or restart Docker
docker restart
```

### Database Connection Failed

**Error:** `connection refused`

**Solution:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Restart if needed
docker restart querybase-postgres-1

# Check logs
docker logs querybase-postgres-1
```

### Frontend Build Errors

**Error:** `Module not found`

**Solution:**
```bash
cd web
rm -rf node_modules package-lock.json
npm install
```

### Migration Failures

**Error:** `migration failed`

**Solution:**
```bash
# Check current migration version
make db-status

# Rollback and retry
make migrate-down
make migrate-up
```

---

## Performance Tips

### Backend

- Use `go run` for development instead of building
- Enable debug mode for detailed logging
- Disable rate limiting in development
- Use connection pooling

### Frontend

- Enable Next.js fast refresh (default)
- Use React DevTools Profiler
- Minimize console.logs in production
- Enable source maps in development

---

## Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_PORT` | API server port | 8080 | No |
| `SERVER_MODE` | debug/release | debug | No |
| `DATABASE_HOST` | PostgreSQL host | localhost | Yes |
| `DATABASE_PORT` | PostgreSQL port | 5432 | No |
| `DATABASE_USER` | PostgreSQL user | - | Yes |
| `DATABASE_PASSWORD` | PostgreSQL password | - | Yes |
| `DATABASE_NAME` | Database name | - | Yes |
| `REDIS_HOST` | Redis host | localhost | Yes |
| `REDIS_PORT` | Redis port | 6379 | No |
| `JWT_SECRET` | JWT signing secret | - | Yes |
| `JWT_EXPIRE_HOURS` | Token expiration | 24h | No |
| `CORS_ALLOWED_ORIGINS` | CORS origins | * | No |

---

## Next Steps

- **[Backend Development](backend.md)** - Go project structure
- **[Frontend Development](frontend.md)** - Next.js project structure
- **[Testing Guide](testing.md)** - How to write tests

---

**Need help?** Check the [main documentation](../README.md) or [troubleshooting](#troubleshooting).
