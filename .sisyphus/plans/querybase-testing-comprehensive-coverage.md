# Comprehensive Testing Plan for QueryBase Query Operations

## TL;DR

> **Quick Summary**: Implement comprehensive test coverage for all query operation workflows including single queries, multi-query transactions, approval workflows, and RBAC enforcement. Create tests that verify privilege limitations across different user roles (admin, regular user, viewer) and group memberships.
>
> **Deliverables**:
> - Complete unit test suite for handlers (handlertests/)
> - Integration test suite for query workflows
> - E2E test suite for user journeys (Playwright)
> - RBAC test matrix coverage
> - Test fixtures for users, groups, permissions
>
> **Estimated Effort**: Large (Multi-week effort)
> **Parallel Execution**: YES - Multiple waves
> **Critical Path**: Test fixtures → Handler tests → Integration tests → E2E tests

---

## Context

### Original Request
Implement comprehensive test coverage for query operations in QueryBase, including:
- Single and multi-query operations
- Preview before submit approval
- Approval/reject/review workflows
- Different users in different user groups with different privileges
- Testing privilege limitations

### Interview Summary
**Key Discussions**:
- User Role System: Admin (full access), User (normal), Viewer (read-only)
- Group Permission System: can_read, can_write, can_approve per DataSource
- Permission Resolution: Unions all group permissions; admin bypasses all
- Self-Approval Prevention: Users cannot approve their own requests
- Multi-Query Transactions: Sequential statement execution with rollback
- Preview Functionality: Dry-run before commit for write operations

**Research Findings**:
- Current Coverage: Service layer ~55% tested, Handler layer 0% tested
- Handler Tests Missing: All 20 handler functions untested
- Integration Tests: Minimal - needs PostgreSQL setup
- RBAC Tests: None exist for permission enforcement

### Exploration Results

**User/Group/Permission System**:
```
User (Role: admin|user|viewer)
  └─ UserGroup (membership)
       └─ Group
            └─ DataSourcePermission
                 ├─ can_read    → SELECT access
                 ├─ can_write   → INSERT/UPDATE/DELETE access
                 └─ can_approve → Can approve write queries

GetEffectivePermissions:
  1. Check if admin → full permissions
  2. Get all group memberships
  3. Union all group permissions for datasource
  4. Return: CanRead, CanWrite, CanApprove, CanSelect, CanInsert, CanUpdate, CanDelete
```

**Approval Workflow**:
```
1. User creates ApprovalRequest (status: pending)
2. System finds eligible approvers (can_approve on datasource, not requester)
3. Approver reviews (creates ApprovalReview with decision)
4. On approval: Start transaction → Execute → Commit
5. On rejection: Update status to rejected
6. Self-approval prevention enforced
7. Duplicate review prevention enforced
```

**Multi-Query Transaction**:
```
1. Parse statements (separated by semicolons)
2. Validate each statement
3. Preview all statements (estimated rows)
4. Create QueryTransaction (status: active)
5. Create QueryTransactionStatement records (sequence 0..N)
6. Execute statements sequentially
7. On success: Commit (status: committed)
8. On failure: Rollback (status: rolled_back)
```

---

## Work Objectives

### Core Objective
Implement comprehensive test coverage for all query operation workflows ensuring privilege enforcement across all user roles and group memberships.

### Concrete Deliverables
- [ ] `internal/api/handlers/query_test.go` - Handler unit tests
- [ ] `internal/api/handlers/multi_query_test.go` - Multi-query handler tests
- [ ] `internal/api/handlers/approval_test.go` - Approval handler tests
- [ ] `internal/service/query_integration_test.go` - Integration tests
- [ ] `internal/service/permission_test.go` - Permission resolution tests
- [ ] `internal/testutils/fixtures/` - Test fixtures for users, groups, permissions
- [ ] `internal/testutils/database/` - Database setup/teardown helpers
- [ ] `web/e2e/query-workflow.spec.ts` - E2E tests for query workflows
- [ ] `web/e2e/approval-workflow.spec.ts` - E2E tests for approval workflows
- [ ] `web/e2e/permission-matrix.spec.ts` - E2E tests for RBAC

### Definition of Done
- [ ] All handler functions have unit tests with >80% coverage
- [ ] RBAC enforcement verified for all operations
- [ ] Each user role tested against permission boundaries
- [ ] Approval workflow tested end-to-end
- [ ] Multi-query transaction tested with rollback scenarios
- [ ] E2E tests run in CI pipeline
- [ ] `go test ./...` passes with >80% coverage
- [ ] `npm run test:e2e` passes

### Must Have
- Unit tests for all query handlers
- Permission enforcement tests for all roles
- Test fixtures for creating test users with different permissions
- Integration tests for approval workflow

### Must NOT Have (Guardrails)
- No production database connections in tests (use testcontainers or mocks)
- No hardcoded credentials in test files
- No skipped tests due to "hard to test" permission scenarios
- No tests that depend on execution order

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (Go testing package, Jest, Playwright)
- **Automated tests**: TDD approach - write tests first, then fix code if needed
- **Framework**: 
  - Backend: Go `testing` + `testify` + `sqlmock` + `testcontainers-go`
  - Frontend E2E: Playwright

### QA Policy
Every task MUST include agent-executed QA scenarios.

- **Backend Unit Tests**: Run with `go test -v -cover ./...`
- **Integration Tests**: Run with `go test -v -tags=integration ./...`
- **E2E Tests**: Run with `npm run test:e2e`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — test infrastructure):
├── Task 1: Test fixtures for users, groups, permissions [quick]
├── Task 2: Database setup/teardown helpers [quick]
├── Task 3: Test utilities for authentication mocking [quick]
└── Task 4: Test data source creation helpers [quick]

Wave 2 (After Wave 1 — handler unit tests):
├── Task 5: ExecuteQuery handler tests [unspecified-high]
├── Task 6: PreviewWriteQuery handler tests [unspecified-high]
├── Task 7: PreviewInsertQuery handler tests [unspecified-high]
├── Task 8: ListQueries and QueryHistory tests [quick]
├── Task 9: ExplainQuery and DryRunDelete tests [quick]
├── Task 10: ExportQuery and GetQueryResults tests [quick]
└── Task 11: Permission check helper tests [quick]

Wave 3 (After Wave 2 — approval workflow tests):
├── Task 12: CreateApprovalRequest handler tests [unspecified-high]
├── Task 13: ReviewApproval handler tests [unspecified-high]
├── Task 14: GetApprovals handler tests [quick]
├── Task 15: Approval eligibility tests [deep]
├── Task 16: Self-approval prevention tests [quick]
└── Task 17: Duplicate review prevention tests [quick]

Wave 4 (After Wave 3 — multi-query tests):
├── Task 18: PreviewMultiQuery handler tests [unspecified-high]
├── Task 19: ExecuteMultiQuery handler tests [unspecified-high]
├── Task 20: CommitMultiQuery tests [unspecified-high]
├── Task 21: RollbackMultiQuery tests [unspecified-high]
└── Task 22: Multi-query parser integration tests [deep]

Wave 5 (After Wave 4 — RBAC enforcement tests):
├── Task 23: Admin role tests (full access) [quick]
├── Task 24: User role with read permissions tests [quick]
├── Task 25: User role with write permissions tests [quick]
├── Task 26: User role with approve permissions tests [quick]
├── Task 27: Viewer role tests (read-only) [quick]
├── Task 28: Group permission inheritance tests [deep]
└── Task 29: Permission denial scenarios tests [deep]

Wave 6 (After Wave 5 — E2E tests):
├── Task 30: Login and authentication flow E2E [visual-engineering]
├── Task 31: Single query execution E2E [visual-engineering]
├── Task 32: Approval workflow E2E [visual-engineering]
├── Task 33: Multi-query transaction E2E [visual-engineering]
├── Task 34: Permission matrix E2E [visual-engineering]
└── Task 35: Error handling E2E [visual-engineering]

Wave FINAL (After ALL tasks — verification):
├── Task F1: Test coverage verification (oracle)
├── Task F2: RBAC enforcement audit (deep)
├── Task F3: Integration test run verification (unspecified-high)
└── Task F4: E2E test run verification (unspecified-high)
```

---

## TODOs

### Wave 1: Test Infrastructure

- [x] 1. Test fixtures for users, groups, permissions

  **What to do**:
  - Create `internal/testutils/fixtures/user_fixtures.go`
  - Implement `CreateTestUser(db, role)` returning User
  - Implement `CreateTestGroup(db, name)` returning Group
  - Implement `AddUserToGroup(db, userID, groupID)` for membership
  - Implement `CreateTestDataSource(db, name)` returning DataSource
  - Implement `GrantPermission(db, groupID, dsID, canRead, canWrite, canApprove)`
  - Create fixture combinations: `SetupAdminUser()`, `SetupRegularUser()`, `SetupViewerUser()`

  **Must NOT do**:
  - Do not use production database connections
  - Do not create fixtures with hardcoded IDs that conflict
  - Do not skip cleanup in fixture functions

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard test fixture creation, straightforward patterns
  - **Skills**: []
    - No special skills needed - standard Go testing patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4)
  - **Blocks**: Tasks 5-35 (all need fixtures)
  - **Blocked By**: None (can start immediately)

  **References**:
  - `internal/models/user.go:20-48` - User model structure with role field
  - `internal/models/group.go:11-48` - Group and UserGroup models
  - `internal/models/datasource.go:83-109` - DataSourcePermission model
  - `internal/service/query.go:64-106` - GetEffectivePermissions() shows permission resolution

  **Acceptance Criteria**:
  - [ ] `go test ./internal/testutils/fixtures/...` passes
  - [ ] Each fixture function returns valid model with UUID
  - [ ] Fixture functions handle cleanup via t.Cleanup()

  **QA Scenarios**:
  ```gherkin
  Scenario: Admin user has full permissions
    Given test database is initialized
    When CreateTestUser(db, RoleAdmin) is called
    Then user.Role == "admin"
    When GetEffectivePermissions(ctx, admin.ID, datasource.ID) is called
    Then CanRead == true AND CanWrite == true AND CanApprove == true
    Evidence: .sisyphus/evidence/task-1-admin-perms.txt

  Scenario: Regular user with group permissions
    Given test database is initialized
    When SetupRegularUser(db, groupID, dsID, canRead=true, canWrite=false) is called
    Then user is added to group with correct permissions
    When GetEffectivePermissions(ctx, user.ID, datasource.ID) is called
    Then CanRead == true AND CanWrite == false
    Evidence: .sisyphus/evidence/task-1-user-perms.txt

  Scenario: Permission inheritance from multiple groups
    Given user in GroupA with can_read on DataSourceX
    And user in GroupB with can_write on DataSourceX
    When GetEffectivePermissions(ctx, user.ID, DataSourceX.ID) is called
    Then CanRead == true AND CanWrite == true
    Evidence: .sisyphus/evidence/task-1-multi-group-perms.txt
  ```

  **Commit**: YES
  - Message: `test(fixtures): add user, group, permission fixtures`
  - Files: `internal/testutils/fixtures/`
  - Pre-commit: `go test ./internal/testutils/fixtures/...`

- [x] 2. Database setup/teardown helpers

  **What to do**:
  - Create `internal/testutils/database/setup.go`
  - Implement `SetupTestDB(t)` returning *gorm.DB with cleanup
  - Implement `SetupTestDBWithContainer(t)` using testcontainers
  - Implement `CleanupTestDB(db)` for transaction rollback
  - Implement `RunTestWithTransaction(t, db, fn)` wrapper
  - Create test configuration for PostgreSQL connection
  - Add build tags for integration tests: `//go:build integration`

  **Must NOT do**:
  - Do not connect to production database
  - Do not leave test data after test completion
  - Do not share database connections across unrelated tests

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard testing infrastructure setup
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4)
  - **Blocks**: Tasks 5-35 (all need database helpers)
  - **Blocked By**: None

  **References**:
  - `internal/service/query_test.go:174-234` - Existing integration test pattern
  - `internal/service/approval_workflow_test.go:16-52` - Existing test setup pattern
  - `go.mod` for testcontainers-go dependency

  **Acceptance Criteria**:
  - [ ] `go test -v ./internal/testutils/database/...` passes
  - [ ] SetupTestDB creates clean database
  - [ ] CleanupTestDB removes all test data
  - [ ] Transaction wrapper isolates test data

  **QA Scenarios**:
  ```gherkin
  Scenario: Database setup creates clean schema
    When SetupTestDB(t) is called
    Then database has all tables created
    And database has no test data
    When test completes
    Then database is cleaned up
    Evidence: .sisyphus/evidence/task-2-db-setup.txt

  Scenario: Transaction wrapper rolls back changes
    Given RunTestWithTransaction(db, func(tx *gorm.DB) {
      tx.Create(&User{Email: "test@test.com"})
    })
    When transaction function completes
    Then User does NOT exist in database
    Evidence: .sisyphus/evidence/task-2-tx-rollback.txt

  Scenario: Testcontainers PostgreSQL setup
    When SetupTestDBWithContainer(t) is called
    Then container is running
    And connection is established
    When test completes
    Then container is stopped
    Evidence: .sisyphus/evidence/task-2-container.txt
  ```

  **Commit**: YES
  - Message: `test(db): add database setup/teardown helpers`
  - Files: `internal/testutils/database/`

- [x] 3. Test utilities for authentication mocking

  **What to do**:
  - Create `internal/testutils/auth/mock_auth.go`
  - Implement `MockAuthMiddleware()` returning gin.HandlerFunc
  - Implement `SetMockUser(c *gin.Context, userID, email, role string)`
  - Implement `CreateTestJWTToken(userID, email, role string) string`
  - Implement `MockAdminContext(c *gin.Context)` for admin role
  - Implement `MockUserContext(c *gin.Context)` for user role
  - Implement `MockViewerContext(c *gin.Context)` for viewer role
  - Implement `CreateTestJWTManager()` returning *auth.JWTManager for testing

  **Must NOT do**:
  - Do not use real JWT secrets in tests
  - Do not bypass authentication in production code paths

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard authentication mocking patterns
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4)
  - **Blocks**: Tasks 5-35 (all need auth mocking)
  - **Blocked By**: None

  **References**:
  - `internal/api/middleware/auth.go:13-53` - AuthMiddleware implementation
  - `internal/api/middleware/rbac.go:9-19` - RequireAdmin implementation
  - `internal/auth/jwt.go` - JWT token creation/validation

  **Acceptance Criteria**:
  - [ ] `go test ./internal/testutils/auth/...` passes
  - [ ] MockAuthMiddleware sets user_id, email, role in context
  - [ ] CreateTestJWTToken produces valid tokens

  **QA Scenarios**:
  ```gherkin
  Scenario: Mock admin context sets correct values
    Given gin context is created
    When MockAdminContext(c) is called
    Then c.GetString("user_id") returns valid UUID
    And c.GetString("role") == "admin"
    Evidence: .sisyphus/evidence/task-3-admin-mock.txt

  Scenario: Mock user context sets correct values
    Given gin context is created
    When MockUserContext(c) is called
    Then c.GetString("role") == "user"
    Evidence: .sisyphus/evidence/task-3-user-mock.txt

  Scenario: Test JWT token is valid
    Given CreateTestJWTToken("user-id", "test@test.com", "admin")
    When token is validated by JWT manager
    Then claims.UserID == "user-id"
    And claims.Role == "admin"
    Evidence: .sisyphus/evidence/task-3-jwt-valid.txt
  ```

  **Commit**: YES
  - Message: `test(auth): add authentication mocking utilities`
  - Files: `internal/testutils/auth/`

- [x] 4. Test data source creation helpers

  **What to do**:
  - Create `internal/testutils/datasource/fixtures.go`
  - Implement `CreateTestPostgreSQLDataSource(db, name) *models.DataSource`
  - Implement `CreateTestMySQLDataSource(db, name) *models.DataSource`
  - Implement `CreateTestDataSourceWithPerms(db, name, groupID, perms) *models.DataSource`
  - Configure test containers for PostgreSQL and MySQL
  - Implement `SetupDataSourceWithTestTable(db, tableName, columns)` for table creation

  **Must NOT do**:
  - Do not use production database credentials
  - Do not create production data sources

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard test fixture creation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3)
  - **Blocks**: Tasks 5-35
  - **Blocked By**: None

  **References**:
  - `internal/models/datasource.go:1-109` - DataSource model definition
  - `internal/service/query_test.go:174-234` - Existing test container pattern

  **Acceptance Criteria**:
  - [ ] Test data sources are created with encrypted passwords
  - [ ] Test tables can be created and dropped
  - [ ] Cleanup removes all test data sources

  **QA Scenarios**:
  ```gherkin
  Scenario: PostgreSQL datasource creation
    When CreateTestPostgreSQLDataSource(db, "test-pg") is called
    Then datasource is saved to database
    And datasource.Type == "postgresql"
    And password is encrypted
    Evidence: .sisyphus/evidence/task-4-pg-datasource.txt

  Scenario: Datasource with permissions
    Given test group exists
    When CreateTestDataSourceWithPerms(db, "test", groupID, canRead=true, canWrite=false)
    Then DataSourcePermission record exists
    And permission has correct flags
    Evidence: .sisyphus/evidence/task-4-ds-perms.txt

  Scenario: Test table creation
    Given datasource connection
    When SetupDataSourceWithTestTable(db, "test_table", "id INT, name VARCHAR(100)") is called
    Then table exists in database
    When test completes
    Then table is dropped
    Evidence: .sisyphus/evidence/task-4-test-table.txt
  ```

  **Commit**: YES
  - Message: `test(datasource): add test data source helpers`
  - Files: `internal/testutils/datasource/`

---

### Wave 2: Handler Unit Tests

- [ ] 5. ExecuteQuery handler tests

  **What to do**:
  - Create `internal/api/handlers/query_test.go`
  - Test `ExecuteQuery` handler for SELECT queries
  - Test `ExecuteQuery` handler for INSERT/UPDATE/DELETE requiring approval
  - Test `ExecuteQuery` with different user roles (admin, user, viewer)
  - Test permission denied scenarios
  - Test invalid request body handling
  - Test data source not found handling
  - Test query execution error handling
  - Test approval creation for write operations
  - Use httptest.NewRecorder for response capture
  - Mock query service and database

  **Must NOT do**:
  - Do not connect to real database (use mocks)
  - Do not skip permission check tests
  - Do not use hardcoded UUIDs that might conflict

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Critical handler with complex logic
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 6-11 in Wave 2)
  - **Parallel Group**: Wave 2
  - **Blocks**: Wave 5 RBAC tests
  - **Blocked By**: Wave 1 (Task 1, 2, 3, 4)

  **References**:
  - `internal/api/handlers/query.go:36-176` - ExecuteQuery handler implementation
  - `internal/api/handlers/query.go:527-560` - checkReadPermission, checkWritePermission
  - `internal/service/query.go:64-106` - GetEffectivePermissions
  - `internal/api/middleware/auth_middleware_test.go:19-116` - Existing middleware tests

  **Acceptance Criteria**:
  - [ ] `go test ./internal/api/handlers/... -run TestExecuteQuery` passes
  - [ ] All user roles tested
  - [ ] Permission denied cases return 403
  - [ ] Handler coverage >80%

  **QA Scenarios**:
  ```gherkin
  Scenario: Admin executes SELECT query successfully
    Given admin user is authenticated
    And data source exists
    When POST /api/v1/queries with {"data_source_id": "...", "query_text": "SELECT 1"}
    Then response status is 200
    And response contains query_id and results
    Evidence: .sisyphus/evidence/task-5-admin-select.txt

  Scenario: Regular user with read permission executes SELECT
    Given user in group with can_read on datasource
    When POST /api/v1/queries with SELECT query
    Then response status is 200
    Evidence: .sisyphus/evidence/task-5-user-select.txt

  Scenario: Regular user without permission denied
    Given user in group without datasource access
    When POST /api/v1/queries with SELECT query
    Then response status is 403
    And error message contains "permission denied"
    Evidence: .sisyphus/evidence/task-5-permission-denied.txt

  Scenario: Write query creates approval request
    Given user with write permission
    When POST /api/v1/queries with UPDATE query
    Then response status is 200
    And response contains "pending_approval"
    And ApprovalRequest is created
    Evidence: .sisyphus/evidence/task-5-write-approval.txt

  Scenario: Viewer cannot execute write queries
    Given viewer role user
    When POST /api/v1/queries with INSERT query
    Then response status is 403
    Evidence: .sisyphus/evidence/task-5-viewer-write-denied.txt
  ```

  **Commit**: YES
  - Message: `test(handlers): add ExecuteQuery handler tests`
  - Files: `internal/api/handlers/query_test.go`
  - Pre-commit: `go test ./internal/api/handlers/...`

- [ ] 6. PreviewWriteQuery handler tests

  **What to do**:
  - Create tests for `PreviewWriteQuery` handler
  - Test DELETE preview returning affected rows
  - Test UPDATE preview returning affected rows
  - Test preview with different user permissions
  - Test preview for non-DELETE/UPDATE queries (should reject)
  - Test preview when no rows match (no_match response)
  - Test error handling for data source connection failures
  - Test permission denied scenarios

  **Must NOT do**:
  - Do not execute actual write operations in tests (mock service)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: Wave 1

  **References**:
  - `internal/api/handlers/query.go:367-412` - PreviewWriteQuery handler
  - `internal/service/query.go:1315-1396` - PreviewWriteQuery service
  - `internal/service/query.go:1400-1509` - PreviewAndValidateWriteQuery

  **Acceptance Criteria**:
  - [ ] Preview returns correct affected rows count
  - [ ] Preview works for DELETE and UPDATE only
  - [ ] Permission checks are verified

  **QA Scenarios**:
  ```gherkin
  Scenario: DELETE preview shows affected rows
    Given user with write permission
    And datasource has table "users" with 100 rows
    When POST /api/v1/queries/preview with DELETE query
    Then response contains "affected_rows": 100
    And response contains "preview_rows" with sample data
    Evidence: .sisyphus/evidence/task-6-delete-preview.txt

  Scenario: UPDATE preview with WHERE clause
    Given user with write permission
    And datasource has table "orders" with status column
    When POST /api/v1/queries/preview with "UPDATE orders SET status='shipped' WHERE id=1"
    Then response contains "affected_rows": 1
    Evidence: .sisyphus/evidence/task-6-update-preview.txt

  Scenario: Preview with no matching rows
    Given user with write permission
    When POST /api/v1/queries/preview with "DELETE FROM users WHERE id=999999"
    Then response status is 200
    And response contains "status": "no_match"
    And response contains "message": "0 rows"
    Evidence: .sisyphus/evidence/task-6-no-match.txt

  Scenario: Preview rejects SELECT query
    Given admin user
    When POST /api/v1/queries/preview with SELECT query
    Then response status is 400
    Evidence: .sisyphus/evidence/task-6-select-rejected.txt
  ```

  **Commit**: YES
  - Message: `test(handlers): add PreviewWriteQuery handler tests`

- [ ] 7. PreviewInsertQuery handler tests

  **What to do**:
  - Create tests for `PreviewInsertQuery` handler
  - Test INSERT...VALUES preview parsing
  - Test INSERT...SELECT preview execution
  - Test preview with column list
  - Test preview without column list
  - Test multi-row INSERT preview (limited to 50 rows)
  - Test permission denied scenarios
  - Test error handling for malformed INSERT statements

  **Must NOT do**:
  - Do not execute actual INSERT (mock preview)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: Wave 1

  **References**:
  - `internal/api/handlers/query.go:413-461` - PreviewInsertQuery handler
  - `internal/service/query.go:1130-1166` - PreviewInsertQuery service
  - `internal/service/query.go:1169-1224` - previewInsertValues
  - `internal/service/query.go:1227-1310` - previewInsertSelect

  **Acceptance Criteria**:
  - [ ] INSERT...VALUES preview returns parsed rows
  - [ ] INSERT...SELECT preview executes SELECT with LIMIT
  - [ ] Column names are correctly extracted

  **QA Scenarios**:
  ```gherkin
  Scenario: INSERT VALUES preview shows rows
    Given user with write permission
    When POST /api/v1/queries/preview-insert with "INSERT INTO users (name, email) VALUES ('Alice', 'alice@test.com')"
    Then response contains columns ["name", "email"]
    And response contains rows with parsed values
    Evidence: .sisyphus/evidence/task-7-insert-values.txt

  Scenario: INSERT SELECT preview limited
    Given user with write permission
    When POST /api/v1/queries/preview-insert with "INSERT INTO archive SELECT * FROM logs"
    Then response contains "total_row_count" > 0
    And "preview_rows" contains at most 50 rows
    Evidence: .sisyphus/evidence/task-7-insert-select.txt

  Scenario: Multi-row INSERT limited preview
    Given INSERT with 100 rows
    When preview is requested
    Then only first 50 rows are shown in preview
    Evidence: .sisyphus/evidence/task-7-multi-row-limit.txt
  ```

  **Commit**: YES
  - Message: `test(handlers): add PreviewInsertQuery handler tests`

- [ ] 8. ListQueries and QueryHistory tests

  **What to do**:
  - Test `ListQueries` handler with pagination
  - Test `ListQueries` filtering by user (regular user sees only own queries)
  - Test `ListQueries` admin sees all queries
  - Test `ListQueryHistory` with pagination
  - Test `ListQueryHistory` with search filters
  - Test `ListQueryHistory` filtering by data source
  - Test permission enforcement (user can only see own history)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2

  **References**:
  - `internal/api/handlers/query.go:300-338` - ListQueries handler
  - `internal/api/handlers/query.go:563-630` - ListQueryHistory handler

  **Acceptance Criteria**:
  - [ ] Pagination works correctly
  - [ ] Regular users only see their own data
  - [ ] Admins see all data

  **Commit**: YES
  - Message: `test(handlers): add ListQueries and QueryHistory tests`

- [ ] 9. ExplainQuery and DryRunDelete tests

  **What to do**:
  - Test `ExplainQuery` handler for EXPLAIN output
  - Test `ExplainQuery` with EXPLAIN ANALYZE option
  - Test `DryRunDelete` handler for DELETE conversion
  - Test permission enforcement
  - Test data source connection error handling

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2

  **References**:
  - `internal/api/handlers/query.go:631-668` - ExplainQuery handler
  - `internal/api/handlers/query.go:669-705` - DryRunDelete handler
  - `internal/service/query.go:619-683` - ExplainQuery service
  - `internal/service/query.go:686-751` - DryRunDelete service

  **Acceptance Criteria**:
  - [ ] EXPLAIN returns query plan
  - [ ] DryRunDelete returns affected rows without modifying data

  **Commit**: YES
  - Message: `test(handlers): add ExplainQuery and DryRunDelete tests`

- [ ] 10. ExportQuery and GetQueryResults tests

  **What to do**:
  - Test `ExportQuery` for CSV export
  - Test `ExportQuery` for JSON export
  - Test `GetQueryResults` with pagination
  - Test `GetQueryResults` with sorting
  - Test large result set handling
  - Test permission enforcement

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2

  **References**:
  - `internal/api/handlers/query.go:707-801` - GetQueryResults handler
  - `internal/api/handlers/query.go:802-851` - ExportQuery handler

  **Acceptance Criteria**:
  - [ ] CSV export has correct format
  - [ ] JSON export has correct structure
  - [ ] Pagination works correctly

  **Commit**: YES
  - Message: `test(handlers): add ExportQuery and GetQueryResults tests`

- [ ] 11. Permission check helper tests

  **What to do**:
  - Test `checkReadPermission` for all user roles
  - Test `checkWritePermission` for all user roles
  - Test permission resolution logic with multiple groups
  - Test admin bypass
  - Test viewer restriction
  - Test group permission inheritance

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2

  **References**:
  - `internal/api/handlers/query.go:527-560` - Permission check helpers
  - `internal/service/query.go:64-106` - GetEffectivePermissions

  **Acceptance Criteria**:
  - [ ] Admin has full access
  - [ ] Viewer has read-only access
  - [ ] Group permissions are correctly unioned

  **Commit**: YES
  - Message: `test(handlers): add permission check tests`

---

### Wave 3: Approval Workflow Tests

- [ ] 12. CreateApprovalRequest handler tests

  **What to do**:
  - Create `internal/api/handlers/approval_test.go`
  - Test approval request creation for write queries
  - Test approval request with preview data
  - Test duplicate approval request handling
  - Test permission enforcement (who can create)
  - Test validation errors
  - Test `RequiresApproval` logic for all operation types

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3

  **References**:
  - `internal/api/handlers/approval.go` - Approval handlers
  - `internal/service/approval.go` - Approval service
  - `internal/service/parser.go:79-103` - RequiresApproval function

  **Acceptance Criteria**:
  - [ ] Approval request created for write queries
  - [ ] Preview data included correctly
  - [ ] Permission enforced

  **Commit**: YES
  - Message: `test(handlers): add CreateApprovalRequest tests`

- [ ] 13. ReviewApproval handler tests

  **What to do**:
  - Test approval with valid approver
  - Test rejection with reason
  - Test self-approval prevention
  - Test duplicate review prevention
  - Test review by non-approver (should fail)
  - Test review of non-pending approval (should fail)
  - Test different user roles as approvers

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3

  **References**:
  - `internal/service/approval.go:323-433` - ReviewApproval logic
  - `internal/service/approval_workflow_test.go:54-464` - Existing workflow tests

  **Acceptance Criteria**:
  - [ ] Approval works with eligible approver
  - [ ] Self-approval blocked
  - [ ] Duplicate review blocked

  **Commit**: YES
  - Message: `test(handlers): add ReviewApproval handler tests`

- [ ] 14. GetApprovals handler tests

  **What to do**:
  - Test listing approvals with pagination
  - Test filtering by status (pending, approved, rejected)
  - Test filtering by data source
  - Test admin sees all approvals
  - Test regular user sees only approvals where they are requester or approver
  - Test sorting by creation time

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3

  **References**:
  - `internal/service/approval.go:151-212` - ListApprovals logic

  **Acceptance Criteria**:
  - [ ] Pagination works
  - [ ] Filters work
  - [ ] Role-based visibility enforced

  **Commit**: YES
  - Message: `test(handlers): add GetApprovals tests`

- [ ] 15. Approval eligibility tests

  **What to do**:
  - Test `GetEligibleApprovers` for various permission configurations
  - Test approver must have `can_approve` permission
  - Test requesters cannot be approvers
  - Test multiple eligible approvers
  - Test no eligible approvers scenario

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Complex permission resolution logic
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3

  **References**:
  - `internal/service/approval.go:434-488` - GetEligibleApprovers logic

  **Acceptance Criteria**:
  - [ ] Correctly identifies approvers
  - [ ] Excludes requesters
  - [ ] Handles no-approver scenario

  **Commit**: YES
  - Message: `test(service): add approval eligibility tests`

- [ ] 16. Self-approval prevention tests

  **What to do**:
  - Test user cannot approve own request
  - Test even if user has approval permission
  - Test applies to all operation types
  - Test error message is clear

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3

  **References**:
  - `internal/service/approval_workflow_test.go:54-88` - Existing tests

  **Acceptance Criteria**:
  - [ ] Self-approval blocked for all users
  - [ ] Clear error message

  **Commit**: YES
  - Message: `test(service): add self-approval prevention tests`

- [ ] 17. Duplicate review prevention tests

  **What to do**:
  - Test user cannot review same approval twice
  - Test works for both approve and reject
  - Test applies to all users including admins

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3

  **References**:
  - `internal/service/approval_workflow_test.go:289-331` - Existing tests

  **Acceptance Criteria**:
  - [ ] Duplicate review blocked
  - [ ] Error message clear

  **Commit**: YES
  - Message: `test(service): add duplicate review prevention tests`

---

### Wave 4: Multi-Query Transaction Tests

- [ ] 18. PreviewMultiQuery handler tests

  **What to do**:
  - Create `internal/api/handlers/multi_query_test.go`
  - Test preview for multiple statements
  - Test preview with SET variables
  - Test preview with different operation types
  - Test permission enforcement for each statement
  - Test error handling for invalid SQL
  - Test transaction control statement blocking (BEGIN, COMMIT, ROLLBACK)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 19-22 in Wave 4)
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: Wave 1

  **References**:
  - `internal/api/handlers/multi_query.go:33-121` - PreviewMultiQuery handler
  - `internal/service/multi_query_service.go:69-164` - PreviewMultiQuery service
  - `internal/service/multi_query_parser_test.go:10-468` - Parser tests

  **Acceptance Criteria**:
  - [ ] All statements previewed correctly
  - [ ] Permission checked per statement
  - [ ] Transaction control blocked

  **Commit**: YES
  - Message: `test(handlers): add PreviewMultiQuery handler tests`

- [ ] 19. ExecuteMultiQuery handler tests

  **What to do**:
  - Test execution of multiple statements
  - Test transaction creation
  - Test permission enforcement
  - Test requires approval for write operations
  - Test error handling and rollback
  - Test statement ordering

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4

  **References**:
  - `internal/api/handlers/multi_query.go:122-208` - ExecuteMultiQuery handler
  - `internal/service/multi_query_service.go:256-XXX` - ExecuteMultiQuery service

  **Acceptance Criteria**:
  - [ ] Statements execute in order
  - [ ] Transaction created correctly
  - [ ] Approval required for writes

  **Commit**: YES
  - Message: `test(handlers): add ExecuteMultiQuery handler tests`

- [ ] 20. CommitMultiQuery tests

  **What to do**:
  - Test commit after successful execution
  - Test commit clears active transaction
  - Test error when no active transaction
  - Test transaction status update

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4

  **References**:
  - `internal/service/query.go:2036-2054` - CommitTransaction
  - `internal/service/multi_query_workflow_test.go:119-154` - Existing tests

  **Acceptance Criteria**:
  - [ ] Commit works after execution
  - [ ] No active transaction after commit

  **Commit**: YES
  - Message: `test(handlers): add CommitMultiQuery tests`

- [ ] 21. RollbackMultiQuery tests

  **What to do**:
  - Test rollback after failure
  - Test rollback with active transaction
  - Testrollback without active transaction
  - Test cleanup of transaction records

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4

  **References**:
  - `internal/service/query.go:2057-2075` - RollbackTransaction
  - `internal/service/multi_query_workflow_test.go:177-229` - Existing tests

  **Acceptance Criteria**:
  - [ ] Rollback works correctly
  - [ ] Transaction cleaned up

  **Commit**: YES
  - Message: `test(handlers): add RollbackMultiQuery tests`

- [ ] 22. Multi-query parser integration tests

  **What to do**:
  - Test `ParseMultipleQueries` with various SQL patterns
  - Test `ValidateMultiQuery` for transaction control blocking
  - Test `IsMultiQuery` detection
  - Test statement separation with semicolons in strings
  - Test comment handling in multi-query
  - Integration with handler flow

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4

  **References**:
  - `internal/service/multi_query_parser.go` - Parser implementation
  - `internal/service/multi_query_parser_test.go` - Existing tests

  **Acceptance Criteria**:
  - [ ] All SQL patterns parsed correctly
  - [ ] Transaction control blocked
  - [ ] Comments handled

  **Commit**: YES
  - Message: `test(service): add multi-query parser integration tests`

---

### Wave 5: RBAC Enforcement Tests

- [ ] 23. Admin role tests (full access)

  **What to do**:
  - Create `internal/service/permission_test.go`
  - Test admin bypasses all permission checks
  - Test admin can execute any query type
  - Test admin can approve any request
  - Test admin can access any data source
  - Test admin can view all query history
  - Test admin can manage all users and groups

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 24-29 in Wave 5)
  - **Parallel Group**: Wave 5
  - **Blocks**: None
  - **Blocked By**: Wave 1, Wave 2, Wave 3

  **References**:
  - `internal/service/query.go:72-76` - Admin bypass logic
  - `internal/models/user.go:14-17` - Role definitions

  **Acceptance Criteria**:
  - [ ] Admin has full permissions
  - [ ] All operations allowed for admin

  **Commit**: YES
  - Message: `test(permission): add admin role tests`

- [ ] 24. User role with read permissions tests

  **What to do**:
  - Test user in group with `can_read` only
  - Test can execute SELECT queries
  - Test cannot execute INSERT/UPDATE/DELETE
  - Test can view own query history
  - Test cannot approve write requests

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5

  **References**:
  - `internal/service/query.go:92-102` - Permission resolution

  **Acceptance Criteria**:
  - [ ] Read-only access enforced
  - [ ] Write operations blocked

  **Commit**: YES
  - Message: `test(permission): add read-only user tests`

- [ ] 25. User role with write permissions tests

  **What to do**:
  - Test user in group with `can_write`
  - Test can execute INSERT/UPDATE/DELETE requests (requires approval)
  - Test can create approval requests
  - Test cannot approve own requests
  - Test can view own requests and history

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5

  **References**:
  - `internal/service/query.go:99-102` - Write permission logic

  **Acceptance Criteria**:
  - [ ] Write access creates approval requests
  - [ ] Cannot self-approve

  **Commit**: YES
  - Message: `test(permission): add write user tests`

- [ ] 26. User role with approve permissions tests

  **What to do**:
  - Test user in group with `can_approve`
  - Test can approve write requests
  - Test cannot approve own requests (even with can_approve)
  - Test approval appears in their approval list
  - Test rejecting requests

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5

  **References**:
  - `internal/service/approval.go:434-488` - Eligible approvers

  **Acceptance Criteria**:
  - [ ] Can approve others' requests
  - [ ] Cannot approve own requests

  **Commit**: YES
  - Message: `test(permission): add approver role tests`

- [ ] 27. Viewer role tests (read-only)

  **What to do**:
  - Test viewer role cannot execute write queries
  - Test viewer can only view queries
  - Test viewer cannot create approval requests
  - Test viewer cannot approve anything
  - Test viewer limited query history access

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5

  **References**:
  - `internal/models/user.go:16` - RoleViewer definition

  **Acceptance Criteria**:
  - [ ] Viewer is read-only
  - [ ] All write operations blocked

  **Commit**: YES
  - Message: `test(permission): add viewer role tests`

- [ ] 28. Group permission inheritance tests

  **What to do**:
  - Test user in multiple groups
  - Test permissions are unioned across groups
  - Test group A has can_read, group B has can_write
  - Test user gets both permissions
  - Test permission precedence when groups overlap

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5

  **References**:
  - `internal/service/query.go:78-103` - Permission union logic

  **Acceptance Criteria**:
  - [ ] Permissions correctly unioned
  - [ ] Multiple groups work

  **Commit**: YES
  - Message: `test(permission): add group inheritance tests`

- [ ] 29. Permission denial scenarios tests

  **What to do**:
  - Test 403 response when user lacks permission
  - Test error message contains permission details
  - Test accessing data source without membership
  - Test executing write query as viewer
  - Test approving request without approval permission

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5

  **References**:
  - `internal/api/handlers/query.go:57-60` - Permission denied response

  **Acceptance Criteria**:
  - [ ] All denials return 403
  - [ ] Error messages clear

  **Commit**: YES
  - Message: `test(permission): add denial scenario tests`

---

### Wave 6: E2E Tests (Playwright)

- [ ] 30. Login and authentication flow E2E

  **What to do**:
  - Create `web/e2e/auth.spec.ts` (update existing)
  - Test login with admin user
  - Test login with regular user
  - Test login with viewer
  - Test invalid credentials
  - Test token persistence
  - Test logout

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`] if available

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 31-35 in Wave 6)
  - **Parallel Group**: Wave 6
  - **Blocks**: Wave FINAL
  - **Blocked By**: Wave 1-5

  **References**:
  - `web/e2e/auth.spec.ts` - Existing auth tests
  - `web/e2e/dashboard.spec.ts` - Existing dashboard tests

  **Acceptance Criteria**:
  - [ ] All roles can login
  - [ ] Invalid login rejected
  - [ ] Logout clears session

  **Commit**: YES
  - Message: `test(e2e): add authentication flow tests`

- [ ] 31. Single query execution E2E

  **What to do**:
  - Create `web/e2e/query-execution.spec.ts`
  - Test SELECT query execution
  - Test query results display
  - Test pagination of results
  - Test export to CSV
  - Test export to JSON
  - Test error handling for invalid SQL
  - Test as different user roles

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6

  **Acceptance Criteria**:
  - [ ] Query executes successfully
  - [ ] Results display correctly
  - [ ] Export works

  **Commit**: YES
  - Message: `test(e2e): add single query execution tests`

- [ ] 32. Approval workflow E2E

  **What to do**:
  - Create `web/e2e/approval-workflow.spec.ts`
  - Test write query creates approval request
  - Test approver sees pending approvals
  - Test approver approves request
  - Test approver rejects request with reason
  - Test query executes after approval
  - Test self-approval blocked
  - Test requester sees approval status

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6

  **Acceptance Criteria**:
  - [ ] Full approval flow works
  - [ ] Self-approval blocked

  **Commit**: YES
  - Message: `test(e2e): add approval workflow tests`

- [ ] 33. Multi-query transaction E2E

  **What to do**:
  - Create `web/e2e/multi-query.spec.ts`
  - Test multi-query preview
  - Test multi-query execution
  - Test commit after success
  - Test rollback on failure
  - Test SET variable handling
  - Test transaction control blocking

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6

  **Acceptance Criteria**:
  - [ ] Multi-query works correctly
  - [ ] Transaction lifecycle works

  **Commit**: YES
  - Message: `test(e2e): add multi-query transaction tests`

- [ ] 34. Permission matrix E2E

  **What to do**:
  - Create `web/e2e/permission-matrix.spec.ts`
  - Test matrix: (role × operation)
    - Admin: all operations allowed
    - User with read: SELECT only
    - User with write: SELECT + write with approval
    - User with approve: can approve write
    - Viewer: SELECT only
  - Test UI shows correct options per role
  - Test disabled buttons for missing permissions
  - Test error messages for denied operations

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6

  **Acceptance Criteria**:
  - [ ] All permission combinations tested
  - [ ] UI reflects permissions correctly

  **Commit**: YES
  - Message: `test(e2e): add permission matrix tests`

- [ ] 35. Error handling E2E

  **What to do**:
  - Create `web/e2e/error-handling.spec.ts`
  - Test network errors
  - Test authentication errors
  - Test authorization errors (403)
  - Test validation errors
  - Test SQL syntax errors
  - Test data source connection errors
  - Test timeout handling
  - Test error messages are user-friendly

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6

  **Acceptance Criteria**:
  - [ ] All error types handled
  - [ ] User-friendly messages

  **Commit**: YES
  - Message: `test(e2e): add error handling tests`

---

## Final Verification Wave

- [ ] F1. **Test coverage verification** — `oracle`

  Run coverage analysis and verify:
  - Handler coverage >80%
  - Service coverage >70%
  - All permission check functions tested
  - All RBAC scenarios covered
  - Output: Coverage report with gaps identified

- [ ] F2. **RBAC enforcement audit** — `deep`

  Audit RBAC implementation:
  - Every API endpoint has permission checks
  - Every write operation requires appropriate permission
  - Admin bypass is consistent
  - No permission check is bypassed
  - Output: RBAC audit report

- [ ] F3. **Integration test run verification** — `unspecified-high`

  Run all integration tests:
  - Tests pass with `go test -v -tags=integration ./...`
  - Tests clean up after themselves
  - No flaky tests
  - Output: Test run results

- [ ] F4. **E2E test run verification** — `unspecified-high`

  Run all E2E tests:
  - Tests pass with `npm run test:e2e`
  - All user flows work
  - All permission scenarios work
  - Output: E2E test results

---

## Commit Strategy

1. Wave 1: `test(fixtures): add test infrastructure`
2. Wave 2: `test(handlers): add query handler tests`
3. Wave 3: `test(approval): add approval workflow tests`
4. Wave 4: `test(multi-query): add multi-query tests`
5. Wave 5: `test(permission): add RBAC enforcement tests`
6. Wave 6: `test(e2e): add E2E tests`
7. Final: `test: verify coverage and cleanup`

---

## Success Criteria

### Verification Commands
```bash
# Backend unit tests
go test -v -cover ./internal/... 

# Backend integration tests
go test -v -tags=integration ./internal/...

# Frontend E2E tests  
npm run test:e2e

# Coverage report
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### Final Checklist
- [ ] All "Must Have" tests present
- [ ] All "Must NOT Have" scenarios absent
- [ ] Handler coverage >80%
- [ ] Service coverage >70%
- [ ] RBAC enforcement verified
- [ ] E2E tests pass in CI
- [ ] No skipped tests for permission scenarios

---

## Appendix: Test Data Matrix

### User Roles Matrix

| Role | Can Read | Can Write | Can Approve | Admin Bypass |
|------|----------|-----------|-------------|---------------|
| Admin | ✓ | ✓ | ✓ | ✓ |
| User (read only) | ✓ | ✗ | ✗ | ✗ |
| User (write) | ✓ | ✓* | ✗ | ✗ |
| User (approve) | ✓ | ✓* | ✓ | ✗ |
| Viewer | ✓ | ✗ | ✗ | ✗ |

*Write operations require approval

### Permission Test Scenarios

| Scenario | User | Group | DataSource | Expected |
|----------|------|-------|------------|----------|
| Admin SELECT | Admin | - | Any | ✓ Allow |
| Admin INSERT | Admin | - | Any | ✓ Allow (no approval) |
| User SELECT | User | GroupA | DS1 (can_read) | ✓ Allow |
| User INSERT | User | GroupA | DS1 (can_read) | ✗ Deny |
| User INSERT | User | GroupA | DS1 (can_write) | ✓ Approval required |
| User APPROVE own | User | GroupA | DS1 (can_approve) | ✗ Cannot approve own |
| User APPROVE other | User | GroupA | DS1 (can_approve) | ✓ Can approve |
| Viewer SELECT | Viewer | GroupA | DS1 (can_read) | ✓ Allow |
| Viewer INSERT | Viewer | GroupA | DS1 (can_write) | ✗ Deny |
| Multi-group User | User | GroupA (read) + GroupB (write) | DS1 | ✓ Read + Write |

### Multi-Query Test Scenarios

| Scenario | Statements | Permission | Expected |
|----------|-----------|-------------|----------|
| All SELECT | SELECT; SELECT | can_read | ✓ Execute directly |
| Mix of read/write | SELECT; INSERT | can_write | ✓ Approval required |
| SET variables | SET @x=1; UPDATE | can_write | ✓ Execute sequentially |
| Transaction control | BEGIN; SELECT | Any | ✗ Block (forbidden) |
| Admin multi-query | INSERT; DELETE | Admin | ✓ Execute directly |