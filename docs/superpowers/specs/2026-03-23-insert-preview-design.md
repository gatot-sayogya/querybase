# INSERT Query Preview Design Document

**Date:** 2026-03-23  
**Feature:** INSERT Query Preview for Approval Workflow  
**Author:** QueryBase Development Team  
**Status:** Approved for Implementation

---

## 1. Executive Summary

### Problem
When an approver reviews an INSERT query approval request, they currently see an error: `"preview is only available for DELETE and UPDATE queries"`. This prevents approvers from understanding what data will be inserted before approving, creating a security and usability gap.

### Solution
Implement INSERT query preview functionality that:
- **For VALUES clauses:** Parses the INSERT statement to extract and display the data to be inserted
- **For SELECT clauses:** Executes the SELECT portion with LIMIT 50 to preview source rows
- **Maximum preview:** Shows up to 50 rows in both cases
- **Safe execution:** Uses read-only operations, no actual INSERT execution

### Success Criteria
- [ ] Approvers can preview INSERT queries with VALUES clause
- [ ] Approvers can preview INSERT queries with SELECT clause
- [ ] Preview displays up to 50 rows maximum
- [ ] Preview shows target table context
- [ ] Preview is safe (read-only, no side effects)
- [ ] Preview handles edge cases (JSON data, special characters, NULL values)

---

## 2. Background & Context

### Current State
- `PreviewWriteQuery()` in `internal/service/query.go` only supports DELETE and UPDATE
- `ApprovalDetail.tsx` tries to call `previewWriteQuery()` for all write operations
- INSERT queries fail with: `"preview is only available for DELETE and UPDATE queries"`

### Related Systems
- **QueryService**: Contains query preview logic
- **Approval Workflow**: 3-phase transaction system (preview → start transaction → commit/rollback)
- **Multi-query support**: Similar preview logic exists for multi-query transactions
- **Frontend**: ApprovalDetail component handles preview display

---

## 3. Requirements

### Functional Requirements

#### 3.1 VALUES Clause Preview (FR-VAL-001 to FR-VAL-005)

**FR-VAL-001:** Parse INSERT statement to extract:
- Target table name
- Column names (if specified)
- Value rows (parsed from VALUES clause)

**FR-VAL-002:** Handle single-row INSERT:
```sql
INSERT INTO users (name, email) VALUES ('John', 'john@example.com')
```

**FR-VAL-003:** Handle multi-row INSERT:
```sql
INSERT INTO users (name, email) VALUES 
  ('John', 'john@example.com'), 
  ('Jane', 'jane@example.com'),
  ('Bob', 'bob@example.com')
```

**FR-VAL-004:** Handle INSERT without explicit columns:
```sql
INSERT INTO users VALUES (1, 'John', 'john@example.com')
```
*Note: When columns not specified, fetch table schema to show column names*

**FR-VAL-005:** Parse values with special characters:
- Single quotes: `'O''Brien'`
- JSON data: `'{"key": "value"}'`
- NULL values: `NULL`
- Numbers, booleans, dates

#### 3.2 SELECT Clause Preview (FR-SEL-001 to FR-SEL-003)

**FR-SEL-001:** Extract SELECT clause from INSERT:
```sql
INSERT INTO audit_log SELECT * FROM temp_logs WHERE processed = false
```
→ Extract: `SELECT * FROM temp_logs WHERE processed = false`

**FR-SEL-002:** Execute extracted SELECT with LIMIT 50:
- Connect to data source
- Add `LIMIT 50` if not present
- Execute and return rows

**FR-SEL-003:** Return row count indicator:
- Show "Showing 50 of 200 rows" when more rows exist
- Help approvers understand insert volume

#### 3.3 Schema Enhancement (FR-SCH-001 to FR-SCH-003)

**FR-SCH-001:** Fetch target table schema:
- Column names
- Column types (for display)
- Required vs optional columns

**FR-SCH-002:** Display schema context:
- Show target table name prominently
- Display column headers in preview table
- Indicate which columns will receive data

**FR-SCH-003:** Handle schema mismatches gracefully:
- If columns not specified, map by position
- Show warning if column count mismatch

#### 3.4 Safety & Constraints (FR-SAF-001 to FR-SAF-004)

**FR-SAF-001:** Read-only operations only:
- Never execute INSERT during preview
- Parse VALUES or execute SELECT, never INSERT

**FR-SAF-002:** No side effects:
- Don't increment sequences
- Don't fire triggers
- Don't modify data

**FR-SAF-003:** Row limit enforcement:
- Maximum 50 rows in preview
- For VALUES: parse all but only return first 50
- For SELECT: add LIMIT 50

**FR-SAF-004:** Error handling:
- Invalid SQL syntax → return parse error
- Table doesn't exist → return schema error
- SELECT fails → return execution error

---

### Non-Functional Requirements

#### 3.5 Performance (NFR-PERF-001 to NFR-PERF-002)

**NFR-PERF-001:** Preview generation < 2 seconds:
- Parse VALUES quickly (no DB round-trip)
- SELECT with LIMIT 50 should be fast

**NFR-PERF-002:** Handle large multi-row INSERTs:
- Parse VALUES clauses with 1000+ rows efficiently
- Only return first 50, but parse all to get total count

#### 3.6 Security (NFR-SEC-001 to NFR-SEC-002)

**NFR-SEC-001:** Permission checks:
- User must have INSERT permission on target table
- User must have SELECT permission for SELECT-based INSERTs

**NFR-SEC-002:** SQL injection prevention:
- Parse VALUES, don't execute as SQL
- Use parameterized queries for SELECT execution
- Validate table and column names

#### 3.7 Usability (NFR-UX-001 to NFR-UX-003)

**NFR-UX-001:** Clear preview display:
- Table format with headers
- Row numbers
- Clear indication of data to be inserted

**NFR-UX-002:** Empty state handling:
- SELECT returns 0 rows → show "No rows to insert"
- Approver can still proceed (unlike UPDATE/DELETE)

**NFR-UX-003:** Visual distinction:
- Different color/icon for INSERT vs UPDATE/DELETE previews
- Green/blue theme for INSERT (indicating "addition")

---

## 4. Architecture

### 4.1 High-Level Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    INSERT Preview Flow                       │
└─────────────────────────────────────────────────────────────┘

1. Approver clicks "Preview" on INSERT approval
   │
   ▼
2. Backend detects INSERT operation type
   │
   ├─► VALUES clause? ──► Parse SQL extract values
   │                        │
   │                        ▼
   │                     Return structured data
   │
   └─► SELECT clause? ──► Extract SELECT
                            │
                            ▼
                         Execute SELECT LIMIT 50
                            │
                            ▼
                         Return result rows
   │
   ▼
3. Frontend displays preview table
   │
   ▼
4. Approver reviews and proceeds to transaction
```

### 4.2 Component Architecture

#### Backend Components

```
┌──────────────────────────────────────────────────────────┐
│                   QueryService                            │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │  PreviewInsertQuery()                               │  │
│  │  ├── detectInsertType(queryText)                    │  │
│  │  │   ├── VALUES → parseInsertValues()              │  │
│  │  │   └── SELECT → executeSelectForInsert()         │  │
│  │  │                                                 │  │
│  │  ├── getTableSchema(tableName)                    │  │
│  │  │   └── Returns: columns, types                   │  │
│  │  │                                                 │  │
│  │  └── buildPreviewResult()                         │  │
│  │      └── Returns: InsertPreviewResult              │  │
│  └────────────────────────────────────────────────────┘  │
│                                                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Existing Methods                                   │  │
│  │  ├── PreviewWriteQuery() (UPDATE/DELETE)           │  │
│  │  └── PreviewMultiQuery() (multi-query)             │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

#### Frontend Components

```
┌──────────────────────────────────────────────────────────┐
│              ApprovalDetail Component                     │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │  handlePreviewQuery()                               │  │
│  │  ├── if INSERT → previewInsertQuery()              │  │
│  │  ├── if UPDATE/DELETE → previewWriteQuery()        │  │
│  │  └── if multi-query → previewMultiQuery()          │  │
│  └────────────────────────────────────────────────────┘  │
│                                                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Conditional Rendering                              │  │
│  │  ├── InsertPreviewPanel (NEW)                      │  │
│  │  ├── DataChangesPanel (existing UPDATE/DELETE)     │  │
│  │  └── MultiQueryPreviewModal (existing multi-query) │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────┐
│         InsertPreviewPanel Component (NEW)                │
├──────────────────────────────────────────────────────────┤
│  Props:                                                   │
│  - preview: InsertPreviewResult                          │
│  - tableName: string                                     │
│  - rowCount: number                                      │
│  - columns: ColumnInfo[]                                 │
│  - rows: RowData[]                                       │
│                                                           │
│  Features:                                                │
│  - Table display with headers                            │
│  - Row numbers (1-50)                                    │
│  - "Showing X of Y" indicator                            │
│  - Target table badge                                    │
│  - Color theme: blue/green (INSERT)                      │
└──────────────────────────────────────────────────────────┘
```

---

## 5. Data Models

### 5.1 Backend Models

```go
// InsertPreviewResult represents the preview for an INSERT query
type InsertPreviewResult struct {
    TableName      string                   `json:"table_name"`
    Columns        []ColumnInfo             `json:"columns"`
    Rows           []map[string]interface{} `json:"rows"`
    TotalRowCount  int                      `json:"total_row_count"`
    PreviewType    InsertPreviewType        `json:"preview_type"` // "values" or "select"
    SelectQuery    string                   `json:"select_query,omitempty"` // For SELECT-based inserts
}

type InsertPreviewType string

const (
    InsertPreviewTypeValues InsertPreviewType = "values"
    InsertPreviewTypeSelect InsertPreviewType = "select"
)

type ColumnInfo struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Nullable bool   `json:"nullable"`
}
```

### 5.2 API Request/Response

**Request:**
```http
POST /api/v1/queries/preview-insert
Content-Type: application/json

{
  "data_source_id": "uuid-here",
  "query_text": "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"
}
```

**Response (VALUES):**
```json
{
  "table_name": "users",
  "columns": [
    {"name": "name", "type": "varchar", "nullable": false},
    {"name": "email", "type": "varchar", "nullable": false}
  ],
  "rows": [
    {"name": "John", "email": "john@example.com"},
    {"name": "Jane", "email": "jane@example.com"}
  ],
  "total_row_count": 2,
  "preview_type": "values"
}
```

**Response (SELECT):**
```json
{
  "table_name": "audit_log",
  "columns": [
    {"name": "id", "type": "int", "nullable": false},
    {"name": "action", "type": "varchar", "nullable": false},
    {"name": "created_at", "type": "timestamp", "nullable": false}
  ],
  "rows": [
    {"id": 1, "action": "login", "created_at": "2024-01-01T10:00:00Z"},
    {"id": 2, "action": "logout", "created_at": "2024-01-01T18:00:00Z"}
  ],
  "total_row_count": 150,
  "preview_type": "select",
  "select_query": "SELECT id, action, created_at FROM temp_logs WHERE processed = false LIMIT 50"
}
```

### 5.3 Frontend Types

```typescript
interface InsertPreviewResult {
  table_name: string;
  columns: ColumnInfo[];
  rows: Record<string, unknown>[];
  total_row_count: number;
  preview_type: 'values' | 'select';
  select_query?: string;
}

interface ColumnInfo {
  name: string;
  type: string;
  nullable: boolean;
}
```

---

## 6. Implementation Details

### 6.1 Backend Implementation

#### 6.1.1 New Method: PreviewInsertQuery

**Location:** `internal/service/query.go`

```go
// PreviewInsertQuery previews the data that will be inserted by an INSERT query
// without actually modifying any data.
// For VALUES: parses the SQL to extract values
// For SELECT: executes the SELECT with LIMIT 50
func (s *QueryService) PreviewInsertQuery(
    ctx context.Context, 
    queryText string, 
    dataSource *models.DataSource
) (*InsertPreviewResult, error) {
    // 1. Validate this is an INSERT
    operationType := DetectOperationType(queryText)
    if operationType != models.OperationInsert {
        return nil, fmt.Errorf("preview is only available for INSERT queries")
    }
    
    // 2. Detect INSERT type (VALUES vs SELECT)
    insertType := detectInsertType(queryText)
    
    // 3. Extract table name
    tableName := extractInsertTableName(queryText)
    
    // 4. Get table schema
    schema, err := s.getTableSchema(ctx, dataSource, tableName)
    if err != nil {
        // Continue without schema, use parsed column names
    }
    
    // 5. Generate preview based on type
    switch insertType {
    case InsertTypeValues:
        return s.previewInsertValues(queryText, tableName, schema)
    case InsertTypeSelect:
        return s.previewInsertSelect(ctx, queryText, tableName, schema, dataSource)
    default:
        return nil, fmt.Errorf("unsupported INSERT type")
    }
}
```

#### 6.1.2 Helper: Parse VALUES Clause

```go
// parseInsertValues extracts column names and value rows from INSERT ... VALUES
func parseInsertValues(queryText string) (*InsertValuesResult, error) {
    // Regex pattern for INSERT INTO table [(cols)] VALUES (vals), (vals), ...
    // Handle:
    // - Optional column list
    // - Multiple value rows
    // - Escaped quotes: 'O''Brien'
    // - JSON data: '{"key": "value"}'
    // - NULL values
    
    pattern := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(?:"?([^"]+)"?|\w+)\s*(?:\(([^)]+)\))?\s*VALUES\s+(.+)`)
    matches := pattern.FindStringSubmatch(queryText)
    
    // Parse column names
    var columns []string
    if matches[2] != "" {
        columns = parseColumnList(matches[2])
    }
    
    // Parse value rows (handling parentheses, commas, quotes)
    valuesSection := matches[3]
    rows, err := parseValueRows(valuesSection)
    
    return &InsertValuesResult{
        Columns: columns,
        Rows: rows,
    }, nil
}

// parseValueRows parses comma-separated value tuples
// Returns array of rows, each row is array of values
func parseValueRows(valuesSection string) ([][]string, error) {
    // Complex parsing to handle:
    // - Nested parentheses
    // - Quoted strings with commas: 'Hello, World'
    // - Escaped quotes: 'It''s'
    // - JSON: '{"a": 1, "b": 2}'
}
```

#### 6.1.3 Helper: Execute SELECT for INSERT

```go
// previewInsertSelect extracts and executes the SELECT for INSERT ... SELECT
func (s *QueryService) previewInsertSelect(
    ctx context.Context,
    queryText string,
    tableName string,
    schema *TableSchema,
    dataSource *models.DataSource,
) (*InsertPreviewResult, error) {
    // 1. Extract SELECT clause
    // INSERT INTO table SELECT ... → extract SELECT ...
    selectMatch := regexp.MustCompile(`(?i)INSERT\s+INTO\s+\w+\s*(?:\([^)]+\))?\s*(SELECT\s+.+)`)
    matches := selectMatch.FindStringSubmatch(queryText)
    
    if len(matches) < 2 {
        return nil, fmt.Errorf("failed to extract SELECT clause")
    }
    
    selectQuery := matches[1]
    
    // 2. Add LIMIT 50 if not present
    if !strings.Contains(strings.ToUpper(selectQuery), "LIMIT") {
        selectQuery += " LIMIT 50"
    }
    
    // 3. Execute SELECT
    dataSourceDB, err := s.connectToDataSource(dataSource)
    if err != nil {
        return nil, err
    }
    
    // 4. Get total count
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_table", selectQuery)
    var totalCount int
    dataSourceDB.Raw(countQuery).Scan(&totalCount)
    
    // 5. Execute limited SELECT
    rows, err := dataSourceDB.Raw(selectQuery).Rows()
    if err != nil {
        return nil, fmt.Errorf("failed to execute SELECT: %w", err)
    }
    defer rows.Close()
    
    // 6. Convert to preview format
    previewRows := scanRowsToMap(rows)
    
    return &InsertPreviewResult{
        TableName: tableName,
        Columns: schema.Columns,
        Rows: previewRows,
        TotalRowCount: totalCount,
        PreviewType: InsertPreviewTypeSelect,
        SelectQuery: selectQuery,
    }, nil
}
```

### 6.2 Frontend Implementation

#### 6.2.1 API Client Method

**Location:** `web/src/lib/api-client.ts`

```typescript
async previewInsertQuery(
  dataSourceId: string,
  queryText: string
): Promise<InsertPreviewResult> {
  const response = await this.client.post('/api/v1/queries/preview-insert', {
    data_source_id: dataSourceId,
    query_text: queryText,
  });
  return response.data;
}
```

#### 6.2.2 New Component: InsertPreviewPanel

**Location:** `web/src/components/approvals/InsertPreviewPanel.tsx`

```typescript
interface InsertPreviewPanelProps {
  preview: InsertPreviewResult;
  onProceed: () => void;
  onCancel: () => void;
}

export default function InsertPreviewPanel({
  preview,
  onProceed,
  onCancel,
}: InsertPreviewPanelProps) {
  return (
    <div className="insert-preview-panel">
      {/* Header */}
      <div className="preview-header">
        <Badge color="blue">INSERT</Badge>
        <span className="table-name">INTO {preview.table_name}</span>
        <span className="row-count">
          {preview.total_row_count} row{preview.total_row_count !== 1 ? 's' : ''} to insert
        </span>
      </div>
      
      {/* Preview Type Indicator */}
      {preview.preview_type === 'select' && (
        <Alert type="info">
          Preview from SELECT query. Showing up to 50 rows.
        </Alert>
      )}
      
      {/* Data Table */}
      <div className="preview-table-container">
        <table className="preview-table">
          <thead>
            <tr>
              <th>#</th>
              {preview.columns.map(col => (
                <th key={col.name}>
                  {col.name}
                  <span className="column-type">{col.type}</span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {preview.rows.map((row, idx) => (
              <tr key={idx}>
                <td className="row-number">{idx + 1}</td>
                {preview.columns.map(col => (
                  <td key={col.name}>
                    {formatCellValue(row[col.name])}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        
        {preview.rows.length < preview.total_row_count && (
          <div className="row-limit-notice">
            Showing {preview.rows.length} of {preview.total_row_count} rows
          </div>
        )}
      </div>
      
      {/* Actions */}
      <div className="preview-actions">
        <Button onClick={onCancel} variant="outline">Cancel</Button>
        <Button onClick={onProceed} variant="primary">
          Proceed to Transaction
        </Button>
      </div>
    </div>
  );
}

function formatCellValue(value: unknown): string {
  if (value === null) return 'NULL';
  if (value === undefined) return '';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}
```

#### 6.2.3 Update ApprovalDetail

**Location:** `web/src/components/approvals/ApprovalDetail.tsx`

```typescript
const handlePreviewQuery = async () => {
  if (!approval) return;
  setPhase('loading_preview');
  setError(null);
  
  try {
    const operationType = approval.operation_type?.toUpperCase();
    
    if (isMultiQuery(approval.query_text)) {
      // Multi-query handling (existing)
      const preview = await previewMultiQuery(approval.data_source_id, [approval.query_text]);
      setMultiQueryPreview(preview);
      setPhase('preview_ready');
    } else if (operationType === 'INSERT') {
      // NEW: INSERT preview
      const preview = await apiClient.previewInsertQuery(
        approval.data_source_id,
        approval.query_text
      );
      setInsertPreview(preview); // NEW state
      setPhase('preview_ready');
    } else if (operationType === 'UPDATE' || operationType === 'DELETE') {
      // Existing UPDATE/DELETE preview
      const preview = await apiClient.previewWriteQuery(
        approval.data_source_id,
        approval.query_text
      );
      setWritePreview(preview);
      setPhase('preview_ready');
    } else {
      throw new Error(`Unsupported operation type: ${operationType}`);
    }
  } catch (err) {
    toast.error(err instanceof Error ? err.message : 'Failed to fetch preview');
    setError(err instanceof Error ? err.message : 'Failed to fetch preview');
    setPhase('idle');
  }
};

// In render method:
{phase === 'preview_ready' && insertPreview && (
  <InsertPreviewPanel
    preview={insertPreview}
    onProceed={handleStartTransaction}
    onCancel={() => {
      setPhase('idle');
      setInsertPreview(null);
    }}
  />
)}
```

---

## 7. Edge Cases & Error Handling

### 7.1 Parsing Edge Cases

| Case | Example | Handling |
|------|---------|----------|
| Escaped quotes | `'O''Brien'` | Parser handles `''` as single quote |
| JSON in values | `'{"key": "val"}'` | Preserved as string, displayed as JSON |
| Multi-line VALUES | `VALUES\n(a),\n(b)` | Normalized to single line before parsing |
| No columns specified | `INSERT INTO t VALUES (1, 2)` | Fetch table schema to get column names |
| Subquery in VALUES | `VALUES ((SELECT MAX(id)+1 FROM t), 'name')` | Display as-is, mark as expression |
| NULL values | `VALUES (1, NULL, 'test')` | Display as `NULL` in preview |

### 7.2 Error Scenarios

| Scenario | Error Type | User Message |
|----------|-----------|--------------|
| Invalid SQL syntax | ParseError | "Failed to parse INSERT statement. Please check your SQL syntax." |
| Table doesn't exist | SchemaError | "Table 'users' does not exist or you don't have access." |
| SELECT execution fails | ExecutionError | "Failed to execute SELECT: [database error message]" |
| Permission denied | PermissionError | "You don't have INSERT permission on table 'users'." |
| Values exceed column count | ValidationError | "INSERT has more values than columns specified." |

### 7.3 Large Dataset Handling

**VALUES with 1000+ rows:**
- Parse all rows to get total count
- Return only first 50 in preview
- Display: "Showing 50 of 1,234 rows to insert"

**SELECT returning 10,000+ rows:**
- Execute `SELECT COUNT(*)` to get total
- Execute `SELECT ... LIMIT 50` for preview
- Display: "SELECT will insert 10,234 rows. Showing first 50."

---

## 8. Security Considerations

### 8.1 SQL Injection Prevention

**VALUES Parsing:**
- Parse values, don't execute as SQL
- Treat values as strings, don't interpolate into queries
- Use regex for pattern matching, not string concatenation

**SELECT Execution:**
- Use database driver's parameterization
- Don't modify user-provided SELECT (except adding LIMIT)
- Validate table names against schema

### 8.2 Permission Model

**Required Permissions:**
- `can_insert` on target table
- `can_select` on source tables (for INSERT...SELECT)

**Enforcement:**
- Check permissions before executing SELECT
- Return 403 if permissions insufficient
- Log permission checks for audit

---

## 9. Testing Strategy

### 9.1 Unit Tests (Backend)

**Parse VALUES Tests:**
```go
func TestParseInsertValues_SingleRow(t *testing.T)
func TestParseInsertValues_MultiRow(t *testing.T)
func TestParseInsertValues_NoColumns(t *testing.T)
func TestParseInsertValues_EscapedQuotes(t *testing.T)
func TestParseInsertValues_JSONData(t *testing.T)
func TestParseInsertValues_NULLValues(t *testing.T)
func TestParseInsertValues_MixedTypes(t *testing.T)
```

**Extract SELECT Tests:**
```go
func TestExtractSelectFromInsert_Basic(t *testing.T)
func TestExtractSelectFromInsert_WithColumns(t *testing.T)
func TestExtractSelectFromInsert_ComplexWhere(t *testing.T)
func TestExtractSelectFromInsert_Joins(t *testing.T)
```

**Integration Tests:**
```go
func TestPreviewInsertQuery_VALUES(t *testing.T)
func TestPreviewInsertQuery_SELECT(t *testing.T)
func TestPreviewInsertQuery_ZeroRows(t *testing.T)
func TestPreviewInsertQuery_FiftyRowLimit(t *testing.T)
func TestPreviewInsertQuery_PermissionDenied(t *testing.T)
```

### 9.2 Frontend Tests

**Component Tests:**
```typescript
describe('InsertPreviewPanel', () => {
  it('renders VALUES preview correctly');
  it('renders SELECT preview correctly');
  it('shows row count indicator');
  it('handles empty result (0 rows)');
  it('formats NULL values');
  it('formats JSON values');
  it('truncates long values');
});
```

**Integration Tests:**
```typescript
describe('ApprovalDetail INSERT Flow', () => {
  it('shows INSERT preview for VALUES clause');
  it('shows INSERT preview for SELECT clause');
  it('handles preview errors gracefully');
  it('proceeds to transaction after preview');
});
```

### 9.3 E2E Tests

```typescript
describe('INSERT Approval Workflow', () => {
  it('approver can preview and approve INSERT with VALUES');
  it('approver can preview and approve INSERT with SELECT');
  it('preview shows correct row count');
  it('preview respects 50 row limit');
});
```

---

## 10. API Endpoints

### 10.1 New Endpoint

**POST /api/v1/queries/preview-insert**

Preview data to be inserted by an INSERT query.

**Request:**
```json
{
  "data_source_id": "uuid",
  "query_text": "INSERT INTO users (name) VALUES ('John')"
}
```

**Response 200 (VALUES):**
```json
{
  "table_name": "users",
  "columns": [{"name": "name", "type": "varchar", "nullable": false}],
  "rows": [{"name": "John"}],
  "total_row_count": 1,
  "preview_type": "values"
}
```

**Response 200 (SELECT):**
```json
{
  "table_name": "audit_log",
  "columns": [...],
  "rows": [...],
  "total_row_count": 150,
  "preview_type": "select",
  "select_query": "SELECT * FROM temp_logs LIMIT 50"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid SQL syntax
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Table doesn't exist
- `422 Unprocessable Entity` - Not an INSERT query

---

## 11. Migration & Rollout

### 11.1 Database Changes
None required. This is a code-only change.

### 11.2 Deployment Steps

1. **Deploy Backend:**
   - Deploy new QueryService methods
   - Deploy new API endpoint
   - Verify endpoint is accessible

2. **Deploy Frontend:**
   - Deploy InsertPreviewPanel component
   - Deploy ApprovalDetail updates
   - Verify UI renders correctly

3. **Verification:**
   - Test VALUES preview with single row
   - Test VALUES preview with multiple rows
   - Test SELECT preview
   - Test edge cases (NULL, JSON, quotes)
   - Test permission enforcement

### 11.3 Rollback Plan

If issues occur:
1. Revert to previous deployment
2. Frontend will show error for INSERT previews (existing behavior)
3. No database changes to rollback

---

## 12. Documentation Updates

### 12.1 Files to Update

1. **docs/user-guide/query-features.md**
   - Add section on INSERT preview
   - Document VALUES vs SELECT preview
   - Add examples

2. **docs/api/endpoints.md**
   - Document `/api/v1/queries/preview-insert`
   - Add request/response examples

3. **docs/user-guide/approval-workflow.md**
   - Update to mention INSERT preview
   - Add screenshots of preview panel

4. **CHANGELOG.md**
   - Add entry for INSERT preview feature

### 12.2 Code Documentation

- Add godoc comments to all new Go functions
- Add JSDoc comments to TypeScript interfaces
- Add inline comments for complex parsing logic

---

## 13. Performance Considerations

### 13.1 Parsing Performance

**VALUES Parsing:**
- O(n) complexity where n = query length
- No database round-trip
- Expected: < 50ms for queries with 1000+ rows

**SELECT Execution:**
- Two queries: COUNT(*) + SELECT LIMIT 50
- Database round-trip required
- Expected: < 1s total

### 13.2 Optimization Strategies

1. **Connection Pooling:** Reuse data source connections
2. **Lazy Schema Loading:** Only fetch schema if columns not specified
3. **Streaming Parse:** For very large VALUES, stream parse without loading all into memory

---

## 14. Future Enhancements (Out of Scope)

These are ideas for future iterations, NOT part of this implementation:

1. **Batch INSERT Preview:** Show summary statistics for very large inserts (>1000 rows)
2. **Foreign Key Validation:** Show warnings if inserting would violate FK constraints
3. **Duplicate Detection:** Highlight rows that would duplicate existing data
4. **Data Type Validation:** Warn if values don't match column types
5. **Generated Column Preview:** Show what generated columns would contain

---

## 15. Success Metrics

### 15.1 Adoption
- [ ] 100% of INSERT approvals use preview before proceeding

### 15.2 Performance
- [ ] VALUES preview < 100ms (p95)
- [ ] SELECT preview < 2s (p95)

### 15.3 Quality
- [ ] Zero SQL injection vulnerabilities
- [ ] 100% test coverage for parsing logic
- [ ] Zero production errors in first 30 days

### 15.4 User Satisfaction
- [ ] Approvers report confidence in understanding INSERT queries
- [ ] No requests for additional INSERT preview features

---

## 16. Approval

**Design Review:**
- [ ] Technical review completed
- [ ] Security review completed
- [ ] UX review completed

**Sign-offs:**
- [ ] Tech Lead: _________________ Date: _______
- [ ] Security: _________________ Date: _______
- [ ] Product: _________________ Date: _______

**Implementation Ready:**
- [ ] Design approved
- [ ] Implementation plan created
- [ ] Development can begin

---

## Appendices

### Appendix A: SQL Parsing Examples

**Example 1: Simple VALUES**
```sql
INSERT INTO users (name, email) VALUES ('John', 'john@example.com')
```
Parsed:
- Table: `users`
- Columns: [`name`, `email`]
- Rows: [[`John`, `john@example.com`]]

**Example 2: Multi-row VALUES**
```sql
INSERT INTO products (name, price) VALUES 
  ('Widget', 9.99),
  ('Gadget', 19.99),
  ('Doohickey', 29.99)
```
Parsed:
- Table: `products`
- Columns: [`name`, `price`]
- Rows: [[`Widget`, `9.99`], [`Gadget`, `19.99`], [`Doohickey`, `29.99`]]

**Example 3: VALUES with special characters**
```sql
INSERT INTO logs (message) VALUES ('Error: O''Brien''s account'), ('{"type": "login"}')
```
Parsed:
- Table: `logs`
- Columns: [`message`]
- Rows: [[`Error: O'Brien's account`], [`{"type": "login"}`]]

**Example 4: INSERT...SELECT**
```sql
INSERT INTO audit_archive SELECT * FROM audit_log WHERE created_at < '2023-01-01'
```
Extracted SELECT:
```sql
SELECT * FROM audit_log WHERE created_at < '2023-01-01' LIMIT 50
```

### Appendix B: Error Message Reference

| Code | Message | When |
|------|---------|------|
| `INSERT_PREVIEW_INVALID_SYNTAX` | "Failed to parse INSERT statement: {details}" | SQL syntax error |
| `INSERT_PREVIEW_TABLE_NOT_FOUND` | "Table '{table}' not found" | Table doesn't exist |
| `INSERT_PREVIEW_PERMISSION_DENIED` | "Permission denied: cannot INSERT into '{table}'" | No INSERT permission |
| `INSERT_PREVIEW_SELECT_FAILED` | "SELECT execution failed: {details}" | SELECT query error |
| `INSERT_PREVIEW_UNSUPPORTED` | "INSERT type not supported" | Complex INSERT patterns |

### Appendix C: UI Mockup

```
┌─────────────────────────────────────────────────────────────────┐
│  INSERT Preview                                      [X]        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  [INSERT] INTO users                                     │   │
│  │  3 rows to insert                                        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌─────┬──────────┬─────────────────┐                          │
│  │  #  │ name     │ email           │                          │
│  ├─────┼──────────┼─────────────────┤                          │
│  │  1  │ John Doe │ john@example.com│                          │
│  │  2  │ Jane Doe │ jane@example.com│                          │
│  │  3  │ Bob Smith│ bob@example.com │                          │
│  └─────┴──────────┴─────────────────┘                          │
│                                                                  │
│  ✓ All 3 rows shown                                             │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  [Cancel]              [Proceed to Transaction]         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

**End of Design Document**
