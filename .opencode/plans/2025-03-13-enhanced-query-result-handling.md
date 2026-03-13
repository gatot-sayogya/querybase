# Enhanced Query Result Handling Design

**Date:** 2025-03-13  
**Status:** Approved  
**Author:** AI Assistant  
**Scope:** QueryBase - Query execution results, empty states, and early validation

---

## 1. Problem Statement

### Current Issues

1. **SELECT with no results:** Shows generic "No results found" message without any table structure or context
2. **Write queries on non-existent rows:** Creates approval requests even when UPDATE/DELETE affects 0 rows, wasting approver time
3. **Inconsistent visual states:** No clear visual distinction between running, completed-empty, completed-with-data, and failed states

### Impact
- Users can't troubleshoot empty SELECT results effectively
- Approvers review queries that won't actually change any data
- Poor UX - users don't understand what's happening with their queries

---

## 2. Proposed Solution

### 2.1 SELECT Queries with Empty Results

**User Experience:**
- Query executes successfully but returns 0 rows
- User sees table structure with column headers intact
- Empty state message appears inside the table area
- Provides helpful troubleshooting tips

**Visual Design:**
```
┌─────────────────────────────────────────────────┐
│ ✅ Query completed in 45ms                      │
│ Rows returned: 0                                │
├─────────────────────────────────────────────────┤
│ id │ name │ email │ created_at │ status        │  ← Headers visible
├────┼──────┼───────┼────────────┼───────────────┤
│                                                 │
│   📭 No rows match your query conditions.      │
│                                                 │
│   💡 Tip: Check your WHERE clause values or    │
│      try broadening your search criteria.      │
│                                                 │
└─────────────────────────────────────────────────┘
```

**Technical Changes:**
- Backend: Ensure column names are always returned in `ExecuteQueryResponse` even when `data` array is empty
- Frontend: Modify `QueryResults` component to handle `data.length === 0` while still showing column headers

---

### 2.2 Early Detection for Write Queries with 0 Rows

**User Experience:**
- User submits UPDATE/DELETE query
- System runs preview check BEFORE creating approval
- If affected_rows === 0, show info dialog immediately
- No approval request created, user can modify query

**Flow:**
```
User submits UPDATE/DELETE
         ↓
Backend runs preview (read-only)
         ↓
    ┌────┴────┐
    ↓         ↓
Rows > 0   Rows == 0
    ↓         ↓
Create     Return warning
Approval   Show info modal
Request    No approval created
```

**Visual Design:**
```
┌─────────────────────────────────────────────────┐
│  ℹ️  Query Would Affect 0 Rows                  │
├─────────────────────────────────────────────────┤
│                                                  │
│   Your UPDATE query matches no existing rows.  │
│                                                  │
│   Query: UPDATE users SET status = 'active'   │
│   WHERE id = 99999                              │
│                                                  │
│   💡 The row with id = 99999 doesn't exist.   │
│                                                 │
│   [Edit Query] [Cancel]                         │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Technical Changes:**
- Backend: Add `PreviewAndValidateWriteQuery()` method in QueryService
- Backend: Modify approval creation flow to check affected rows first
- Backend: Add new response type: `ValidationResult` with status `no_match`
- Frontend: Handle new response status in QueryExecutor
- Frontend: Show informative modal instead of redirecting to approvals

---

### 2.3 Distinct Visual States

Each query state should have unique visual styling:

#### State 1: Query Running
**Visual:**
- Animated spinner/progress indicator
- Blue theme color
- "Executing query..." message
- Live duration counter

**Location:** `QueryExecutor.tsx` main execution flow

#### State 2: Query Completed - With Data
**Visual:**
- Green checkmark icon
- "Query completed in Xms" message
- Full data table with results
- Row count displayed

**Location:** `QueryResults.tsx` with `results.data.length > 0`

#### State 3: Query Completed - Empty Results
**Visual:**
- Green checkmark icon (query succeeded!)
- "Query completed in Xms, 0 rows returned"
- Table structure with headers visible
- Empty state message inside table area
- Light gray background for empty area

**Location:** `QueryResults.tsx` with `results.data.length === 0`

#### State 4: Query Failed
**Visual:**
- Red error icon
- "Query failed" header
- Error message in red
- Stack trace (if available, collapsible)
- Retry button

**Location:** `QueryExecutor.tsx` error handling

#### State 5: Write Query - No Rows Match (NEW)
**Visual:**
- Blue info icon
- "No rows match" header
- Query preview
- Explanation of why
- Action buttons: Edit Query / Cancel

**Location:** `QueryExecutor.tsx` new handler for `no_match` status

---

## 3. API Changes

### New DTO: ValidationResult
```go
type ValidationResult struct {
    Valid           bool                   `json:"valid"`
    Status          string                 `json:"status"` // "ok", "no_match", "error"
    Message         string                 `json:"message"`
    AffectedRows    int                    `json:"affected_rows"`
    PreviewRows     []map[string]interface{} `json:"preview_rows,omitempty"`
    Columns         []string               `json:"columns,omitempty"`
    Suggestion      string                 `json:"suggestion,omitempty"`
}
```

### Modified: ExecuteQueryResponse
```go
type ExecuteQueryResponse struct {
    QueryID          string                   `json:"query_id"`
    Status           string                   `json:"status"` // "completed", "failed", "no_match"
    RowCount         *int                     `json:"row_count"`
    ExecutionTime    *int                     `json:"execution_time_ms"`
    Data             []map[string]interface{} `json:"data,omitempty"`
    Columns          []ColumnInfo             `json:"columns"`
    ErrorMessage     string                   `json:"error_message,omitempty"`
    RequiresApproval bool                     `json:"requires_approval"`
    ApprovalID       string                   `json:"approval_id,omitempty"`
    Validation       *ValidationResult        `json:"validation,omitempty"` // NEW
}
```

---

## 4. Frontend Component Changes

### QueryResults.tsx
**Changes:**
- Modify line 38: Add null/undefined checks for `results.data`
- Modify line 91: Show table headers even when data is empty
- Add new prop: `showEmptyTable` (default: true)
- Style empty state area with light background

### QueryExecutor.tsx
**Changes:**
- Add state: `validationResult` for early validation responses
- Modify execution flow: Call validation before creating approval
- Handle new `no_match` status with modal
- Add visual state components for each query state
- Modify line 459+: Add better null checking for results

### New Component: QueryValidationModal.tsx
**Props:**
- `isOpen: boolean`
- `validation: ValidationResult`
- `onEdit: () => void`
- `onCancel: () => void`

**Displays:**
- Info icon + message
- Affected rows count (0)
- Query preview
- Suggestion/help text
- Action buttons

### QueryStatusIndicator.tsx (NEW)
**Props:**
- `status: 'running' | 'completed' | 'empty' | 'failed' | 'no_match'`
- `executionTime?: number`
- `rowCount?: number`

**Renders appropriate:**
- Icon (spinner, checkmark, info, error)
- Color theme
- Status message
- Metrics

---

## 5. Backend Changes

### QueryService
**New Method:**
```go
func (s *QueryService) PreviewAndValidateWriteQuery(
    ctx context.Context, 
    queryText string, 
    dataSource *models.DataSource
) (*ValidationResult, error)
```

**Responsibilities:**
1. Parse query to detect operation type
2. Connect to data source
3. Run SELECT equivalent to count affected rows
4. Return validation result
5. NO transaction created, NO changes made

### Modified: CreateApprovalRequest
**Flow:**
```go
func (h *ApprovalHandler) CreateApprovalRequest(...) {
    // 1. Validate query syntax
    // 2. Validate schema (tables exist)
    // 3. For write queries:
    //    a. Run preview/validation
    //    b. If affected_rows == 0:
    //       - Return ValidationResult with status "no_match"
    //       - DO NOT create approval request
    //    c. If affected_rows > 0:
    //       - Continue with approval creation
}
```

---

## 6. Database Changes

None required - using existing query_results table structure.

---

## 7. Error Handling

### Scenarios:

1. **Validation query fails** (e.g., syntax error)
   - Treat as regular query error
   - Show error state

2. **Validation timeout** (slow query)
   - Skip validation, proceed to approval
   - Add warning: "Unable to verify affected rows, please review carefully"

3. **Validation returns error** (connection lost)
   - Skip validation, proceed to approval
   - Log error for debugging

---

## 8. Testing Strategy

### Backend Tests:
- Test `PreviewAndValidateWriteQuery` with:
  - UPDATE that affects 0 rows
  - UPDATE that affects > 0 rows
  - DELETE that affects 0 rows
  - Invalid query syntax
  - Connection errors

### Frontend Tests:
- Test `QueryResults` with empty data array
- Test `QueryExecutor` handling of `no_match` status
- Test visual state transitions
- Test modal interactions

### Integration Tests:
- Full flow: Submit UPDATE with no matches → See validation modal → Edit query
- Full flow: Submit UPDATE with matches → Create approval → Redirect to approvals

---

## 9. Success Criteria

1. ✅ SELECT queries returning 0 rows show table with column headers and helpful message
2. ✅ UPDATE/DELETE affecting 0 rows show validation modal, no approval created
3. ✅ Each query state has distinct visual appearance
4. ✅ Users can easily distinguish between "no data" and "query error"
5. ✅ Performance: Validation adds < 100ms to approval flow
6. ✅ All existing functionality continues to work

---

## 10. Implementation Phases

### Phase 1: Empty SELECT Results
- Backend: Ensure columns always returned
- Frontend: Modify QueryResults to show empty table

### Phase 2: Write Query Validation
- Backend: Add PreviewAndValidateWriteQuery method
- Backend: Modify approval flow
- Frontend: Handle no_match status

### Phase 3: Visual States
- Frontend: Create QueryStatusIndicator component
- Frontend: Update all state transitions
- Frontend: Add animations

### Phase 4: Testing & Polish
- Unit tests
- Integration tests
- UI polish

---

## 11. Open Questions

1. **Q:** Should we show the actual SELECT query used for validation?
   **A:** Yes, helps users understand what was checked

2. **Q:** What about INSERT queries that might violate constraints?
   **A:** Out of scope - INSERT validation is more complex, handle in future iteration

3. **Q:** Should we cache validation results?
   **A:** No, data may change between validation and approval

---

## 12. Related Documentation

- Multi-Query Transaction Design: `.opencode/plans/2025-01-13-multi-query-transaction-design.md`
- Approval Workflow: `internal/api/handlers/approval.go`
- Query Execution: `internal/service/query.go`

---

**Approved by:** Human Partner  
**Next Step:** Invoke `writing-plans` skill to create implementation plan
