# Code Style and Conventions for QueryBase

## Go Code Style

### Imports
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

### Naming Conventions
- **Exported identifiers:** PascalCase (`QueryService`, `ExecuteQuery`)
- **Unexported identifiers:** camelCase (`decryptPassword`, `activeTransactions`)
- **Interfaces:** End with "er" or descriptive noun (`Handler`, `Service`)
- **Test files:** `*_test.go` suffix
- **Test functions:** `Test<FunctionName>_<Scenario>` (e.g., `TestJWTManager_ValidateToken_Invalid`)

### Types
- Use structs with GORM tags for data models
- Use constants with iota for enums
- Prefer explicit types over `interface{}`

### Error Handling
- Wrap with context: `fmt.Errorf("failed to connect: %w", err)`
- Use `%w` for error wrapping
- Return errors; don't log and return nil
- Handle errors explicitly (no `_` for error returns)

### Functions
- Keep functions under 50 lines when possible
- Use `context.Context` as first parameter
- Return `(result, error)` pattern
- Use named return values sparingly

### Structs
```go
type QueryService struct {
    db                 *gorm.DB
    encryptionKey      []byte
    activeTransactions map[uuid.UUID]*ActiveTransaction
    txMutex            sync.RWMutex
}
```

### Testing
- Use testify: `assert` for assertions, `require` for fatal checks
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Use `t.Parallel()` for independent tests

## TypeScript/React Code Style

### Imports Order
1. React hooks
2. Third-party libraries
3. Internal components/utilities
4. Types

### Naming Conventions
- **Components:** PascalCase (`Button.tsx`, `QueryResults.tsx`)
- **Hooks:** camelCase with `use` prefix (`useDashboardStats`)
- **Utilities:** camelCase (`cn`, `formatDate`)
- **Types/Interfaces:** PascalCase (`ButtonProps`, `User`)

### Component Structure
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

### Styling
- Use Tailwind CSS exclusively
- Use `cn()` utility for conditional classes
- Dark mode support with `dark:` prefix

### Testing
- Use React Testing Library
- Test user interactions, not implementation
- Mock API calls with MSW or jest mocks

## Project Structure Conventions

### Backend (Go)
```
cmd/
  api/main.go              # API server entry
  worker/main.go           # Worker entry

internal/
  api/
    dto/                   # Data Transfer Objects
    handlers/              # HTTP handlers
    middleware/            # Auth, CORS, rate limiting
    routes/                # Route definitions
  auth/                    # JWT, password hashing
  config/                  # Configuration
  database/                # DB connections
  models/                  # GORM models
  queue/                   # Background jobs
  repository/              # Data access layer
  service/                 # Business logic
  validation/              # Input validation
```

### Frontend (Next.js)
```
web/src/
  app/                     # Next.js App Router pages
  components/              # React components
  lib/                     # Utilities
  stores/                  # Zustand stores
  types/                   # TypeScript types
  __tests__/               # Unit tests
```

## Security Guidelines
- Never commit secrets (`.env` files)
- Never log or expose passwords, tokens, or encryption keys
- Use environment variables for sensitive data
- Use AES-256 for data source credentials
- Use bcrypt for password hashing
