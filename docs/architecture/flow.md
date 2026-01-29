# QueryBase Visual Flow Diagram

Quick visual reference for QueryBase query execution flow.

---

## Main Flow: User to Query Execution

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           USER REQUEST                                     │
│  POST /api/v1/queries                                                      │
│  {                                                                         │
│    "data_source_id": "uuid",                                              │
│    "query_text": "SELECT * FROM users WHERE status = 'active'"            │
│  }                                                                         │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                    AUTH MIDDLEWARE                                          │
│  ✓ Validate JWT token                                                     │
│  ✓ Extract user_id, role                                                   │
│  ✓ Set context                                                             │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                   ROUTE: ExecuteQuery                                      │
│  ✓ Parse request DTO                                                       │
│  ✓ Get user_id from context                                                │
│  ✓ Call QueryService.ExecuteQuery()                                        │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                  SERVICE: DetectOperationType                              │
│  ✓ Parse SQL query                                                         │
│  ✓ Match against patterns (SELECT, INSERT, UPDATE, DELETE, etc.)           │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
         ┌──────────────┐         ┌──────────────┐
         │   SELECT     │         │  WRITE       │
         │  Operation   │         │  Operations   │
         │              │         │              │
         │  (Direct)    │         │ (Approval)    │
         └──────┬───────┘         └──────┬───────┘
                │                        │
                │                        │
                ▼                        ▼
```

---

## Path A: SELECT Query (Direct Execution)

```
SELECT QUERY FLOW (Read Operations)
═════════════════════════════════════

┌────────────────────────────────────────────────────────────────────────────┐
│  1. VALIDATE DATA SOURCE                                                   │
│     ✓ Check data_source exists in database                                 │
│     ✓ Get connection details (host, port, username, encrypted password)    │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  2. CHECK PERMISSIONS                                                       │
│     ✓ Query user_permissions view                                          │
│     ✓ Verify can_read = true                                               │
│     ✓ OR user is admin                                                     │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  3. CONNECT TO DATA SOURCE                                                  │
│     ✓ Decrypt password (AES-256-GCM)                                       │
│     ✓ Build DSN string                                                     │
│     ✓ Open connection (PostgreSQL or MySQL)                                │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  4. EXECUTE QUERY                                                           │
│     ✓ db.Raw(queryText).Rows()                                             │
│     ✓ Stream results row by row                                            │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  5. PROCESS RESULTS                                                         │
│     ✓ Get column names                                                     │
│     ✓ Scan rows into maps                                                  │
│     ✓ Convert byte[] to string                                             │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  6. CACHE RESULTS                                                           │
│     ✓ Serialize to JSON                                                    │
│     ✓ Create query_result record                                           │
│     ✓ Save to database                                                     │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  7. LOG HISTORY                                                             │
│     ✓ Create query_history entry                                           │
│     ✓ Record: user_id, query_text, row_count, execution_time              │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  8. RETURN RESPONSE                                                         │
│     ✓ JSON with: query_id, status, row_count, data, columns               │
│     ✓ HTTP 200 OK                                                           │
└────────────────────────────────────────────────────────────────────────────┘

Total Time: ~100-500ms (depending on query)
```

---

## Path B: Write Query (Approval Workflow)

```
WRITE QUERY FLOW (INSERT/UPDATE/DELETE/DDL)
═══════════════════════════════════════════

┌────────────────────────────────────────────────────────────────────────────┐
│  1. VALIDATE DATA SOURCE                                                   │
│     ✓ Check data_source exists in database                                 │
│     ✓ Get connection details                                                │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  2. CHECK PERMISSIONS                                                       │
│     ✓ Query user_permissions view                                          │
│     ✓ Verify can_write = true                                              │
│     ✓ OR user is admin                                                     │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  3. VALIDATE SQL SYNTAX                                                     │
│     ✓ Check empty query                                                     │
│     ✓ Check balanced parentheses                                            │
│     ✓ Check unterminated strings                                            │
│     ✓ Check required keywords (VALUES, SET, FROM)                          │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  4. CREATE APPROVAL REQUEST                                                 │
│     ✓ Generate approval_request_id (UUID)                                  │
│     ✓ Insert into approval_requests:                                       │
│       - query_text, operation_type, data_source_id, requested_by          │
│     ✓ Set status = 'pending'                                               │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  5. FIND ELIGIBLE APPROVERS                                                 │
│     ✓ Query user_permissions view                                          │
│     ✓ Filter users with can_approve = true                                 │
│     ✓ For current data_source_id                                           │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  6. SEND NOTIFICATIONS                                                      │
│     ✓ For each eligible approver:                                          │
│       - Insert notification record                                         │
│       - Send Google Chat webhook                                          │
│       - Include query_text, operation_type, approval_id                    │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  7. RETURN PENDING RESPONSE                                                 │
│     ✓ JSON with: approval_id, status='pending', requires_approval=true    │
│     ✓ HTTP 202 Accepted                                                    │
└────────────────────────────────────────────────────────────────────────────┘
                             │
                             │  (Wait for approver action)
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                   APPROVER REVIEW PHASE                                    │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Approver submits:                                                     │ │
│  │  POST /api/v1/approvals/:id/review                                     │ │
│  │  {                                                                     │ │
│  │    "decision": "approved" or "rejected",                              │ │
│  │    "comments": "Looks good"                                           │ │
│  │  }                                                                     │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
         ┌──────────────┐         ┌──────────────┐
         │  APPROVED    │         │   REJECTED   │
         └──────┬───────┘         └──────────────┘
                │
                ▼
```

---

## Path B1: Approved Query Execution

```
APPROVED QUERY EXECUTION
═════════════════════════

┌────────────────────────────────────────────────────────────────────────────┐
│  1. CREATE APPROVAL REVIEW                                                  │
│     ✓ Insert approval_review record                                        │
│     ✓ Store: decision, comments, reviewer_id, reviewed_at                │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  2. UPDATE APPROVAL STATUS                                                  │
│     ✓ UPDATE approval_requests SET status = 'approved'                    │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  3. ENQUEUE TASK (Redis Queue)                                              │
│     ✓ Create task: ExecuteApprovedQuery                                    │
│     ✓ Payload: {approval_id: uuid}                                        │
│     ✓ Push to Asynq queue                                                  │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                   BACKGROUND WORKER PICKS UP TASK                          │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │  Worker: ExecuteApprovedQueryHandler                                  │ │
│  │  - Runs asynchronously                                                │ │
│  │  - Processes one approval at a time                                   │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  4. START TRANSACTION (Optional Preview)                                  │
│     ✓ BEGIN transaction on data source                                    │
│     ✓ Execute query in transaction                                        │
│     ✓ Return preview results to approver                                  │
│     ✓ Wait for: COMMIT or ROLLBACK decision                              │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
         ┌──────────────┐         ┌──────────────┐
         │   COMMIT     │         │  ROLLBACK    │
         └──────┬───────┘         └──────────────┘
                │
                ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  5. FINALIZE EXECUTION (After Commit)                                      │
│     ✓ Create query_result record                                           │
│     ✓ Update approval_requests status = 'executed'                        │
│     ✓ Log to query_history                                                 │
│     ✓ Notify requester of success                                         │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  6. SEND COMPLETION NOTIFICATION                                            │
│     ✓ Google Chat webhook to requester                                    │
│     ✓ Include: row_count, execution_time, status                         │
└────────────────────────────────────────────────────────────────────────────┘

Total Time: 2-5 minutes (including approval wait)
Actual Execution: 100-500ms
```

---

## Special Features Flow

### EXPLAIN Query Flow

```
EXPLAIN QUERY FLOW
═════════════════

User Request:
POST /api/v1/queries/explain
{
  "data_source_id": "uuid",
  "query_text": "SELECT * FROM users WHERE email = '...'",
  "analyze": false  // or true for EXPLAIN ANALYZE
}
    │
    ▼
┌───────────────────────────────────┐
│  1. Validate Data Source          │
│     ✓ Check permissions (can_read)│
└───────────┬───────────────────────┘
            ▼
┌───────────────────────────────────┐
│  2. Build EXPLAIN Query           │
│     EXPLAIN SELECT ...            │
│     or EXPLAIN ANALYZE SELECT ... │
└───────────┬───────────────────────┘
            ▼
┌───────────────────────────────────┐
│  3. Execute EXPLAIN               │
│     ✓ Run on data source          │
│     ✓ Get execution plan          │
└───────────┬───────────────────────┘
            ▼
┌───────────────────────────────────┐
│  4. Return Plan                   │
│  {                                │
│    "plan": [...],                 │
│    "raw_output": "..."            │
│  }                                │
└───────────────────────────────────┘
```

### Dry Run DELETE Flow

```
DRY RUN DELETE FLOW
═══════════════════

User Request:
POST /api/v1/queries/dry-run
{
  "data_source_id": "uuid",
  "query_text": "DELETE FROM users WHERE status = 'inactive'"
}
    │
    ▼
┌───────────────────────────────────┐
│  1. Validate DELETE Operation     │
│     ✓ Check operation type        │
│     ✓ Must be DELETE              │
└───────────┬───────────────────────┘
            ▼
┌───────────────────────────────────┐
│  2. Convert to SELECT             │
│  DELETE FROM users WHERE ...      │
│     ↓                             │
│  SELECT * FROM users WHERE ...    │
└───────────┬───────────────────────┘
            ▼
┌───────────────────────────────────┐
│  3. Execute SELECT                │
│     ✓ Run converted query         │
│     ✓ Get affected rows           │
└───────────┬───────────────────────┘
            ▼
┌───────────────────────────────────┐
│  4. Return Preview                │
│  {                                │
│    "affected_rows": 3,            │
│    "query": "SELECT * ...",       │
│    "rows": [...]                  │
│  }                                │
└───────────────────────────────────┘
```

---

## Database Flow Summary

```
QueryBase PostgreSQL Database
═════════════════════════════

┌─────────────────┐
│   Authentication│
│   & AuthZ       │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  1. VALIDATE TOKEN                                 │
│     SELECT * FROM users WHERE id = user_id         │
└────────┬────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  2. CHECK PERMISSIONS                               │
│     SELECT * FROM user_permissions                  │
│     WHERE user_id = ? AND data_source_id = ?        │
└────────┬────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  3. GET DATA SOURCE                                 │
│     SELECT * FROM data_sources WHERE id = ?         │
│     - Returns encrypted password                    │
└────────┬────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  4a. FOR SELECT: EXECUTE & CACHE                    │
│     INSERT INTO query_results (...)                 │
│     INSERT INTO query_history (...)                 │
│                                                     │
│  4b. FOR WRITE: CREATE APPROVAL                    │
│     INSERT INTO approval_requests (...)             │
│     INSERT INTO notifications (...)                 │
└─────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  5. APPROVAL WORKFLOW                               │
│     INSERT INTO approval_reviews (...)              │
│     UPDATE approval_requests SET status = ...       │
└─────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  6. EXECUTE (after approval)                        │
│     INSERT INTO query_results (...)                 │
│     INSERT INTO query_history (...)                 │
│     UPDATE approval_requests SET status = 'executed'│
└─────────────────────────────────────────────────────┘
```

---

## Quick Decision Tree

``                        USER SUBMITS QUERY
                                   │
                        ┌──────────┴──────────┐
                        │   Is Authenticated? │
                        └──────────┬──────────┘
                                   │ NO
                    ┌──────────────┴──────────────┐
                    │                             │
                    YES                           RETURN 401
                    │                             │
                    ▼                             │
        ┌─────────────────────────┐               │
        │  What operation type?  │               │
        └────────────┬────────────┘               │
                     │                             │
        ┌────────────┼────────────┐               │
        │            │            │               │
        ▼            ▼            ▼               │
    SELECT    INSERT/UPDATE/DELETE   DDL          │
        │            │            │               │
        ▼            ▼            ▼               │
    ┌────────┐  ┌─────────┐  ┌─────────┐         │
    │Direct │  │Approval│  │Approval │         │
    │Exec   │  │Workflow│  │Workflow│         │
    └───┬────┘  └────┬────┘  └────┬────┘         │
        │            │            │               │
        ▼            ▼            ▼               │
   [Execute]    [Create      [Create           │
     & Return    Approval]    Approval]         │
                  Request      Request          │
                        │            │            │
                        ▼            ▼            │
                  [Wait for    [Wait for       │
                   Approval]    Approval]       │
                        │            │            │
                  ┌───┴────────┐   └───────┐     │
                  │            │           │     │
             Approved      Rejected    Approved │
                  │            │         │     │
                  ▼            │         │     │
            [Execute]         │         │     │
            [Return]          │         │     │
                               │         │     │
                        Notify    Notify   │     │
                        Requester Requester│
                           ✓         ✗      │
                                      │     │
                                      └─────┴──→ RETURN
```

---

## Time Estimates

```
SELECT Query (Direct):
  - Auth: 1-5ms
  - Permission check: 5-10ms
  - Connect to data source: 10-50ms
  - Execute query: 50-500ms (varies)
  - Cache results: 10-20ms
  - Log history: 5-10ms
  ────────────────────────────────
  Total: 80-600ms

Write Query (Approval):
  - Auth: 1-5ms
  - Permission check: 5-10ms
  - Validate SQL: 5-10ms
  - Create approval: 10-20ms
  - Find approvers: 10-20ms
  - Send notifications: 50-100ms
  ────────────────────────────────
  Total: 80-165ms (for submission)

  Approval Wait: 2 minutes to 7 days (human)

  Execution (after approval):
  - Worker pickup: 1-5 seconds (queue poll)
  - Connect to data source: 10-50ms
  - Start transaction: 5-10ms
  - Execute query: 50-500ms
  - Commit: 10-50ms
  - Cache & log: 10-20ms
  - Send notification: 50-100ms
  ────────────────────────────────
  Total: 100-700ms (actual execution)
```

---

## Key Takeaways

1. **SELECT queries** execute immediately (80-600ms)
2. **Write queries** go through approval workflow (minutes to days)
3. **Authentication** via JWT on every request
4. **Authorization** via user_permissions view (RBAC)
5. **All queries** logged to query_history
6. **SELECT results** cached in query_results table
7. **Write operations** tracked via approval_requests
8. **Notifications** sent via Google Chat webhooks
9. **Background workers** execute approved write queries
10. **Audit trail** complete and queryable
