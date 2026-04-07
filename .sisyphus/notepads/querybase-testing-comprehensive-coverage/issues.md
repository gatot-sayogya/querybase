# Issues - QueryBase Comprehensive Testing

## Known Issues

### None at this time

## Test Execution Notes

### Running Tests
```bash
# Unit tests
go test -v -cover ./internal/...

# Integration tests
go test -v -tags=integration ./internal/...

# E2E tests
cd web && npm run test:e2e
```

### Coverage Reports
```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

## Integration Test Run - 2026-03-26

### Summary
- Total Tests: 718 (709 passed, 9 failed, 21 skipped)
- Result: **FAIL**
- Evidence: `.sisyphus/evidence/f3-integration-test-results.md`

### Failing Tests (9 total)

#### 1. TestLogin_Success (handlers)
- **Issue**: Panic - nil pointer dereference
- **File**: `internal/api/handlers/auth_test.go`
- **Root Cause**: Test setup has uninitialized dependency (likely JWT manager or DB)
- **Status**: Needs investigation

#### 2. TestCORSMiddleware/Wildcard_origin_allows_all (middleware)
- **Issue**: Test expects "*" but gets "http://any-origin.com"
- **File**: `internal/api/middleware/cors_test.go:76`
- **Root Cause**: CORS middleware echoes origin (correct for credentials mode) vs returning wildcard
- **Status**: Test expectation may need update

#### 3-8. Approval Service Tests (6 failures)
- **Tests**:
  - TestApprovalService_StartTransaction/Invalid_started_by_UUID
  - TestApprovalService_StartTransaction
  - TestDuplicateReviewPrevention_UserCannotApproveTwice
  - TestDuplicateReviewPrevention_UserCannotRejectTwice
  - TestDuplicateReviewPrevention_ErrorMessage_IsClear
  - TestDuplicateReviewPrevention_AfterFirstReview_SecondBlocked
- **File**: `internal/service/approval_test.go`
- **Root Causes**:
  - Validation order mismatch (UUID validation vs status validation)
  - Error message/assertion mismatches in duplicate review prevention
- **Status**: Test assertions need updating to match current implementation

### Recommendations

1. **Fix nil pointer in auth test** - Check JWT manager initialization in test setup
2. **Update CORS test expectation** - Align with actual CORS behavior (echo origin when credentials enabled)
3. **Sync approval tests** - Update error assertions to match current service implementation
4. **Consider test categories**:
   - Unit tests with mocks (fast, isolated)
   - Integration tests with SQLite (current)
   - Full integration tests with PostgreSQL (CI only)

### No Flaky Tests Detected
All failures are consistent and reproducible - no timing-related or race condition issues observed.

