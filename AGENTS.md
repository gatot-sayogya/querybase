# AGENTS.md - Agentic Coding Guidelines for QueryBase

This file contains essential guidelines for AI agents working on the QueryBase codebase.

## Project Overview

QueryBase is a web-based database exploration platform with:
- **Backend**: Go (Gin framework, GORM ORM, PostgreSQL primary, Redis queue)
- **Frontend**: Next.js 15+ with TypeScript, Tailwind CSS, Zustand state management
- **Architecture**: API server + background worker + Next.js frontend

## Build Commands

### Go Backend
```bash
# Build
make build              # Build all binaries (native)
make build-api          # Build API server only
make build-worker       # Build worker only

# Run
make run-api            # Run API server (localhost:8080)
make run-worker         # Run background worker
make docker-up          # Start PostgreSQL + Redis
make migrate-up         # Run database migrations

# Dependencies
make deps               # Download Go dependencies

# Lint/Format
make fmt                # Format Go code with go fmt
make lint               # Run golangci-lint
```

### Frontend (web/)
```bash
cd web
npm install
npm run dev             # Start dev server (localhost:3000)
npm run build           # Production build
npm run lint            # ESLint
```

## Test Commands

### Go Backend
```bash
# Run all tests
make test

# Run a single test file
make test-auth          # Run auth package tests
make test-service       # Run service package tests
go test -v ./internal/auth/...  # Specific package

# Run a specific test function
go test -v -run TestJWTManager_GenerateToken ./internal/auth/...
go test -v -run TestJWTManager_ValidateToken_Invalid ./internal/auth/...

# Run tests with coverage
make test-coverage      # Generate coverage report
make test-race          # Run with race detector
make test-short         # Skip DB-dependent tests
make test-bench         # Run benchmarks
```

### Frontend
```bash
cd web
npm test                # Run Jest tests
npm run test:watch      # Watch mode
npm run test:coverage   # Coverage report
npm run test:e2e        # Playwright E2E tests
npm run test:e2e:ui     # Playwright UI mode
```

## Code Style Guidelines

### Go

#### Imports
Group imports: stdlib → external → internal (blank line between groups).
```go
import (
    "context"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "github.com/yourorg/querybase/internal/models"
)
```

#### Naming Conventions
- **Exported**: PascalCase (`QueryService`, `ExecuteQuery`)
- **Unexported**: camelCase (`decryptPassword`, `activeTransactions`)
- **Interfaces**: End with "er" or descriptive noun (`Handler`, `Service`)
- **Test files**: `*_test.go` suffix
- **Test functions**: `Test<FunctionName>_<Scenario>` (`TestJWTManager_ValidateToken_Invalid`)

#### Types
- Use structs with GORM tags for data models
- Use constants with iota for enums
- Prefer explicit types over `interface{}`

#### Error Handling
- Wrap with context: `fmt.Errorf("failed to connect: %w", err)`
- Use `%w` for error wrapping
- Return errors; don't log and return nil
- Handle errors explicitly (no `_` for error returns)

#### Functions
- Keep functions under 50 lines when possible
- Use `context.Context` as first parameter
- Return `(result, error)` pattern
- Use named return values sparingly

#### Structs
```go
type QueryService struct {
    db                 *gorm.DB
    encryptionKey      []byte
    activeTransactions map[uuid.UUID]*ActiveTransaction
    txMutex            sync.RWMutex
}
```

### TypeScript/React (Frontend)

#### Imports Order
1. React hooks
2. Third-party libraries
3. Internal components/utilities
4. Types

#### Naming Conventions
- **Components**: PascalCase (`Button.tsx`, `QueryResults.tsx`)
- **Hooks**: camelCase with `use` prefix (`useDashboardStats`)
- **Utilities**: camelCase (`cn`, `formatDate`)
- **Types/Interfaces**: PascalCase (`ButtonProps`, `User`)

#### Component Structure
```typescript
import { forwardRef } from 'react';
import { cn } from '@/lib/utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary';
  loading?: boolean;
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', ...props }, ref) => {
    // implementation
  }
);

Button.displayName = 'Button';
export default Button;
```

#### Styling
- Use Tailwind CSS exclusively
- Use `cn()` utility for conditional classes
- Dark mode support with `dark:` prefix

## Testing Guidelines

### Go Tests
- Use testify: `assert` for assertions, `require` for fatal checks
- Table-driven tests for multiple scenarios
- Mock external dependencies
- Use `t.Parallel()` for independent tests

### Frontend Tests
- Use React Testing Library
- Test user interactions, not implementation
- Mock API calls with MSW or jest mocks

## Project Structure

```
cmd/
  api/                  # API server entry point
  worker/               # Background worker entry point

internal/
  api/                  # HTTP layer
    dto/                # Data transfer objects
    handlers/           # HTTP handlers
    middleware/         # Auth, CORS, rate limiting
    routes/             # Route definitions
  auth/                 # JWT, password hashing
  config/               # Configuration management
  database/             # DB connections
  models/               # GORM models
  queue/                # Background job definitions
  repository/           # Data access layer
  service/              # Business logic
  validation/           # Input validation

web/
  src/
    app/                # Next.js App Router
    components/         # React components
    lib/                # Utilities
    stores/             # Zustand stores
    types/              # TypeScript types
    __tests__/          # Unit tests
  e2e/                  # Playwright tests
```

## Important Notes

- **Never commit secrets**: `.env` files contain sensitive data
- **Run tests before committing**: `make test` for Go, `npm test` for frontend
- **Format code**: `make fmt` formats Go code automatically
- **Database migrations**: Always use migration files in `migrations/`
- **Security**: Never log or expose passwords, tokens, or encryption keys

<!-- BEGIN BEADS INTEGRATION -->
## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Dolt-powered version control with native sync
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**

```bash
bd ready --json
```

**Create new issues:**

```bash
bd create "Issue title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd create "Issue title" --description="What this issue is about" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**

```bash
bd update <id> --claim --json
bd update bd-42 --priority 1 --json
```

**Complete work:**

```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task atomically**: `bd update <id> --claim`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" --description="Details about what was found" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs via Dolt:

- Each write auto-commits to Dolt history
- Use `bd dolt push`/`bd dolt pull` for remote sync
- No manual export/import needed!

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and docs/QUICKSTART.md.

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

<!-- END BEADS INTEGRATION -->
