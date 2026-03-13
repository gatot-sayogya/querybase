# Multi-Query Transaction Support

**Date:** 2025-01-13  
**Status:** Approved  
**Approach:** Extend Existing Transaction System

## Overview

This design extends QueryBase to support executing multiple SQL queries within a single database transaction. Users can submit multiple queries (e.g., UPDATE followed by INSERT) separated by semicolons, which are executed atomically with a single approval flow.

## Requirements

1. Support any number of queries (2+) in a single transaction
2. All queries in the transaction must be approved as a single unit
3. Display preview data for all write operations before execution
4. Show summary view with option to expand details per query
5. Maintain atomicity - all queries succeed or all rollback
6. Reuse existing approval infrastructure

## Architecture

### Core Concept
Extend the existing `query_transactions` system to support multiple queries per transaction by adding a child table to store individual SQL statements.

### Key Changes
1. **Database Schema:** Add `query_transaction_statements` table
2. **API Layer:** New endpoints for multi-query operations
3. **Service Layer:** Sequential statement execution within transaction
4. **Preview System:** Generate previews for all write operations

### Transaction Flow
1. User submits multiple queries separated by semicolons
2. Backend parses and validates each query
3. Generate previews for all write operations
4. Single approval covers all queries
5. Execute all queries in one database transaction
6. Return aggregated results with per-query details

## Database Schema

### New Table: `query_transaction_statements`

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Statement ID |
| `transaction_id` | UUID (FK → query_transactions.id) | Parent transaction |
| `sequence` | INT | Execution order (0, 1, 2...) |
| `query_text` | TEXT | Individual SQL statement |
| `operation_type` | VARCHAR(20) | SELECT/INSERT/UPDATE/DELETE |
| `status` | VARCHAR(20) | pending/success/failed |
| `affected_rows` | INT | Rows affected |
| `error_message` | TEXT | Error details if failed |
| `preview_data` | JSONB | Preview rows (write ops) |
| `before_data` | JSONB | Before state for audit |
| `after_data` | JSONB | After state for audit |
| `execution_time_ms` | INT | Execution duration |
| `created_at` | TIMESTAMP | Record creation |

### Modified Table: `query_transactions`

Keep existing fields for backward compatibility. The `query_text` field stores the full batch text for reference, while individual statements are stored in the new table.

## API Design

### New DTOs

```go
// MultiQueryRequest represents a multi-query execution request
type MultiQueryRequest struct {
    DataSourceID string   `json:"data_source_id" binding:"required"`
    QueryTexts   []string `json:"query_texts" binding:"required,min=1"`
    Name         string   `json:"name"`
    Description  string   `json:"description"`
}

// MultiQueryPreviewRequest represents a preview request
type MultiQueryPreviewRequest struct {
    DataSourceID string   `json:"data_source_id" binding:"required"`
    QueryTexts   []string `json:"query_texts" binding:"required,min=1"`
}

// StatementResult represents individual statement results
type StatementResult struct {
    Sequence     int                    `json:"sequence"`
    QueryText    string                 `json:"query_text"`
    OperationType string                `json:"operation_type"`
    Status       string                 `json:"status"`
    AffectedRows int                    `json:"affected_rows"`
    RowCount     int                    `json:"row_count"`
    Columns      []ColumnInfo           `json:"columns,omitempty"`
    Data         []map[string]interface{} `json:"data,omitempty"`
    ErrorMessage string                 `json:"error_message,omitempty"`
    Preview      *WritePreview          `json:"preview,omitempty"`
}

// MultiQueryResponse represents the execution response
type MultiQueryResponse struct {
    QueryID          string            `json:"query_id"`
    Status           string            `json:"status"`
    IsMultiQuery     bool              `json:"is_multi_query"`
    StatementCount   int               `json:"statement_count"`
    TotalAffectedRows int              `json:"total_affected_rows"`
    ExecutionTimeMs  int               `json:"execution_time_ms"`
    Statements       []StatementResult `json:"statements"`
    ErrorMessage     string            `json:"error_message,omitempty"`
    RequiresApproval bool              `json:"requires_approval"`
    ApprovalID       string            `json:"approval_id,omitempty"`
}
```

### New Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/queries/multi/preview` | Preview all statements |
| POST | `/api/v1/queries/multi/execute` | Execute multi-query transaction |
| GET | `/api/v1/queries/multi/:id/statements` | Get statement results |
| POST | `/api/v1/queries/multi/:id/commit` | Commit transaction |
| POST | `/api/v1/queries/multi/:id/rollback` | Rollback transaction |

## Execution Flow

### Frontend Flow
1. User enters multiple queries in SQL editor (semicolon-separated)
2. Frontend parses and splits queries (handles semicolons inside strings/comments)
3. Sends array to `/multi/preview` endpoint
4. Displays preview modal with expandable details per query
5. User approves → sends to `/multi/execute`
6. Polls for results, displays summary with expand option

### Backend Flow
1. **Parse Phase:** Split query text, validate each statement
2. **Permission Phase:** Check permissions for each operation type
3. **Preview Phase:** Generate previews for all write operations
4. **Approval Phase:** Create single approval with all queries attached
5. **Execution Phase:**
   - Begin database transaction
   - Execute statements sequentially
   - Track affected rows per statement
   - Capture audit data per statement
   - Commit on success, rollback on failure
6. **Results Phase:** Return aggregated results with per-statement details

## Error Handling

- **Parse Error:** Return which statement failed to parse with position
- **Permission Error:** Return which statement lacks permissions
- **Execution Error:** Rollback entire transaction, return statement number that failed
- **Partial Results:** Never stored - atomic guarantee maintained

## UI/UX Design

### SQL Editor
- Support semicolon-separated queries
- Visual indicator showing query separation
- Syntax highlighting for each statement

### Preview Modal
- Summary view showing:
  - Total statements count
  - Total affected rows estimate
  - List of operations (INSERT/UPDATE/DELETE)
- Expandable sections for each statement:
  - SQL preview
  - Affected rows estimate
  - Sample data (first 10 rows)

### Results View
- Summary panel:
  - Total execution time
  - Total affected rows
  - Success/failure status
- Expandable statement cards:
  - Statement SQL
  - Execution time
  - Affected rows
  - Result data (for SELECTs)
  - Error details (if failed)

## Testing Strategy

### Unit Tests
- Query splitting logic (handles edge cases)
- Multi-query preview generation
- Statement execution order
- Error handling and rollback

### Integration Tests
- End-to-end multi-query execution
- Transaction atomicity verification
- Audit data capture per statement
- Approval flow for multi-query

### Edge Cases
- Semicolons inside string literals
- Empty statements
- Mixed SELECT and write operations
- Very large batch sizes (performance)
- Statement that affects 0 rows

## Migration Strategy

1. Create new `query_transaction_statements` table
2. Backfill existing single-statement transactions
3. Keep `query_transactions.query_text` for backward compatibility
4. Update approval handlers to support multi-query
5. Deploy frontend changes for multi-query UI

## Security Considerations

- All statements must pass permission checks individually
- SQL injection prevention via parameterized queries
- Audit logging for each statement
- Transaction timeout to prevent long-running locks
- Statement count limit to prevent abuse

## Performance Considerations

- Preview generation parallelized where possible
- Statement count limit (default: 50)
- Total batch size limit (default: 100KB)
- Transaction timeout (default: 30 seconds)
- Connection pooling for concurrent executions

## Future Enhancements

- Query templates for common multi-query patterns
- Save/load multi-query batches
- Batch execution history and replay
- Statement-level rollback within transaction
- Dry-run mode for validation only

---

## Approval

**Approved by:** User  
**Date:** 2025-01-13  
**Status:** Ready for Implementation
