# Development Documentation

Development, testing, and building guides for QueryBase.

**Last Updated:** January 29, 2026

## Quick Links

- **[Local Development Setup](setup.md)** - Complete setup guide for local development
- **[Testing Guide](testing.md)** - Unit tests, integration tests, E2E tests
- **[Build Guide](build.md)** - Multi-platform build instructions
- **[Session Summary](session-summary.md)** - Development history and status

---

## Getting Started

### Prerequisites

- **Go** 1.21+ - [Download](https://go.dev/dl/)
- **Node.js** 18+ and **npm** - [Download](https://nodejs.org/)
- **Docker** and **Docker Compose** - [Download](https://www.docker.com/products/docker-desktop)
- **Make** - Included on macOS/Linux

### Quick Setup

```bash
# 1. Start infrastructure
make docker-up

# 2. Run migrations
make migrate-up

# 3. Download dependencies
make deps

# 4. Build
make build

# 5. Run (3 terminals)
make run-api      # Terminal 1: API server
make run-worker   # Terminal 2: Background worker
cd web && npm run dev  # Terminal 3: Frontend
```

**Access:** http://localhost:3000 (admin@querybase.local / admin123)

---

## Development Workflow

### Making Backend Changes

1. Edit Go files in `internal/`
2. Restart API server: `make run-api`
3. Test changes: `make test`

### Making Frontend Changes

1. Edit React/Next.js files in `web/src/`
2. Next.js hot-reloads automatically
3. Test changes: `cd web && npm test`

### Running Tests

```bash
# Backend
make test              # All tests
make test-coverage     # With coverage
make test-auth         # Auth package only

# Frontend
cd web
npm test              # Unit tests
npm run test:e2e      # E2E tests
```

---

## Documentation

### 1. [Local Development Setup](setup.md)
**Complete Setup Guide**

Covers:
- Prerequisites and installation
- Project setup
- Running development servers
- Database management
- Configuration
- Hot reload setup
- Debugging
- Troubleshooting

**When to Read:**
- Setting up development environment for the first time
- Configuring local development tools
- Debugging setup issues

### 2. [Testing Guide](testing.md)
**Testing Strategies and Guidelines**

Covers:
- Unit tests (backend)
- Unit tests (frontend)
- Integration tests
- E2E tests with Playwright
- Test coverage
- Writing new tests
- Running tests

**When to Read:**
- Writing new tests
- Understanding test structure
- Running specific test suites
- Checking test coverage

### 3. [Build Guide](build.md)
**Multi-Platform Build Instructions**

Covers:
- Building for native platform
- Building for multiple platforms (ARM64, AMD64)
- Cross-compilation
- Docker builds
- Build troubleshooting
- Production deployment

**When to Read:**
- Building QueryBase from source
- Creating deployment packages
- Cross-compiling for different platforms
- Troubleshooting build issues

### 4. [Integration Tests](integration-tests.md)
**End-to-End API Testing**

Covers:
- Integration test setup
- API testing strategies
- Test scenarios
- Running integration tests
- Test data setup

**When to Read:**
- Writing integration tests
- Testing API endpoints
- Understanding integration test structure

### 5. [Session Summary](session-summary.md)
**Development History & Status**

Contains:
- Complete development timeline
- Session summaries
- Implementation decisions
- Feature additions
- Bug fixes and improvements
- Current project status

**When to Read:**
- Understanding project evolution
- Learning implementation details
- Reviewing decision history
- Contributing to the project

---

## Code Organization

### Backend Structure

```
internal/
├── api/                 # API layer
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Auth, CORS, logging, RBAC
│   ├── routes/          # Route definitions
│   └── dto/             # Request/response DTOs
├── models/              # GORM models
├── service/             # Business logic
├── auth/                # JWT, password hashing
├── config/              # Configuration
├── database/            # DB connections
└── queue/               # Job queue
```

### Frontend Structure

```
web/src/
├── app/                 # Next.js App Router pages
│   ├── dashboard/       # Query editor, history, approvals
│   ├── admin/           # Users, groups, data sources
│   └── login/           # Authentication
├── components/          # React components
│   ├── query/           # Query-related components
│   ├── admin/           # Admin components
│   ├── approvals/       # Approval components
│   └── layout/          # Layout components
├── stores/              # Zustand state stores
├── lib/                 # Utilities
├── types/               # TypeScript types
└── __tests__/           # Frontend unit tests
```

---

## Common Development Tasks

### Adding a New API Endpoint

1. Add DTO to `internal/api/dto/`
2. Add handler to `internal/api/handlers/`
3. Add service method to `internal/service/`
4. Register route in `internal/api/routes/`
5. Add tests

### Adding a New Frontend Page

1. Create page in `web/src/app/`
2. Add components in `web/src/components/`
3. Update types in `web/src/types/`
4. Add store if needed in `web/src/stores/`
5. Add tests

### Database Migration

1. Create up migration in `migrations/postgresql/`
2. Create down migration in `migrations/postgresql/`
3. Test: `make migrate-up` and `make migrate-down`
4. Update models if needed

---

## Development Tools

### Make Commands

```bash
# Dependencies
make deps           # Download Go dependencies

# Database
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make db-shell       # Open PostgreSQL shell

# Build
make build          # Build for current platform
make build-all      # Build for all platforms
make list           # List built binaries

# Run
make run-api        # Run API server
make run-worker     # Run background worker

# Test
make test           # Run all tests
make test-short     # Short tests only
make test-coverage  # With coverage

# Lint
make fmt            # Format code
make lint           # Run linter
make clean          # Clean build artifacts
```

### Go Tools

```bash
# Run tests
go test ./... -v

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Run
go run ./cmd/api/main.go
```

### Frontend Tools

```bash
cd web

# Development
npm run dev         # Start dev server

# Testing
npm test            # Unit tests
npm run test:e2e    # E2E tests

# Building
npm run build       # Production build
npm run start       # Run production build
```

---

## Testing Strategy

### Unit Tests
- **Backend:** Go testing framework
- **Frontend:** Jest + React Testing Library
- Fast execution
- No external dependencies

### Integration Tests
- Test with real PostgreSQL and Redis
- API endpoint testing
- Slower but comprehensive

### E2E Tests
- Playwright for browser automation
- Test complete user flows
- Slowest but most realistic

### Test Coverage
- **Target:** >80% coverage
- **Current:** ~85% (service layer)
- View with: `make test-coverage`

---

## Contributing

See [CLAUDE.md](../../CLAUDE.md) for:
- Code style guidelines
- Pull request process
- Commit message conventions
- Code review checklist

---

## Related Documentation

- **[Getting Started](../getting-started/)** - Setup and installation
- **[User Guide](../user-guide/)** - Feature usage
- **[Architecture](../architecture/)** - System design
- **[API Reference](../api/)** - API endpoints
- **[CLAUDE.md](../../CLAUDE.md)** - Complete project guide

---

## Quick Reference

### Start Development
```bash
make docker-up && make migrate-up && make deps
make build
make run-api        # Terminal 1
make run-worker     # Terminal 2
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

**Need help?** Check [setup.md](setup.md) for detailed setup instructions.
