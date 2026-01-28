# QueryBase

A database explorer system with query execution, approval workflows, and group-based access control.

## Features

- **Query Execution**: Run SQL queries against PostgreSQL and MySQL data sources
- **Transaction-Based Approval Workflow**: Preview write operations before committing ✨ NEW
- **SQL Validation**: Validate SQL syntax before submission ✨ NEW
- **Group-Based Access Control**: RBAC similar to Redash
- **Google Chat Integration**: Webhook notifications for approvals and query results
- **User Management**: Complete CRUD operations for users and groups
- **Query History**: Track all query executions with pagination
- **Multi-Architecture Support**: Binaries for ARM64 and AMD64

## Tech Stack

- **Backend**: Go 1.22 (Gin framework, GORM)
- **Frontend**: Next.js + Tailwind CSS (planned)
- **Database**: PostgreSQL 15
- **Queue**: Redis 7 (Asynq for job processing)
- **Data Sources**: PostgreSQL, MySQL

## Documentation

**📚 Complete Documentation:** [docs/](docs/)

- **[Getting Started](docs/getting-started/)** - Quick start and setup
- **[User Guides](docs/guides/)** - How to use QueryBase features
- **[Architecture](docs/architecture/)** - System design and flow diagrams
- **[Development](docs/development/)** - Testing, building, and contributing
- **[Features](docs/features/)** - Feature implementation details

**Quick Links:**
- [Query Features Guide](docs/guides/query-features.md) - EXPLAIN and Dry Run
- [Quick Reference](docs/guides/quick-reference.md) - Daily usage reference
- [Flow Diagrams](docs/architecture/flow.md) - Visual system flow
- [Testing Guide](docs/development/testing.md) - How to test

## Project Status

**Backend: 98% Complete** ✅

### ✅ Fully Implemented
- Database schema and migrations (4 migrations applied)
- GORM models (User, Group, DataSource, Query, Approval, QueryTransaction ✨ NEW)
- JWT authentication with bcrypt password hashing
- User management (CRUD + password change)
- Group management (CRUD + user assignment)
- Query execution with result storage
- SQL parser for operation detection and validation ✨ UPDATED
- **Transaction-based approval workflow** (preview before commit) ✨ NEW
- Data source management (PostgreSQL & MySQL)
- Redis queue with background worker (for job processing)
- Google Chat notifications
- Group-based RBAC permissions
- Multi-architecture build support
- Query history pagination API

### 🚧 TODO (Remaining 2%)

**Current Focus: Core Workflow + Dashboard UI** 🎯
**Backend Polish:** See [Core Workflow Plan](docs/CORE_WORKFLOW_PLAN.md)
- Query results pagination
- Query export (CSV/JSON)
- Approval comments
- Data source health checks

**Frontend Development:** See [Dashboard UI - Current Workflow](docs/DASHBOARD_UI_CURRENT_WORKFLOW.md) ✨ HIGH PRIORITY
- Phase 1-2: Foundation (auth, layout)
- Phase 3-4: SQL Editor & Results
- Phase 5: Approval Dashboard
- Phase 6: Admin Features
- Phase 7-8: Polish & Optimization

**Future Features (Backend):**
- Schema Introspection API
- Folder System
- Tag System
- WebSocket Support

**Infrastructure:**
- Performance benchmarks
- CORS middleware
- Request logging middleware
- Rate limiting middleware
- Encrypted frontend-backend communication
- Unit tests
- Integration tests

## Quick Start

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- Make

### 1. Start Services

```bash
# Start PostgreSQL and Redis
make docker-up
```

### 2. Run Migrations

```bash
make migrate-up
```

This creates the database schema and a default admin user:
- **Email**: admin@querybase.local
- **Username**: admin
- **Password**: admin123 (⚠️ CHANGE IN PRODUCTION!)

### 3. Run API Server

```bash
make deps
make run-api
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/change-password` - Change password

### User Management (Admin Only)
- `POST /api/v1/auth/users` - Create user
- `GET /api/v1/auth/users` - List all users
- `GET /api/v1/auth/users/:id` - Get user details with groups
- `PUT /api/v1/auth/users/:id` - Update user
- `DELETE /api/v1/auth/users/:id` - Delete user

### Group Management (Admin Only)
- `POST /api/v1/groups` - Create group
- `GET /api/v1/groups` - List all groups (paginated)
- `GET /api/v1/groups/:id` - Get group details with users
- `PUT /api/v1/groups/:id` - Update group
- `DELETE /api/v1/groups/:id` - Delete group
- `POST /api/v1/groups/:id/users` - Add user to group
- `DELETE /api/v1/groups/:id/users` - Remove user from group
- `GET /api/v1/groups/:id/users` - List users in group

### Queries
- `POST /api/v1/queries` - Execute query (SELECT runs immediately, writes create approval)
- `POST /api/v1/queries/save` - Save query for later use
- `GET /api/v1/queries` - List queries (paginated)
- `GET /api/v1/queries/:id` - Get query details with results
- `DELETE /api/v1/queries/:id` - Delete saved query
- `GET /api/v1/queries/history` - List query execution history (paginated) ✨ NEW
- `POST /api/v1/queries/validate` - Validate SQL syntax before submission ✨ NEW

### Approvals
- `GET /api/v1/approvals` - List approval requests
- `GET /api/v1/approvals/:id` - Get approval details
- `POST /api/v1/approvals/:id/review` - Review (approve/reject)
- `POST /api/v1/approvals/:id/transaction-start` - Start transaction for preview ✨ NEW

### Transactions ✨ NEW
- `POST /api/v1/transactions/:id/commit` - Commit an active transaction
- `POST /api/v1/transactions/:id/rollback` - Rollback an active transaction
- `GET /api/v1/transactions/:id` - Get transaction status

### Data Sources
- `GET /api/v1/datasources` - List data sources
- `POST /api/v1/datasources` - Create data source (admin)
- `GET /api/v1/datasources/:id` - Get data source details
- `PUT /api/v1/datasources/:id` - Update data source (admin)
- `DELETE /api/v1/datasources/:id` - Delete data source (admin)
- `POST /api/v1/datasources/:id/test` - Test connection
- `GET /api/v1/datasources/:id/permissions` - Get permissions (admin)
- `PUT /api/v1/datasources/:id/permissions` - Set permissions (admin)
- `GET /api/v1/datasources/:id/approvers` - Get eligible approvers

### Health Check
- `GET /health` - API health status

## Development Commands

```bash
# Docker
make docker-up          # Start PostgreSQL and Redis
make docker-down        # Stop services
make docker-logs        # View logs

# Database
make migrate-up        # Run migrations
make migrate-down      # Rollback migrations
make db-shell          # Open PostgreSQL shell

# Build (Native Architecture)
make build             # Build all binaries for native platform
make build-api         # Build API server (native)
make build-worker      # Build worker (native)

# Build (All Architectures)
make build-all         # Build for all platforms (ARM64 + AMD64)
make build-api-multi   # Build API server for all platforms
make build-worker-multi# Build worker for all platforms
./build.sh native      # Build script for native platform
./build.sh all        # Build script for all platforms

# List Binaries
make list              # List all built binaries with sizes

# Run
make run-api           # Run API server
make run-worker        # Run worker

# Development
make deps              # Download dependencies
make test              # Run tests
make test-coverage     # Run tests with coverage
make fmt               # Format code
make lint              # Run linter
make clean             # Clean artifacts
```

### Multi-Architecture Build

QueryBase supports building for multiple architectures:

**Supported Platforms:**
- Linux ARM64 (aarch64) & AMD64 (x86_64)
- macOS ARM64 (Apple Silicon) & AMD64 (Intel)
- Windows AMD64 (x86_64)

**Build for all platforms:**
```bash
make build-all
# or
./build.sh all
```

**Build for specific platform:**
```bash
./build.sh linux-arm64
./build.sh darwin-arm64
./build.sh windows-amd64
```

See [BUILD.md](BUILD.md) for detailed build instructions.

## Configuration

Configuration can be set via:
1. `config/config.yaml` - Main configuration file
2. `.env` - Environment variables override (see `.env.example`)
3. Environment variables - Runtime overrides

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
  secret: change-this-secret-in-production
  expire_hours: 24h
```

**⚠️ Security Notes:**
- Change `jwt.secret` in production
- Change default admin password immediately
- Use strong passwords for database
- Enable SSL for database connections in production
- Set `server.mode` to `release` in production

## Project Structure

```
querybase/
├── cmd/
│   ├── api/main.go                 # API server entry point
│   └── worker/main.go              # Background worker entry point
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP handlers (auth, query, approval, datasource, group)
│   │   ├── middleware/            # Auth, CORS, logging, RBAC middleware
│   │   ├── routes/                # Route definitions
│   │   └── dto/                   # Request/response DTOs
│   ├── models/                    # GORM models (User, Group, DataSource, Query, Approval, Notification)
│   ├── service/                   # Business logic
│   ├── queue/                     # Asynq job queue
│   ├── database/                  # DB connections (PostgreSQL, MySQL)
│   ├── auth/                      # JWT, password hashing
│   └── config/                    # Configuration loading
├── migrations/                     # SQL migrations
├── docker/                         # Docker configuration (docker-compose, Dockerfiles)
├── config/
│   └── config.yaml                # Application configuration
├── web/                            # Frontend (Next.js - TODO)
├── build.sh                        # Multi-architecture build script
├── BUILD.md                        # Build guide
├── go.mod, go.sum                  # Go dependencies
├── Makefile                        # Development commands
├── .env.example                    # Environment variables template
├── README.md                       # This file
└── CLAUDE.md                      # Developer guide
```

## Security Notes

- ⚠️ Change the default admin password immediately
- ⚠️ Update `jwt.secret` in production
- ⚠️ Use environment variables for sensitive data
- ⚠️ Enable SSL for database connections in production
- ⚠️ Set `server.mode` to `release` in production
- ⚠️ Use strong passwords and enable SSL for external data sources

## Current Status

### Backend: 98% Complete ✅

**Implemented Features:**
- ✅ Database schema with 14 tables (including query_transactions) ✨ NEW
- ✅ GORM models with relationships (including QueryTransaction) ✨ NEW
- ✅ JWT authentication with bcrypt
- ✅ User management (Create, Read, Update, Delete, Change Password)
- ✅ Group management (Create, Read, Update, Delete, User Assignment)
- ✅ Query execution engine with SQL parser and validation ✨ UPDATED
- ✅ Query result storage in PostgreSQL
- ✅ **Transaction-based approval workflow** (preview before commit) ✨ NEW
- ✅ Data source management (PostgreSQL & MySQL)
- ✅ Encrypted password storage (AES-256-GCM)
- ✅ Group-based RBAC with permissions
- ✅ Redis queue with Asynq (for background jobs only)
- ✅ Background worker process
- ✅ Google Chat webhook integration
- ✅ Multi-architecture build support (ARM64 + AMD64)
- ✅ Query history pagination API ✨ NEW

**Statistics:**
- ~5,500 lines of Go code
- 5 handler files (auth, query, approval, datasource, group)
- 7 DTO files (including transaction DTOs) ✨ NEW
- 35 API endpoints (including transaction endpoints) ✨ NEW
- 14 database tables (including query_transactions) ✨ NEW
- Support for 5 different platforms
- 4 database migrations applied ✨ NEW

**What's Left (~2%):**
- Middleware (CORS, logging, rate limiting)
- Performance benchmarks
- Unit tests
- Integration tests
- Frontend (Next.js + Tailwind CSS)

## Documentation

- **[README.md](README.md)** - This file (project overview)
- **[CLAUDE.md](CLAUDE.md)** - Comprehensive developer guide
- **[BUILD.md](BUILD.md)** - Multi-architecture build instructions
- **[.claude/SESSION_SUMMARY.md](.claude/SESSION_SUMMARY.md)** - Development session logs

## License

MIT
