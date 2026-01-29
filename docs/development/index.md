# Development Guide

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker and Docker Compose
- Make (optional but recommended)

## Initial Setup

### 1. Clone and Setup

```bash
# Clone repository
git clone <repository-url>
cd querybase

# Start infrastructure (PostgreSQL, Redis)
make docker-up

# Run database migrations
make migrate-up

# Download dependencies
make deps
```

### 2. Development Servers

Open 3 terminal windows:

**Terminal 1 - API Server:**
```bash
make run-api
# Server runs on http://localhost:8080
```

**Terminal 2 - Background Worker:**
```bash
make run-worker
# Processes background jobs
```

**Terminal 3 - Frontend:**
```bash
cd web
npm install
npm run dev
# Next.js runs on http://localhost:3000
```

## Development Workflow

### Code Organization

**Backend (Go):**
```
internal/
├── api/          # HTTP layer (handlers, middleware, routes)
├── auth/         # Authentication (JWT, password)
├── config/       # Configuration management
├── database/     # Database connections
├── models/       # Data models
├── service/      # Business logic
├── queue/        # Background jobs
└── validation/   # Input validation
```

**Frontend (Next.js):**
```
web/src/
├── app/          # Pages (App Router)
├── components/   # React components
├── stores/       # State management (Zustand)
├── lib/          # Utilities (API client, helpers)
└── types/        # TypeScript types
```

### Adding New API Endpoint

1. **Define DTO** (`internal/api/dto/`):
```go
type MyRequestDTO struct {
    Field string `json:"field" binding:"required"`
}
```

2. **Add Handler** (`internal/api/handlers/`):
```go
func (h *MyHandler) HandleMyEndpoint(c *gin.Context) {
    var req MyRequestDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // ... handler logic
}
```

3. **Register Route** (`internal/api/routes/routes.go`):
```go
api.GET("/my-endpoint", myHandler.HandleMyEndpoint)
```

### Adding New Frontend Page

1. **Create Page** (`web/src/app/my-page/page.tsx`):
```typescript
export default function MyPage() {
  return <div>My Page Content</div>
}
```

2. **Add Navigation** (update layout components if needed)

3. **Fetch Data** (use stores or API client)

## Testing

### Backend Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test ./internal/service -v
```

### Frontend Tests

```bash
cd web

# Unit tests
npm test

# E2E tests
npm run test:e2e
```

## Hot Reload

- **Backend:** No built-in hot reload. Restart after code changes:
  ```bash
  # Press Ctrl+C to stop, then:
  make run-api
  ```

- **Frontend:** Next.js has hot reload enabled by default
  ```bash
  npm run dev  # Auto-reloads on save
  ```

## Database Migrations

### Creating New Migration

```bash
# Generate migration name
make migrate-create name=add_new_feature

# This creates empty files in migrations/postgresql/
```

### Writing Migration

**Up migration** (`migrations/postgresql/000006_add_new_feature.up.sql`):
```sql
ALTER TABLE users ADD COLUMN new_field VARCHAR(255);
CREATE INDEX idx_users_new_field ON users(new_field);
```

**Down migration** (`migrations/postgresql/000006_add_new_feature.down.sql`):
```sql
DROP INDEX IF EXISTS idx_users_new_field;
ALTER TABLE users DROP COLUMN IF EXISTS new_field;
```

### Running Migrations

```bash
make migrate-up    # Apply migrations
make migrate-down  # Rollback last migration
```

## Debugging

### API Logs

```bash
# API logs are written to api.log in project root
tail -f api.log
```

### Worker Logs

```bash
# Worker logs are written to worker.log
tail -f worker.log
```

### Database Queries

Enable query logging in development:

**In `cmd/api/main.go`:**
```go
db = db.Debug()  // Enable SQL logging
```

## Environment Configuration

### Local Development

**Backend (.env):**
```bash
SERVER_MODE=debug
DATABASE_SSLMODE=disable
LOG_LEVEL=debug
```

**Frontend (web/.env.local):**
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Production Considerations

- Set `SERVER_MODE=release`
- Use strong passwords
- Enable SSL for database
- Configure CORS for production domain
- Set up monitoring and logging

## Common Development Tasks

### Adding Data Source

1. Go to Admin → Data Sources
2. Click "Add Data Source"
3. Fill in connection details
4. Test connection
5. Set permissions (groups that can access)
6. Save

### Executing Queries

1. Select data source from dropdown
2. Write SQL query in editor
3. Click "Run Query"
4. View results in table below
5. Export if needed (CSV/JSON)

### Approval Workflow

1. User submits write query (INSERT/UPDATE/DELETE)
2. Query goes to approval queue
3. Approvers see request in dashboard
4. Approver reviews SQL and makes decision
5. On approve: Worker executes query
6. On reject: User sees rejection reason

## Troubleshooting

### API Server Not Starting

```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill process if needed
kill -9 <PID>
```

### Worker Not Processing Jobs

```bash
# Check worker logs
tail -f worker.log

# Check Redis connection
redis-cli ping
```

### Frontend Build Errors

```bash
# Clear Next.js cache
cd web
rm -rf .next
npm run dev
```

### Database Connection Issues

```bash
# Check PostgreSQL status
make docker-logs | grep postgres

# Restart database
make docker-down
make docker-up
make migrate-up
```

## Performance Optimization

### Database Indexing

Add indexes for frequently queried columns:
```sql
CREATE INDEX idx_table_column ON table(column);
```

### Query Optimization

- Use EXPLAIN ANALYZE to profile queries
- Add appropriate indexes
- Limit result sets with LIMIT clause
- Use connection pooling

### Frontend Optimization

- Use React.memo for expensive components
- Implement virtualization for large lists
- Lazy load components
- Optimize bundle size

## Code Style

### Go Code Formatting

```bash
make fmt  # Format all Go code
```

### Linting

```bash
make lint  # Run linter
```

### TypeScript

```bash
cd web
npm run type-check  # Type checking
```

## Building for Production

### Native Platform

```bash
make build
```

### All Platforms

```bash
./build.sh all
```

Binaries will be in `bin/` directory:
- `api-linux-amd64`, `api-linux-arm64`
- `api-darwin-amd64`, `api-darwin-arm64`
- `api-windows-amd64.exe`
- Same for worker binaries

## Deployment

### Using Docker

```bash
# Build all images
make build-all

# Start services
docker-compose -f docker/docker-compose.yml up -d
```

### Manual Deployment

1. Build binaries for target platform
2. Set up infrastructure (PostgreSQL, Redis)
3. Run migrations
4. Configure environment variables
5. Start API server
6. Start worker
7. Deploy frontend (build `npm run build`)

## Security Best Practices

### Never Commit

- `.env` files
- Database passwords
- JWT secrets
- API keys
- Certificates

### Always Use

- Strong passwords in production
- HTTPS in production
- Environment variables for secrets
- Rate limiting on public endpoints
- Input validation
- Parameterized queries

## Getting Help

1. Check this guide first
2. Search [docs/api](docs/api/) for API documentation
3. Check [docs/features](docs/features/) for feature guides
4. Open an issue on GitHub

---

**Happy Coding! 🚀**
