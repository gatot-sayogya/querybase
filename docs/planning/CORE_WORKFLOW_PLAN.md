# Core Workflow Focus Plan
**Polish Current User Experience Before Adding New Features**

**Date:** January 28, 2025
**Status:** Ready for Implementation
**Approach:** Complete core workflow, defer Schema API and Folder System

---

## Executive Summary

**Strategic Decision:** Focus on completing and polishing the **current user workflow** before adding Schema API, Folder System, and other planned features.

**Current User Workflow:**
1. **Login & Authentication** ✅ Complete
2. **Write Query** ✅ Mostly complete
3. **Display Query Results** ✅ Mostly complete
4. **Approval Workflow** ✅ Complete with transaction preview
5. **Notifications** ✅ Complete (Google Chat)
6. **Data Source Management** ✅ Complete
7. **RBAC** ✅ Complete

**New Strategy:**
- ✅ Polish existing features to production-ready quality
- ✅ Fix any gaps in current workflow
- ✅ Add missing UX enhancements
- ✅ Improve error handling and edge cases
- ❌ Defer Schema API, Folder System, Tags to future phases

**Timeline:** 2-3 weeks to production-ready MVP

---

## Current Workflow Analysis

### Workflow 1: Login & Authentication ✅

**Current State:**
```yaml
Features:
  ✅ JWT login endpoint
  ✅ Password hashing (bcrypt)
  ✅ Get current user
  ✅ Change password
  ✅ User CRUD (admin)
  ✅ Group management
  ✅ RBAC middleware

Status: COMPLETE
```

**Potential Improvements:**
- ⚠️ Password reset flow (forgot password)
- ⚠️ Session management (logout all devices)
- ⚠️ Two-factor authentication (future)
- ⚠️ OAuth integration (future)

**Priority:** 🟢 LOW - Current auth is sufficient for MVP

**Decision:** Keep as-is, add password reset in Phase 2

---

### Workflow 2: Write Query ✅

**Current State:**
```yaml
Features:
  ✅ Query execution endpoint
  ✅ SQL parser (operation detection)
  ✅ SQL validation endpoint
  ✅ Automatic approval creation for writes
  ✅ Dry run DELETE
  ✅ EXPLAIN support

Status: MOSTLY COMPLETE
```

**User Journey:**
```
1. User logs in
   └─> ✅ Works

2. User selects data source
   └─> ✅ GET /api/v1/datasources

3. User writes SQL query
   └─> ✅ Can execute via POST /api/v1/queries

4. System validates SQL
   └─> ✅ POST /api/v1/queries/validate

5. System detects operation type
   └─> ✅ Parser detects SELECT vs write

6a. SELECT queries → Execute immediately
    └─> ✅ Works

6b. Write queries → Create approval request
    └─> ✅ Works
```

**Gaps Identified:**

| Gap | Impact | Effort | Priority |
|-----|--------|--------|----------|
| Query templates/samples | Nice to have | 1-2 days | 🟢 LOW |
| Query history (search) | Useful | Already exists | 🟢 LOW |
| Multiple query tabs | Frontend only | Frontend | 🟢 LOW |
| Save query before execute | Already exists | N/A | ✅ COMPLETE |

**Decision:** Current query workflow is complete. No backend work needed.

---

### Workflow 3: Display Query Results ✅

**Current State:**
```yaml
Features:
  ✅ Query execution
  ✅ Result storage (JSONB)
  ✅ Query history pagination
  ✅ Column metadata (names, types)
  ✅ Row count tracking

Status: MOSTLY COMPLETE
```

**User Journey:**
```
1. User executes query
   └─> ✅ POST /api/v1/queries

2. System returns results
   └─> ✅ Returns data immediately

3. Results stored in DB
   └─> ✅ QueryResult table

4. User views query history
   └─> ✅ GET /api/v1/queries/history (paginated)

5. User views saved query
   └─> ✅ GET /api/v1/queries/:id
```

**Gaps Identified:**

| Gap | Impact | Effort | Priority |
|-----|--------|--------|----------|
| **Result pagination** | Large result sets | 2-3 days | 🟡 MEDIUM |
| **Query result export** | Download data | 1-2 days | 🟡 MEDIUM |
| **Query result caching** | Performance | Already done | ✅ COMPLETE |
| **Display different data types** | Frontend work | Frontend | 🟢 LOW |

**Analysis:**

**Pagination Gap:**
- Current: Returns entire result set (could be 10,000+ rows)
- Problem: Frontend must handle all data at once
- Impact: Slow rendering, memory issues
- **Solution:** Add pagination endpoint (from gap analysis)
- **Effort:** 2-3 days
- **Priority:** 🟡 MEDIUM - Not blocking, but should fix

**Export Gap:**
- Current: Can view results in browser only
- Problem: Can't download for analysis
- Impact: Minor inconvenience
- **Solution:** Add CSV/JSON export endpoint
- **Effort:** 1-2 days
- **Priority:** 🟡 MEDIUM - Nice to have

**Decision:** Add pagination and export to current workflow improvement

---

### Workflow 4: Approval Workflow ✅

**Current State:**
```yaml
Features:
  ✅ Approval request creation
  ✅ Approval request listing (with filters)
  ✅ Review (approve/reject) endpoint
  ✅ Transaction-based preview
  ✅ Dry run DELETE
  ✅ Commit/rollback transaction
  ✅ Eligible approvers endpoint
  ✅ Google Chat notifications

Status: COMPLETE - Flagship feature
```

**User Journey:**
```
1. User submits write query
   └─> ✅ Creates approval request

2. Approvers view pending approvals
   └─> ✅ GET /api/v1/approvals

3. Approver reviews request
   └─> ✅ GET /api/v1/approvals/:id

4. Approver starts transaction
   └─> ✅ POST /api/v1/approvals/:id/transaction-start

5. System executes in transaction mode
   └─> ✅ Shows results (preview)

6. Approver commits or rolls back
   └─> ✅ POST /api/v1/transactions/:id/commit
   └─> ✅ POST /api/v1/transactions/:id/rollback

7. Google Chat notification sent
   └─> ✅ Webhook integration
```

**Gaps Identified:**

| Gap | Impact | Effort | Priority |
|-----|--------|--------|----------|
| **Approval comments** | Collaboration | 2-3 days | 🟡 MEDIUM |
| **Bulk approve/reject** | Convenience | 1-2 days | 🟢 LOW |
| **Approval history view** | Audit trail | Already exists | ✅ COMPLETE |

**Comments Gap Analysis:**

**Current:**
- Approvers can approve/reject
- No way to discuss WHY they approved/rejected
- No communication thread on approval

**Proposed:**
- Add comments to approval requests
- Allow discussion before decision
- Document reasoning for audit trail

**Is it blocking?**
- ❌ NO - System works without comments
- ⚠️ YES - Useful for team collaboration
- 🟡 MEDIUM - Should add, but not urgent

**Decision:** Add comments in polish phase (Week 2-3)

---

### Workflow 5: Notifications ✅

**Current State:**
```yaml
Features:
  ✅ Google Chat webhook integration
  ✅ Notification on approval request created
  ✅ Notification on approval decision
  ✅ Notification on query executed
  ✅ Notification configs per data source

Status: COMPLETE
```

**Gaps Identified:**

| Gap | Impact | Effort | Priority |
|-----|--------|--------|----------|
| Email notifications | Nice to have | 2-3 days | 🟢 LOW |
| Slack integration | Nice to have | 1-2 days | 🟢 LOW |
| In-app notifications | Frontend work | Frontend | 🟢 LOW |
| Notification preferences | Nice to have | 1-2 days | 🟢 LOW |

**Decision:** Current Google Chat notifications are sufficient for MVP

---

### Workflow 6: Data Source Management ✅

**Current State:**
```yaml
Features:
  ✅ List data sources
  ✅ Create data source (admin)
  ✅ Update data source (admin)
  ✅ Delete data source (admin)
  ✅ Test connection
  ✅ Get/set permissions
  ✅ Get eligible approvers
  ✅ Encrypted password storage (AES-256-GCM)
  ✅ Support for PostgreSQL and MySQL

Status: COMPLETE
```

**User Journey:**
```
1. Admin adds new data source
   └─> ✅ POST /api/v1/datasources

2. Admin tests connection
   └─> ✅ POST /api/v1/datasources/:id/test

3. Admin sets permissions
   └─> ✅ PUT /api/v1/datasources/:id/permissions

4. Users query data source
   └─> ✅ Works
```

**Gaps Identified:**

| Gap | Impact | Effort | Priority |
|-----|--------|--------|----------|
| **Health check status** | Monitoring | 1 day | 🟢 LOW |
| **Usage statistics** | Analytics | 1-2 days | 🟢 LOW |
| **Connection pooling stats** | Performance | 1-2 days | 🟢 LOW |
| Schema browser | ✨ DEFERRED | 3-4 days | 🔴 Future |

**Health Check Gap:**

**Current:**
- Can test connection on-demand
- No way to see if data source is currently healthy
- No monitoring dashboard

**Proposed:**
- Add health check endpoint
- Show status in data source list
- Track connection latency

**Is it blocking?**
- ❌ NO - Manual test works
- ⚠️ YES - Nice for monitoring
- 🟢 LOW - Quality of life improvement

**Decision:** Add health check in polish phase (Week 2-3)

---

### Workflow 7: RBAC ✅

**Current State:**
```yaml
Features:
  ✅ JWT authentication
  ✅ Role-based access control (admin, user, viewer)
  ✅ Group-based permissions
  ✅ User-group assignment
  ✅ Data source permissions (can_read, can_write, can_approve)
  ✅ user_permissions view
  ✅ RBAC middleware

Status: COMPLETE - Enterprise-grade
```

**User Journey:**
```
1. User logs in
   └─> ✅ JWT token issued

2. User accesses endpoint
   └─> ✅ Auth middleware validates token

3. User attempts write operation
   └─> ✅ RBAC checks permissions

4. User tries to approve
   └─> ✅ RBAC checks can_approve permission

5. Admin manages permissions
   └─> ✅ PUT /api/v1/datasources/:id/permissions
```

**Gaps Identified:**

| Gap | Impact | Effort | Priority |
|-----|--------|--------|----------|
| Fine-grained permissions | Nice to have | 3-5 days | 🟢 LOW |
| Permission inheritance | Nice to have | 2-3 days | 🟢 LOW |
| Audit log for permissions | Compliance | Already exists | ✅ COMPLETE |

**Decision:** Current RBAC is sufficient for MVP. No changes needed.

---

## Current Workflow Completion Plan

### Phase 1: Critical Improvements (Week 1) 🟡

**Goal:** Fix gaps that affect core workflow usability

#### Task 1: Query Results Pagination (2-3 days)

**Problem:** Large result sets (10,000+ rows) cause performance issues

**Solution:**
```go
// Add to internal/service/query.go
func (s *QueryService) GetPaginatedResults(
    queryID uuid.UUID,
    page, limit int,
    sortColumn, sortDirection string,
) (*PaginatedResultDTO, error)

// Add to internal/api/handlers/query.go
func (h *Handler) GetQueryResults(c *gin.Context)
```

**API Endpoint:**
```
GET /api/v1/queries/:id/results?page=1&limit=100
```

**Files to Update:**
- `internal/service/query.go` - Add pagination logic
- `internal/api/handlers/query.go` - Add pagination endpoint
- `internal/api/dto/query.go` - Add PaginatedResultDTO

**Testing:**
- Unit tests for pagination logic
- Integration test with 10,000 row result set
- Verify sort functionality

---

#### Task 2: Query Result Export (1-2 days)

**Problem:** Users can't download query results

**Solution:**
```go
// Create internal/service/export.go
func (s *ExportService) ExportQueryResults(queryID uuid.UUID, format string) ([]byte, string, error)

// Create internal/api/handlers/export.go
func (h *Handler) ExportQueryResults(c *gin.Context)
```

**API Endpoint:**
```
GET /api/v1/queries/:id/export?format=csv
GET /api/v1/queries/:id/export?format=json
```

**Files to Create:**
- `internal/service/export.go` - Export service
- `internal/api/handlers/export.go` - Export handler
- `internal/api/dto/export.go` - Export DTOs

**Testing:**
- Test CSV export with special characters
- Test JSON export with nested data
- Verify Content-Disposition headers

---

### Phase 2: Collaboration Features (Week 2) 🟡

**Goal:** Improve team collaboration in approval workflow

#### Task 3: Approval Comments (2-3 days)

**Problem:** No way to discuss approval requests

**Solution:**
```sql
-- Migration 000006
CREATE TABLE approval_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Files to Create:**
- `internal/models/approval.go` - Add ApprovalComment model
- `internal/service/comment.go` - Comment service
- `internal/api/handlers/approval.go` - Add comment endpoints
- `internal/api/dto/approval.go` - Add comment DTOs
- `migrations/000006_add_approval_comments.up.sql` - Migration

**API Endpoints:**
```
POST   /api/v1/approvals/:id/comments
GET    /api/v1/approvals/:id/comments
PUT    /api/v1/approvals/:id/comments/:commentId
DELETE /api/v1/approvals/:id/comments/:commentId
```

**Testing:**
- Test comment creation
- Test comment retrieval (pagination)
- Test comment updates (only by owner or admin)
- Test cascade delete on approval deletion

---

#### Task 4: Data Source Health Check (1 day)

**Problem:** No visibility into data source health

**Solution:**
```go
// Add to internal/service/datasource.go
func (s *DataSourceService) GetHealthStatus(dataSourceID string) (*HealthStatusDTO, error)
```

**API Endpoint:**
```
GET /api/v1/datasources/:id/health
```

**Response:**
```json
{
  "status": "healthy",
  "latency": "12ms",
  "lastCheck": "2024-01-15T10:30:00Z",
  "version": "PostgreSQL 15.2",
  "connections": 5
}
```

**Files to Update:**
- `internal/service/datasource.go` - Add health check method
- `internal/api/handlers/datasource.go` - Add health endpoint
- `internal/api/dto/datasource.go` - Add HealthStatusDTO

**Testing:**
- Test with healthy data source
- Test with unreachable data source
- Test with slow data source

---

### Phase 3: Polish & Quality (Week 3) 🟢

**Goal:** Production-ready quality

#### Task 5: Error Handling Improvements (2-3 days)

**Current State:** Basic error handling

**Improvements Needed:**
- Consistent error response format
- Better error messages
- HTTP status code correctness
- Request ID for tracing

**Solution:**
```go
// Create internal/api/middleware/errors.go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details string `json:"details,omitempty"`
    RequestID string `json:"requestId,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Handle errors
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            // Log error
            // Return consistent error response
        }
    }
}
```

**Files to Create:**
- `internal/api/middleware/errors.go` - Error handling middleware
- `internal/api/errors/` - Error types

---

#### Task 6: Request Validation Improvements (1-2 days)

**Current State:** Basic validation

**Improvements Needed:**
- Validate all request DTOs
- Custom validation error messages
- Input sanitization

**Solution:**
```go
// Use github.com/go-playground/validator/v10
type CreateQueryRequest struct {
    DataSourceID string `json:"dataSourceId" binding:"required,uuid"`
    QueryText    string `json:"queryText" binding:"required,min=1"`
    Save         bool   `json:"save"`
}

// Custom validation messages
func registerCustomValidations(v *validator.Validate) {
    v.RegisterValidation("sql", validateSQL)
}
```

---

#### Task 7: Documentation & Testing (2-3 days)

**Tasks:**
- Update API documentation
- Add integration tests for all workflows
- Performance testing
- Security audit

---

## Deferred Features (Future Phases)

### Deferred to Phase 4 (After MVP): 🔴

**Schema & Organization Features:**
- ❌ Schema Introspection API (3-4 days)
  - Autocomplete (frontend can use text search for now)
  - Schema browser (frontend can show simple list for now)
- ❌ Folder System (3-4 days)
  - Use flat query list with search/filter
  - Add folders when query volume grows
- ❌ Tag System (2-3 days)
  - Use folders and search for now
  - Add tags for advanced organization later

**Real-time Features:**
- ❌ WebSocket Support (4-5 days)
  - Use polling (3-5 second intervals)
  - Add WebSocket if performance metrics show need

**Polish Features:**
- ❌ SQL Formatting Endpoint (1-2 days)
  - Frontend can use client-side formatter
- ❌ Query Comparison API (2-3 days)
  - Nice-to-have, not critical
- ❌ Bulk Operations (1-2 days)
  - Low priority, few approvals will be batched

**Rationale for Deferring:**

1. **Current workflow is functional**
   - Users can login, query, approve, manage data sources
   - RBAC works correctly
   - No blocking gaps in core experience

2. **Frontend can work around missing features**
   - Autocomplete: Text search instead of schema-aware
   - Organization: Search/filter instead of folders
   - Real-time: Polling instead of WebSocket

3. **Focus on quality over quantity**
   - Better to have 7 polished features than 15 half-baked features
   - Production-ready MVP beats buggy feature-rich system

4. **User feedback should guide next steps**
   - Deploy current workflow
   - Gather user feedback
   - Prioritize based on actual usage patterns

---

## Implementation Timeline

### Week 1: Critical Improvements

**Day 1-3: Query Pagination**
- Implement pagination in query service
- Add pagination endpoint
- Write tests
- Update documentation

**Day 4-5: Query Export**
- Implement export service
- Add export endpoint (CSV, JSON)
- Write tests
- Update documentation

**Deliverables:**
- ✅ Paginated query results
- ✅ Export functionality (CSV/JSON)

---

### Week 2: Collaboration & Monitoring

**Day 1-3: Approval Comments**
- Create migration
- Implement comment service and handlers
- Write tests
- Update documentation

**Day 4-5: Health Checks & Testing**
- Implement health check endpoint
- Integration tests
- Performance testing

**Deliverables:**
- ✅ Approval comments feature
- ✅ Data source health monitoring

---

### Week 3: Polish & Production Ready

**Day 1-2: Error Handling**
- Consistent error responses
- Request tracing
- Better error messages

**Day 3-4: Validation & Security**
- Request validation improvements
- Security audit
- Documentation

**Day 5: Performance Testing**
- Load testing
- Optimization
- Production deployment prep

**Deliverables:**
- ✅ Production-ready backend
- ✅ Comprehensive documentation
- ✅ Performance benchmarks

---

## Success Criteria

### Core Workflow Completeness ✅

**Authentication:**
- ✅ Users can login
- ✅ Users can change password
- ✅ Admins can manage users
- ✅ RBAC works correctly

**Query Execution:**
- ✅ Users can execute SELECT queries
- ✅ Users can submit write queries
- ✅ Results display correctly
- ✅ Large result sets are paginated
- ✅ Results can be exported

**Approval Workflow:**
- ✅ Write queries create approval requests
- ✅ Approvers can review requests
- ✅ Transaction preview works
- ✅ Approvers can discuss (comments)
- ✅ Notifications sent via Google Chat

**Data Source Management:**
- ✅ Admins can add data sources
- ✅ Connections can be tested
- ✅ Permissions can be configured
- ✅ Health status is visible

**Quality:**
- ✅ Error handling is consistent
- ✅ Error messages are clear
- ✅ Performance is acceptable (< 500ms for queries)
- ✅ Security best practices followed

---

## Files to Create/Update

### Week 1

**Update (2 files):**
- `internal/service/query.go` - Add pagination
- `internal/api/handlers/query.go` - Add pagination endpoint
- `internal/api/dto/query.go` - Add pagination DTOs

**Create (3 files):**
- `internal/service/export.go` - Export service
- `internal/api/handlers/export.go` - Export handler
- `internal/api/dto/export.go` - Export DTOs

### Week 2

**Create (6 files):**
- `internal/models/approval.go` - Add ApprovalComment model
- `internal/service/comment.go` - Comment service
- `internal/api/dto/approval.go` - Add comment DTOs
- `migrations/000006_add_approval_comments.up.sql` - Migration
- `migrations/000006_add_approval_comments.down.sql` - Rollback
- Update `internal/api/handlers/datasource.go` - Add health check

### Week 3

**Create (2 files):**
- `internal/api/middleware/errors.go` - Error handling
- `internal/api/errors/` - Error types

**Update (all handlers):**
- Improve error handling
- Add request tracing

**Create (test files):**
- Integration tests
- Performance tests

---

## Summary

### Current Workflow Status

| Workflow | Status | Gaps | Priority |
|----------|--------|------|----------|
| Login & Auth | ✅ COMPLETE | Minor | 🟢 LOW |
| Write Query | ✅ COMPLETE | None | ✅ DONE |
| Display Results | ⚠️ MOSTLY COMPLETE | Pagination, Export | 🟡 MEDIUM |
| Approval | ✅ COMPLETE | Comments | 🟡 MEDIUM |
| Notifications | ✅ COMPLETE | None | ✅ DONE |
| Data Source Mgmt | ✅ COMPLETE | Health check | 🟢 LOW |
| RBAC | ✅ COMPLETE | None | ✅ DONE |

### Implementation Focus

**Week 1:**
- ✅ Query pagination (2-3 days)
- ✅ Query export (1-2 days)

**Week 2:**
- ✅ Approval comments (2-3 days)
- ✅ Health check endpoint (1 day)

**Week 3:**
- ✅ Error handling polish (2-3 days)
- ✅ Testing & documentation (2-3 days)

**Total:** 3 weeks to production-ready MVP

### Deferred Features (Future Phases)

**Phase 4+ (After MVP Launch):**
- Schema Introspection API
- Folder System
- Tag System
- WebSocket Support
- SQL Formatter
- Query Comparison
- Advanced analytics

**Rationale:**
- Current workflow is functional
- Focus on quality over quantity
- User feedback should guide priorities
- Production-ready MVP beats feature-rich buggy system

---

**Last Updated:** January 28, 2025
**Status:** Ready for Implementation
**Next Step:** Start query pagination implementation (Week 1, Day 1)
