# Decisions - QueryBase Comprehensive Testing

## Test Infrastructure Decisions

### Database Testing
- Decision: Use testcontainers-go for integration tests
- Rationale: Provides real PostgreSQL without production dependencies

### Authentication Mocking
- Decision: Create dedicated auth test utilities
- Rationale: Consistent mocking across all handler tests

### Fixture Pattern
- Decision: Use factory functions that return models
- Rationale: Allows flexible test data creation with cleanup

## RBAC Testing Strategy

### Permission Resolution
- Test at service layer for logic
- Test at handler layer for enforcement
- Test E2E for user experience

### Coverage Goals
- Handler coverage >80%
- Service coverage >70%
- All permission paths tested
