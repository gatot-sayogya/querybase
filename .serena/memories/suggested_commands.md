# Suggested Commands for QueryBase

## Go Backend Commands

### Build
- `make build` - Build all binaries (native)
- `make build-api` - Build API server only
- `make build-worker` - Build worker only
- `make build-all` - Build for all architectures

### Run
- `make run-api` - Run API server (http://localhost:8080)
- `make run-worker` - Run background worker
- `make docker-up` - Start PostgreSQL and Redis
- `make docker-down` - Stop Docker services
- `make docker-logs` - View Docker logs

### Database
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback migrations
- `make migrate-status` - Check migration status
- `make db-shell` - Open PostgreSQL shell

### Dependencies
- `make deps` - Download Go dependencies
- `go mod tidy` - Clean up dependencies

### Lint/Format
- `make fmt` - Format Go code with go fmt
- `make lint` - Run golangci-lint

### Testing
- `make test` - Run all tests
- `make test-short` - Run short tests (skip DB tests)
- `make test-race` - Run tests with race detector
- `make test-coverage` - Generate coverage report
- `make test-coverage-html` - Generate HTML coverage report
- `make test-bench` - Run benchmarks
- `make test-auth` - Run auth package tests
- `make test-service` - Run service package tests
- `go test -v ./internal/auth/...` - Run specific package
- `go test -v -run TestJWTManager_GenerateToken ./internal/auth/...` - Run specific test

### Clean
- `make clean` - Clean build artifacts
- `make list` - List available build artifacts

## Frontend Commands (in web/ directory)

### Development
- `npm run dev` - Start dev server (localhost:3000)
- `npm install` - Install dependencies

### Build
- `npm run build` - Production build
- `npm run start` - Start production server

### Lint/Format
- `npm run lint` - Run ESLint

### Testing
- `npm test` - Run Jest tests
- `npm run test:watch` - Watch mode
- `npm run test:coverage` - Coverage report
- `npm run test:e2e` - Playwright E2E tests
- `npm run test:e2e:ui` - Playwright UI mode
- `npm run test:e2e:headed` - Playwright headed mode
- `npm run test:e2e:debug` - Playwright debug mode

## System Commands (Darwin/macOS)

### File Operations
- `ls` - List directory contents
- `cd <dir>` - Change directory
- `pwd` - Print working directory
- `cat <file>` - Display file contents
- `mkdir <dir>` - Create directory
- `rm <file>` - Remove file
- `rm -r <dir>` - Remove directory recursively
- `cp <src> <dst>` - Copy file
- `mv <src> <dst>` - Move/rename file

### Search
- `grep <pattern> <file>` - Search for pattern in file
- `find . -name "*.go"` - Find files by name
- `grep -r <pattern> .` - Recursive search

### Process Management
- `ps aux` - List running processes
- `kill <pid>` - Kill process
- `lsof -i :8080` - Check what's using port 8080

### Git
- `git status` - Check repository status
- `git add <file>` - Stage changes
- `git commit -m "message"` - Commit changes
- `git push` - Push to remote
- `git pull` - Pull from remote
- `git log --oneline -10` - View recent commits
- `git diff` - View uncommitted changes

## Issue Tracking (bd/beads)

- `bd ready --json` - Show issues ready to work
- `bd list --status=open --json` - List open issues
- `bd show <id> --json` - Show issue details
- `bd create "title" -t bug|feature|task -p 0-4 --json` - Create issue
- `bd update <id> --claim --json` - Claim issue
- `bd close <id> --reason "Done" --json` - Close issue
- `bd dep add <from> <to> --type blocks --json` - Add dependency
