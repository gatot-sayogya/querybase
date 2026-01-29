# QueryBase Development Guide

**Last Updated:** January 29, 2026

## Current Implementation Status

**Backend Completion: ~95%** ✅
**Frontend Completion: ~90%** ✅

### Completed Features

#### Core Infrastructure
- ✅ Database layer with PostgreSQL and MySQL support
- ✅ GORM models for all entities
- ✅ JWT authentication with bcrypt password hashing
- ✅ Configuration management (YAML + environment variables)
- ✅ Docker development environment
- ✅ Multi-architecture build system

#### API Layer
- ✅ RESTful API with Gin framework
- ✅ Authentication middleware
- ✅ Role-Based Access Control (RBAC) middleware
- ✅ CORS middleware (environment-configurable)
- ✅ Rate limiting (query execution only)
- ✅ All DTOs defined

#### Query Engine
- ✅ SQL parser for operation detection
- ✅ Query execution service
- ✅ Result caching in PostgreSQL
- ✅ Query history tracking
- ✅ Transaction support (start/commit/rollback)
- ✅ Query result pagination

#### Approval Workflow
- ✅ Approval request creation
- ✅ Single-stage approval process
- ✅ Approval review handlers
- ✅ Transaction management
- ✅ Comment system on approvals

#### Data Source Management
- ✅ CRUD operations for data sources
- ✅ PostgreSQL and MySQL support
- ✅ Encrypted password storage
- ✅ Connection testing
- ✅ Permission system (can_read, can_write, can_approve)
- ✅ Permission-based data source filtering

#### Schema Management
- ✅ Schema inspection (tables, columns, types)
- ✅ Polling-based schema synchronization (60s)
- ✅ Manual "Sync Now" functionality
- ✅ Background worker sync (every 5 minutes)
- ✅ Health tracking for data sources
- ✅ Schema caching with 5-minute freshness

#### Background Jobs
- ✅ Redis queue (Asynq)
- ✅ Background worker process
- ✅ Schema sync tasks
- ✅ Query execution tasks
- ✅ Notification tasks

#### User & Group Management
- ✅ User CRUD operations
- ✅ Group CRUD operations
- ✅ User-group assignment
- ✅ Group-based permissions

#### Frontend
- ✅ Next.js 15+ with App Router
- ✅ TypeScript + Zustand state management
- ✅ Tailwind CSS styling
- ✅ SQL editor with Monaco Editor
- ✅ Intelligent autocomplete (tables/columns)
- ✅ Query results viewer (pagination, sorting)
- ✅ Approval dashboard
- ✅ Admin panel (users, groups, data sources)
- ✅ Schema browser with polling
- ✅ Permission-based UI filtering
- ✅ Authentication and authorization
- ✅ Query history and export

### Partially Implemented

#### Notification System
- ✅ Google Chat webhook integration
- ✅ Notification configuration model
- ⚠️ Notification worker (basic implementation)

### Not Implemented

#### Frontend Polish
- ⏳ Query result export UI improvements
- ⏳ Advanced query visualization
- ⏳ Query templates/saved queries library
- ⏳ Performance optimization
- ⏳ UX improvements

#### Testing
- ⏳ Comprehensive unit tests (some tests exist)
- ⏳ Integration tests
- ⏳ E2E tests (Playwright configured)

---

## Project Structure

```
querybase/
├── cmd/                          # Application entry points
│   ├── api/main.go              # API server
│   └── worker/main.go           # Background worker
│
├── internal/                    # Private Go code
│   ├── api/                     # API layer
│   │   ├── dto/                 # Data Transfer Objects
│   │   ├── handlers/            # HTTP handlers
│   │   ├── middleware/          # Auth, CORS, RBAC, rate limiting
│   │   └── routes/              # Route definitions
│   ├── auth/                    # JWT + password hashing
│   ├── config/                  # Configuration (YAML + env)
│   ├── database/                # DB connections (PostgreSQL, MySQL)
│   ├── models/                  # GORM models
│   ├── queue/                   # Asynq background jobs
│   ├── service/                 # Business logic
│   └── validation/              # Input validation
│
├── migrations/                   # Database migrations
│   ├── postgresql/              # PostgreSQL migrations
│   └── mysql/                   # MySQL data source schemas
│
├── tests/                        # Reorganized test structure
│   ├── unit/                    # Unit tests (moved from internal/)
│   └── integration/             # Integration tests
│
├── web/                         # Next.js frontend
│   ├── src/
│   │   ├── app/                  # App Router pages
│   │   ├── components/           # React components
│   │   ├── stores/               # Zustand stores
│   │   ├── lib/                  # Utilities
│   │   ├── types/                # TypeScript types
│   │   └── __tests__/            # Unit tests
│   ├── e2e/                      # Playwright E2E tests
│   └── [config files]
│
├── docker/                       # Docker configurations
├── config/                       # Configuration files
├── docs/                        # Documentation
├── Makefile                      # Build/run commands
└── [build scripts]
```

---

## Technology Stack

### Backend
- **Language:** Go 1.21+
- **Framework:** Gin
- **ORM:** GORM
- **Primary Database:** PostgreSQL 15
- **Cache/Queue:** Redis 7 (Asynq)
- **Auth:** JWT (golang-jwt/jwt)
- **Password:** bcrypt

### Frontend
- **Framework:** Next.js 15.5.10+ (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Editor:** Monaco Editor
- **State:** Zustand
- **HTTP:** Axios

### DevOps
- **Containers:** Docker + Docker Compose
- **Build:** Make + shell scripts
- **Migrations:** Manual SQL

---

## Development Workflow

### 1. Setup

```bash
# Start infrastructure
make docker-up

# Run migrations
make migrate-up

# Download dependencies
make deps

# Build
make build
```

### 2. Development

```bash
# Terminal 1: API server
make run-api

# Terminal 2: Worker
make run-worker

# Terminal 3: Frontend
cd web && npm run dev
```

### 3. Testing

```bash
# Backend tests
make test
make test-coverage

# Frontend tests
cd web && npm test
cd web && npm run test:e2e
```

### 4. Building for Production

```bash
# Native platform
make build

# All platforms (ARM64 + AMD64)
./build.sh all
```

---

## Key Implementation Details

### Schema Synchronization

**Architecture:** Polling-based (not WebSocket)

**Mechanism:**
1. Frontend polls every 60 seconds
2. Backend caches schema for 5 minutes
3. Background worker syncs all schemas every 5 minutes
4. Manual "Sync Now" button forces immediate sync

**Benefits:**
- ✅ No 429 errors (rate limiting bypassed)
- ✅ Reduced API calls (cache-first)
- ✅ Real-time enough for schema changes
- ✅ Simpler than WebSocket implementation

### Rate Limiting

**Current Policy:**
- **Rate Limited:** Query execution only (`/api/v1/queries`)
- **Not Rate Limited:**
  - Schema endpoints
  - Authentication
  - Data sources
  - Approvals
  - Groups
  - Health check

**Configuration:** Token bucket algorithm (60 req/min)

### CORS Configuration

**Priority Order:**
1. Environment variables (`CORS_ALLOWED_ORIGINS`)
2. Config file (`config/config.yaml`)
3. Code defaults

**Setup:** See [docs/api/CORS_SETUP.md](docs/api/CORS_SETUP.md)

### Permission System

**Three Levels:**
- `can_read`: Execute SELECT queries
- `can_write`: Submit write operation requests
- `can_approve`: Approve/reject write operations

**Access Control:**
- Group-based permissions per data source
- Users inherit group permissions
- UI filters data sources by permissions
- API enforces permissions on all operations

---

## Common Issues & Solutions

### Issue: "setWebSocketStatus is not a function"
**Cause:** Old WebSocket code still present
**Fix:** Removed WebSocket provider, switched to polling

### Issue: "429 Error when switching menus"
**Cause:** Rate limiting on schema endpoints
**Fix:** Rate limiting now only applies to query execution

### Issue: Autocomplete not showing tables/columns
**Cause:** Schema not loaded or editor not mounted yet
**Fix:**
1. Wait for schema to load (check "Last sync" time)
2. Try clicking "Sync Now" if needed
3. Check browser console for errors

### Issue: Worker not syncing schemas
**Cause:** Encryption key not configured
**Fix:** Worker now uses `cfg.JWT.Secret` from config

---

## Recent Changes (January 2026)

### Week 4: Polish & Optimization
- ✅ Fixed autocomplete semicolon bug
- ✅ Reorganized project structure
- ✅ Separated tests into `/tests` folder
- ✅ Organized migrations by database type
- ✅ Updated documentation

### Week 3: Schema Synchronization
- ✅ Implemented polling-based schema sync
- ✅ Added background worker periodic sync (5 min)
- ✅ Added "Sync Now" button
- ✅ Fixed rate limiting for schema endpoints
- ✅ Schema caching with metadata

### Week 2: Core Features
- ✅ Completed approval workflow
- ✅ Implemented query execution with transaction support
- ✅ Added data source permission filtering
- ✅ Created admin panels (users, groups, data sources)
- ✅ Built SQL editor with autocomplete

### Week 1: Infrastructure
- ✅ Set up project structure
- ✅ Configured PostgreSQL and Redis
- ✅ Implemented authentication
- ✅ Created database models
- ✅ Built API endpoints

---

## Development Guidelines

### Adding New Features

1. **Backend:**
   - Add model to `internal/models/`
   - Add DTO to `internal/api/dto/`
   - Add service to `internal/service/`
   - Add handler to `internal/api/handlers/`
   - Register route in `internal/api/routes/`

2. **Frontend:**
   - Add component to `web/src/components/`
   - Add page to `web/src/app/`
   - Update types in `web/src/types/`
   - Add store if needed in `web/src/stores/`

3. **Database:**
   - Create up migration in `migrations/postgresql/`
   - Create down migration in `migrations/postgresql/`
   - Test with `make migrate-up` and `make migrate-down`

### Testing Before Committing

1. Run `make fmt` (Go code formatting)
2. Run `make test` (unit tests)
3. Test manually in browser
4. Check for linter warnings

### Git Commit Messages

```
feat: add new feature
fix: fix bug in existing feature
docs: update documentation
refactor: refactor code structure
test: add tests
chore: update dependencies
```

---

## Next Steps

### Immediate (This Week)
- [ ] Add request logging middleware
- [ ] Write integration tests
- [ ] Add E2E tests for critical paths

### Short-term (This Month)
- [ ] Implement remaining notification features
- [ ] Add query result export functionality
- [ ] Create comprehensive test suite

### Long-term (Next Quarter)
- [ ] Query templates/saved queries library
- [] Advanced query visualization (charts, graphs)
- [] Query performance analysis
- [ ] Multi-database JOIN support

---

## File Organization

### Test Files Location
- **Backend:** `/tests/unit/` (moved from `internal/`)
- **Frontend:** `web/src/__tests__/` and `web/e2e/`

### Migration Files
- **PostgreSQL:** `/migrations/postgresql/`
- **MySQL Schemas:** `/migrations/mysql/` (for data sources)

### Documentation
- **Root:** `/README.md` (overview)
- **Development:** This file (`/CLAUDE.md`)
- **Detailed:** `/docs/` (organized by topic)

---

## Quick Reference

### Start Development
```bash
make docker-up && make migrate-up && make deps
make build
make run-api  # Terminal 1
make run-worker  # Terminal 2
cd web && npm run dev  # Terminal 3
```

### Run Tests
```bash
make test-coverage
cd web && npm test
```

### Build for Production
```bash
make build-all
```

---

**This guide is maintained alongside the codebase.** For the latest project status, always check the main README.md file.
