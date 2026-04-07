# Learnings - QueryBase Comprehensive Testing

## Project Conventions

### RBAC System
- **Admin Role**: Full bypass of all permission checks
- **User Role**: Subject to group-based permissions
- **Viewer Role**: Read-only access, cannot execute writes

### Permission Model
```
User → UserGroup → Group → DataSourcePermission
```

Permissions are unioned across all groups a user belongs to.

### Key Permission Functions
- `GetEffectivePermissions()` - Resolves effective permissions for a user on a datasource
- `checkReadPermission()` - Handler helper for read operations
- `checkWritePermission()` - Handler helper for write operations
- `RequireAdmin()` - Middleware for admin-only endpoints

### Test Patterns
- Use `internal/testutils/fixtures` for test data
- Use `internal/testutils/auth` for authentication mocking
- Handler tests use `httptest.NewRecorder`
- Service tests may use testcontainers for integration

## File Locations

### Handlers
- `internal/api/handlers/query.go` - Query handlers
- `internal/api/handlers/approval.go` - Approval handlers
- `internal/api/handlers/multi_query.go` - Multi-query handlers

### Services
- `internal/service/query.go` - Query execution, permission resolution
- `internal/service/approval.go` - Approval workflow
- `internal/service/multi_query_service.go` - Multi-query transactions

### Middleware
- `internal/api/middleware/rbac.go` - RBAC middleware
- `internal/api/middleware/auth.go` - Authentication middleware

### Test Files
- `internal/api/handlers/query_test.go`
- `internal/api/handlers/approval_test.go`
- `internal/api/handlers/multi_query_test.go`
- `internal/service/permission_test.go`
