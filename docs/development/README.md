# Development Documentation

Development, testing, and building guides for QueryBase.

## Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [Documentation](#documentation)

## Overview

This section contains documentation for:
- Testing strategies and guidelines
- Multi-platform build instructions
- Development session history
- Test failure analysis
- MySQL testing

## Prerequisites

**Required:**
- Go 1.22 or later
- Docker and Docker Compose
- Make (optional)

**Optional:**
- golangci-lint (linting)
- PostgreSQL client tools
- Redis CLI

## Getting Started

### 1. Clone and Setup

```bash
git clone https://github.com/yourorg/querybase.git
cd querybase
make docker-up      # Start PostgreSQL and Redis
make migrate-up    # Run database migrations
make deps          # Download Go dependencies
```

### 2. Run Tests

```bash
make test          # Run all tests
make test-short    # Skip database-dependent tests
make test-coverage # Run with coverage report
```

### 3. Run Development Server

```bash
make run-api       # Start API server (http://localhost:8080)
make run-worker    # Start background worker
```

## Documentation

### 1. [Testing Guide](testing.md)
**Comprehensive Testing Documentation**

Covers:
- Unit tests (table-driven tests)
- Integration tests (database-dependent)
- Test coverage
- Running tests
- Writing new tests
- Multi-database testing

**Commands:**
```bash
make test          # All tests
make test-short    # Short tests only
make test-coverage # With coverage report
make test-auth     # Auth package only
make test-service  # Service package only
```

**When to Read:**
- Writing new tests
- Understanding test structure
- Running specific test suites
- Checking test coverage

### 2. [Build Guide](build.md)
**Multi-Platform Build Instructions**

Covers:
- Building for native platform
- Building for multiple platforms (ARM64, AMD64)
- Cross-compilation
- Docker builds
- Build troubleshooting

**Supported Platforms:**
- Linux ARM64 & AMD64
- macOS ARM64 & AMD64 (Apple Silicon + Intel)
- Windows AMD64

**Commands:**
```bash
make build         # Build for current platform
make build-all     # Build for all platforms
make list          # List built binaries
```

**When to Read:**
- Building QueryBase from source
- Creating deployment packages
- Cross-compiling for different platforms
- Troubleshooting build issues

### 3. [Session Summary](session-summary.md)
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

### 4. [Test Failures](test-failures.md)
**Test Failure Analysis** (All Resolved ✅)

Documents:
- Previously failing tests
- Root cause analysis
- Solutions implemented
- Test status updates

**Status:** All tests now passing (90/90 = 100%)

**When to Read:**
- Understanding past issues
- Learning from fixes
- Test maintenance

### 5. [Test Summary](test-summary.md)
**Test Results Summary**

Quick reference for:
- Test counts by package
- Pass rates
- Coverage reports
- Benchmark results

### 6. [MySQL Testing](mysql-testing.md)
**MySQL-Specific Testing**

Covers:
- Testing with MySQL data sources
- MySQL connection testing
- MySQL-specific considerations
- Cross-database testing

## Development Workflow

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes**
   - Edit code
   - Write/update tests
   - Update documentation

3. **Test changes**
   ```bash
   make test-short    # Quick tests
   make test          # Full test suite
   make fmt          # Format code
   make lint         # Run linter
   ```

4. **Build**
   ```bash
   make build         # Build for testing
   ```

5. **Commit**
   ```bash
   git add .
   git commit -m "Description of changes"
   ```

### Running Tests

**Quick Test Cycle:**
```bash
# Run specific package tests
go test -short ./internal/service -v

# Run with race detector
go test -race ./internal/service

# Run specific test
go test -short ./internal/service -run TestExtractTableNames -v
```

**Full Test Suite:**
```bash
# All tests (including integration tests)
make test

# Short tests only (faster)
make test-short

# With coverage
make test-coverage
```

### Building

**Development Build:**
```bash
make build         # Current platform only
./bin-api-<os>-<arch>  # Run the binary
```

**Production Build:**
```bash
make build-all     # All platforms
make list          # See all binaries
```

## Code Organization

```
querybase/
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   └── worker/            # Background worker
├── internal/              # Private application code
│   ├── api/              # API layer
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # Auth, CORS, logging, RBAC
│   │   ├── routes/       # Route definitions
│   │   └── dto/          # Request/response DTOs
│   ├── models/           # GORM models
│   ├── service/          # Business logic
│   ├── auth/             # JWT, password hashing
│   ├── config/           # Configuration
│   ├── database/         # DB connections
│   └── queue/            # Job queue
├── migrations/           # SQL migrations
├── web/                  # Frontend (TODO)
├── docker/               # Docker configuration
├── config/               # YAML configuration
└── docs/                 # Documentation (this folder)
```

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
make test-auth      # Auth package only
make test-service   # Service package only

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

# Run linter
golangci-lint run

# Build
go build -o bin/api ./cmd/api

# Run
go run ./cmd/api/main.go
```

## Testing Strategy

### Unit Tests
- Test individual functions and methods
- No external dependencies
- Fast execution (< 1ms per test)
- Located in `*_test.go` files

### Integration Tests
- Test with real PostgreSQL
- Test with real Redis
- Slower execution (~100ms per test)
- Tagged with `!short` tag

### Test Coverage
- Target: >80% coverage
- Current: ~85% (service layer)
- View with: `make test-coverage`

## Contributing

See [CLAUDE.md](../../CLAUDE.md) for:
- Code style guidelines
- Pull request process
- Commit message conventions
- Code review checklist

## Related Documentation

- **[Getting Started](../getting-started/)** - Setup and installation
- **[User Guides](../guides/)** - Feature usage
- **[Architecture](../architecture/)** - System design
- **[CLAUDE.md](../../CLAUDE.md)** - Complete project guide

---

**Last Updated:** January 27, 2025
