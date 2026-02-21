# QueryBase Visual Flow Diagram

Quick visual reference for QueryBase query execution flow.

---

## Main Flow: User to Query Execution

```mermaid
flowchart TD
    Req["<b>USER REQUEST</b><br/>POST /api/v1/queries"] --> Auth["<b>AUTH MIDDLEWARE</b><br/>Validate JWT & Set Context"]
    Auth --> Route["<b>ROUTE: ExecuteQuery</b><br/>Parse DTO & Call Service"]
    Route --> Detect["<b>SERVICE: DetectOperationType</b><br/>Parse SQL & Match Patterns"]

    Detect --> Type{Operation Type?}
    Type -- "SELECT<br/>(Read)" --> PathA["<b>Path A: SELECT Query</b><br/>(Direct Execution)"]
    Type -- "WRITE<br/>(INSERT/UPDATE/DELETE/DDL)" --> PathB["<b>Path B: Write Query</b><br/>(Approval Workflow)"]
```

---

## Path A: SELECT Query (Direct Execution)

```mermaid
flowchart TD
    StartA([Path A Start]) --> DS["<b>1. VALIDATE DATA SOURCE</b><br/>Check existence & Get connection details"]
    DS --> Perms["<b>2. CHECK PERMISSIONS</b><br/>Verify can_read from user_permissions"]
    Perms --> Connect["<b>3. CONNECT TO DATA SOURCE</b><br/>Decrypt password & Open connection"]
    Connect --> Exec["<b>4. EXECUTE QUERY</b><br/>db.Raw().Rows() & Stream results"]
    Exec --> Process["<b>5. PROCESS RESULTS</b><br/>Scan rows into maps & Format types"]
    Process --> Cache["<b>6. CACHE RESULTS</b><br/>Serialize to JSON & Store in query_results"]
    Cache --> Log["<b>7. LOG HISTORY</b><br/>Create query_history entry"]
    Log --> Response["<b>8. RETURN RESPONSE</b><br/>JSON results & HTTP 200 OK"]
```

**Total Time:** ~100-500ms (depending on query)

---

## Path B: Write Query (Approval Workflow)

```mermaid
flowchart TD
    StartB([Path B Start]) --> DS["<b>1. VALIDATE DATA SOURCE</b><br/>Check existence & Get connection details"]
    DS --> Perms["<b>2. CHECK PERMISSIONS</b><br/>Verify can_write from user_permissions"]
    Perms --> Syntax["<b>3. VALIDATE SQL SYNTAX</b><br/>Parentheses, keywords, unterminated strings"]
    Syntax --> AppReq["<b>4. CREATE APPROVAL REQUEST</b><br/>Insert pending request into database"]
    AppReq --> FindApprovers["<b>5. FIND ELIGIBLE APPROVERS</b><br/>Filter users with can_approve = true"]
    FindApprovers --> NotifyApprovers["<b>6. SEND NOTIFICATIONS</b><br/>Google Chat webhooks to all approvers"]
    NotifyApprovers --> ResponseB["<b>7. RETURN PENDING RESPONSE</b><br/>HTTP 202 Accepted & approval_id"]

    ResponseB -- "(Wait for approver)" --> Review["<b>APPROVER REVIEW PHASE</b><br/>Review Decision & Comments"]
    Review --> Decision{Decision?}
    Decision -- Approved --> PathB1[Path B1: Approved Flow]
    Decision -- Rejected --> Rejected[Notify Requester ✗]
```

---

## Path B1: Approved Query Execution

```mermaid
flowchart TD
    StartB1([Path B1 Start]) --> Review["<b>1. CREATE APPROVAL REVIEW</b><br/>Insert review record & store decision"]
    Review --> Update["<b>2. UPDATE APPROVAL STATUS</b><br/>Set status = 'approved'"]
    Update --> Enqueue["<b>3. ENQUEUE TASK (Redis)</b><br/>Push ExecuteApprovedQuery to queue"]

    Enqueue -- "Asynq Worker" --> Worker["<b>BACKGROUND WORKER</b><br/>ExecuteApprovedQueryHandler"]
    Worker --> Tx["<b>4. TRANSACTIONAL EXECUTE</b><br/>BEGIN -> Execute -> COMMIT"]
    Tx --> Final["<b>5. FINALIZE EXECUTION</b><br/>Update status = 'executed' & Log history"]
    Final --> NotifyB1["<b>6. SEND NOTIFICATION</b><br/>Google Chat webhook completion"]
```

**Total Time:** 2-5 minutes (human wait) + 100-500ms (exec)

---

## Special Features Flow

### EXPLAIN Query Flow

```mermaid
flowchart TD
    ExpReq["<b>EXPLAIN REQUEST</b><br/>POST /api/v1/queries/explain"] --> ExpVal["<b>1. Validate Data Source</b><br/>Check permissions"]
    ExpVal --> ExpBuild["<b>2. Build EXPLAIN Query</b><br/>Prepend EXPLAIN [ANALYZE] to SQL"]
    ExpBuild --> ExpExec["<b>3. Execute EXPLAIN</b><br/>Run on target source"]
    ExpExec --> ExpResp["<b>4. Return Plan</b><br/>Return raw & formatted plan data"]
```

### Dry Run DELETE Flow

```mermaid
flowchart TD
    ReqDry["<b>DRY RUN DELETE REQUEST</b><br/>POST /api/v1/queries/dry-run"] --> Val["<b>1. Validate DELETE Operation</b><br/>Verify it's actually a DELETE"]
    Val --> Conv["<b>2. Convert to SELECT</b><br/>DELETE FROM users WHERE...<br/>↓<br/>SELECT * FROM users WHERE..."]
    Conv --> ExecDry["<b>3. Execute SELECT</b><br/>Run on target source & Get count"]
    ExecDry --> RespDry["<b>4. Return Preview</b><br/>Affected rows + Row data preview"]
```

---

## Database Flow Summary (Sequence)

```mermaid
sequenceDiagram
    participant User
    participant Auth as Auth & AuthZ
    participant DS as Data Sources (PostgreSQL)
    participant Worker as Worker (Redis/Asynq)
    participant Results as Query Results

    User->>Auth: 1. Validate Token (SELECT * FROM users)
    Auth->>DS: 2. Check Permissions (SELECT * FROM user_permissions)
    DS->>Auth: Permissions Valid
    Auth->>DS: 3. Get Data Source (SELECT host, pass... FROM data_sources)

    alt SELECT Operation
        Auth->>Results: 4a. Execute & Cache (INSERT query_results)
        Auth->>Results: INSERT query_history
    else WRITE Operation
        Auth->>DS: 4b. Create Approval (INSERT approval_requests)
        Auth->>DS: INSERT notifications
        User->>DS: 5. Approval Workflow (INSERT approval_reviews)
        DS->>Worker: 6. Pick up approved job
        Worker->>Results: INSERT query_results & history
        Worker->>DS: UPDATE approval_requests status='executed'
    end
```

---

## Quick Decision Tree

```mermaid
flowchart TD
    Start([User Submits Query]) --> Auth{Authenticated?}
    Auth -- NO --> 401[Return 401]
    Auth -- YES --> Type{Operation Type?}

    Type -- SELECT --> Direct[Direct Execution]
    Type -- "INSERT/UPDATE/DELETE" --> Approval[Approval Workflow]
    Type -- DDL --> Approval

    Direct --> Execute([Execute & Return])

    Approval --> CreateReq[Create Approval Request]
    CreateReq --> Wait[Wait for Approval]

    Wait --> Decision{Approval Decision?}
    Decision -- Approved --> ExecuteApp[Execute Approved Query]
    Decision -- Rejected --> Notify[Notify Requester ✗]

    ExecuteApp --> NotifyApp[Notify Requester ✓]

    Notify --> Finish([Return])
    NotifyApp --> Finish
    Execute --> Finish
```

---

## Time Estimates

| Operation          | Phase             | Time Estimate |
| :----------------- | :---------------- | :------------ |
| **SELECT**         | Total Lifecycle   | 80-600ms      |
| **Write (Submit)** | Submission only   | 80-165ms      |
| **Write (Wait)**   | Human approval    | 2m - 7 days   |
| **Write (Exec)**   | Worker processing | 100-700ms     |

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
