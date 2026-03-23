# INSERT Query Preview Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement INSERT query preview feature for approval workflow that shows approvers what data will be inserted (up to 50 rows) before they approve.

**Architecture:** Hybrid approach - parse VALUES clauses directly (no DB execution) or execute SELECT clauses with LIMIT 50. Backend validates and generates preview, frontend displays in new InsertPreviewPanel component.

**Tech Stack:** Go (Gin/GORM), TypeScript/React (Next.js), PostgreSQL/MySQL

**Based on Spec:** `docs/superpowers/specs/2026-03-23-insert-preview-design.md`

---

## File Structure Overview

| File | Responsibility | Type |
|------|----------------|------|
| `internal/service/query.go` | Add PreviewInsertQuery method and helpers | Modify |
| `internal/api/handlers/query.go` | Add preview-insert endpoint handler | Modify |
| `internal/api/dto/query.go` | Add InsertPreviewResult DTO | Modify |
| `internal/api/routes.go` | Register preview-insert endpoint | Modify |
| `web/src/lib/api-client.ts` | Add previewInsertQuery API method | Modify |
| `web/src/lib/api/insert-preview.ts` | TypeScript types for insert preview | Create |
| `web/src/components/approvals/InsertPreviewPanel.tsx` | UI component for INSERT preview | Create |
| `web/src/components/approvals/ApprovalDetail.tsx` | Route to INSERT preview for INSERT ops | Modify |

---

## Chunk 1: Backend Core - VALUES Parser

### Task 1.1: Add MaxInsertPreviewRows constant

**Files:**
- Modify: `internal/service/query.go`

- [ ] **Step 1: Add constant at top of file**

Add after imports:
```go
// MaxInsertPreviewRows limits the number of rows shown in INSERT preview
const MaxInsertPreviewRows = 50
```

- [ ] **Step 2: Commit**

```bash
git add internal/service/query.go
git commit -m "feat: add MaxInsertPreviewRows constant for INSERT preview"
```

---

### Task 1.2: Create InsertParser struct and parseInsertValues method

**Files:**
- Modify: `internal/service/query.go`

- [ ] **Step 1: Add InsertParser struct after PreviewResult struct**

Find PreviewResult struct (around line 815), add after it:

```go
// InsertParser handles parsing of INSERT statements
type InsertParser struct{}

// InsertValuesResult holds parsed column names and value rows
type InsertValuesResult struct {
	Columns []string
	Rows    [][]string
}

// parseInsertValues extracts column names and value rows from INSERT ... VALUES
func (p *InsertParser) parseInsertValues(queryText string) (*InsertValuesResult, error) {
	// Regex pattern for INSERT INTO table [(cols)] VALUES (vals), (vals), ...
	pattern := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(?:"?([^"]+)"?|\w+)\s*(?:\(([^)]+)\))?\s*VALUES\s+(.+)`)
	matches := pattern.FindStringSubmatch(queryText)

	if len(matches) < 4 {
		return nil, fmt.Errorf("failed to parse INSERT statement: does not match expected pattern")
	}

	// Parse column names
	var columns []string
	if matches[2] != "" {
		columns = parseColumnList(matches[2])
	}

	// Parse value rows
	valuesSection := matches[3]
	rows, err := p.parseValueRows(valuesSection)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VALUES: %w", err)
	}

	return &InsertValuesResult{
		Columns: columns,
		Rows:    rows,
	}, nil
}
```

- [ ] **Step 2: Add parseColumnList helper function**

Add after parseInsertValues:

```go
// parseColumnList splits a comma-separated column list
func parseColumnList(columnsStr string) []string {
	columns := strings.Split(columnsStr, ",")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
		// Remove quotes if present
		columns[i] = strings.Trim(columns[i], `"`)
		columns[i] = strings.Trim(columns[i], "`")
	}
	return columns
}
```

- [ ] **Step 3: Add parseValueRows method to InsertParser**

Add to InsertParser:

```go
// parseValueRows parses comma-separated value tuples
func (p *InsertParser) parseValueRows(valuesSection string) ([][]string, error) {
	var rows [][]string
	var currentRow []string
	var currentValue strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	parenDepth := 0

	for i, ch := range valuesSection {
		switch {
		case !inQuotes && (ch == '\'' || ch == '"'):
			// Start of quoted string
			inQuotes = true
			quoteChar = ch

		case inQuotes && ch == quoteChar:
			// Check if escaped (double quote)
			if i+1 < len(valuesSection) && rune(valuesSection[i+1]) == quoteChar {
				currentValue.WriteRune(ch) // Add single quote
				// Skip next char (it's the escape)
			} else {
				// End of quoted string
				inQuotes = false
				quoteChar = 0
			}

		case inQuotes:
			// Inside quoted string, just accumulate
			currentValue.WriteRune(ch)

		case ch == '(' && parenDepth == 0:
			// Start of a row tuple
			parenDepth++

		case ch == '(':
			// Nested parenthesis
			parenDepth++
			currentValue.WriteRune(ch)

		case ch == ')' && parenDepth == 1:
			// End of row tuple
			parenDepth--
			// Save current value
			if currentValue.Len() > 0 {
				currentRow = append(currentRow, strings.TrimSpace(currentValue.String()))
				currentValue.Reset()
			}
			// Save row
			if len(currentRow) > 0 {
				rows = append(rows, currentRow)
				currentRow = nil
			}

		case ch == ')' && parenDepth > 1:
			// End of nested parenthesis
			parenDepth--
			currentValue.WriteRune(ch)

		case ch == ',' && parenDepth == 1:
			// Comma between values in a row
			if currentValue.Len() > 0 {
				currentRow = append(currentRow, strings.TrimSpace(currentValue.String()))
				currentValue.Reset()
			}

		case ch == ',' && parenDepth == 0:
			// Comma between rows
			continue

		default:
			// Regular character
			if !(ch == ' ' && currentValue.Len() == 0) {
				currentValue.WriteRune(ch)
			}
		}
	}

	return rows, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/service/query.go
git commit -m "feat: add InsertParser with parseInsertValues and parseValueRows"
```

---

### Task 1.3: Write unit tests for VALUES parser

**Files:**
- Create: `internal/service/insert_parser_test.go`

- [ ] **Step 1: Create test file with basic tests**

```go
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInsertValues_SingleRow(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Equal(t, []string{"name", "email"}, result.Columns)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, "'John'", result.Rows[0][0])
	assert.Equal(t, "'john@example.com'", result.Rows[0][1])
}

func TestParseInsertValues_MultiRow(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name) VALUES ('John'), ('Jane'), ('Bob')"
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Len(t, result.Rows, 3)
	assert.Equal(t, "'John'", result.Rows[0][0])
	assert.Equal(t, "'Jane'", result.Rows[1][0])
	assert.Equal(t, "'Bob'", result.Rows[2][0])
}

func TestParseInsertValues_NoColumns(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users VALUES (1, 'John', 'john@example.com')"
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Nil(t, result.Columns) // Columns not specified
	assert.Len(t, result.Rows[0], 3)
}

func TestParseInsertValues_EscapedQuotes(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name) VALUES ('O''Brien')"
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Equal(t, "'O''Brien'", result.Rows[0][0])
}

func TestParseInsertValues_JSONData(t *testing.T) {
	parser := &InsertParser{}
	query := `INSERT INTO logs (data) VALUES ('{"type": "login", "user_id": 123}')`
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Contains(t, result.Rows[0][0], `"type"`)
	assert.Contains(t, result.Rows[0][0], `"user_id"`)
}

func TestParseInsertValues_NULLValues(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name, email) VALUES ('John', NULL)"
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Equal(t, "NULL", result.Rows[0][1])
}

func TestParseInsertValues_MixedTypes(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO products (name, price, active) VALUES ('Widget', 9.99, true)"
	
	result, err := parser.parseInsertValues(query)
	
	assert.NoError(t, err)
	assert.Equal(t, "'Widget'", result.Rows[0][0])
	assert.Equal(t, "9.99", result.Rows[0][1])
	assert.Equal(t, "true", result.Rows[0][2])
}

func TestParseInsertValues_InvalidSyntax(t *testing.T) {
	parser := &InsertParser{}
	query := "SELECT * FROM users"
	
	_, err := parser.parseInsertValues(query)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}
```

- [ ] **Step 2: Run tests to verify they pass**

```bash
cd /Users/gatotsayogya/Project/querybase
go test -v ./internal/service/... -run TestParseInsertValues
```

Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add internal/service/insert_parser_test.go
git commit -m "test: add unit tests for InsertParser VALUES parsing"
```

---

## Chunk 2: Backend Core - SELECT Parser and Preview Methods

### Task 2.1: Add helpers for SELECT-based INSERT

**Files:**
- Modify: `internal/service/query.go`

- [ ] **Step 1: Add detectInsertType function**

Add after InsertParser methods:

```go
// InsertPreviewType represents the type of INSERT preview
type InsertPreviewType string

const (
	InsertPreviewTypeValues InsertPreviewType = "values"
	InsertPreviewTypeSelect InsertPreviewType = "select"
)

// detectInsertType determines if INSERT uses VALUES or SELECT
func detectInsertType(queryText string) InsertPreviewType {
	upperQuery := strings.ToUpper(queryText)
	
	// Check for SELECT keyword after VALUES section or at end
	// VALUES clause ends with ), pattern: VALUES (...)
	// SELECT clause: INSERT INTO ... SELECT ...
	if strings.Contains(upperQuery, "SELECT") && 
	   !strings.Contains(upperQuery, "VALUES") {
		return InsertPreviewTypeSelect
	}
	
	// Check if SELECT comes after a closing paren (end of VALUES)
	// This is a VALUES with a subquery, not INSERT...SELECT
	selectIndex := strings.Index(upperQuery, "SELECT")
	valuesIndex := strings.Index(upperQuery, "VALUES")
	
	if selectIndex > valuesIndex && valuesIndex != -1 {
		// SELECT after VALUES - might be subquery in VALUES
		// Check if there's a closing paren before SELECT
		between := upperQuery[valuesIndex:selectIndex]
		if strings.Contains(between, "(") && strings.Contains(between, ")") {
			// Likely VALUES with subquery, not INSERT...SELECT
			return InsertPreviewTypeValues
		}
	}
	
	if selectIndex != -1 && (valuesIndex == -1 || selectIndex < valuesIndex) {
		return InsertPreviewTypeSelect
	}
	
	return InsertPreviewTypeValues
}
```

- [ ] **Step 2: Add extractSelectFromInsert function**

```go
// extractSelectFromInsert extracts the SELECT clause from INSERT...SELECT
func extractSelectFromInsert(queryText string) (string, error) {
	// Pattern: INSERT INTO ... [optional columns] SELECT ...
	upperQuery := strings.ToUpper(queryText)
	selectIndex := strings.Index(upperQuery, "SELECT")
	
	if selectIndex == -1 {
		return "", fmt.Errorf("no SELECT clause found in INSERT statement")
	}
	
	selectQuery := queryText[selectIndex:]
	return strings.TrimSpace(selectQuery), nil
}
```

- [ ] **Step 3: Add extractInsertTableName function**

```go
// extractInsertTableName extracts the target table name from INSERT
func extractInsertTableName(queryText string) string {
	// Pattern: INSERT INTO table_name [...]
	re := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(?:"?([^"\s]+)"?|\w+)`)
	matches := re.FindStringSubmatch(queryText)
	
	if len(matches) > 1 && matches[1] != "" {
		return matches[1]
	}
	
	return ""
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/service/query.go
git commit -m "feat: add INSERT type detection and extraction helpers"
```

---

### Task 2.2: Add PreviewInsertQuery method

**Files:**
- Modify: `internal/service/query.go`

- [ ] **Step 1: Add InsertPreviewResult type**

After PreviewResult struct:

```go
// InsertPreviewResult represents the preview for an INSERT query
type InsertPreviewResult struct {
	TableName     string                   `json:"table_name"`
	Columns       []ColumnInfo             `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	TotalRowCount int                      `json:"total_row_count"`
	PreviewType   InsertPreviewType        `json:"preview_type"`
	SelectQuery   string                   `json:"select_query,omitempty"`
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}
```

- [ ] **Step 2: Add PreviewInsertQuery method**

```go
// PreviewInsertQuery previews the data that will be inserted by an INSERT query
// without actually modifying any data.
// For VALUES: parses the SQL to extract values
// For SELECT: executes the SELECT with LIMIT 50
func (s *QueryService) PreviewInsertQuery(
	ctx context.Context,
	queryText string,
	dataSource *models.DataSource,
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
	if tableName == "" {
		return nil, fmt.Errorf("failed to extract table name from INSERT statement")
	}

	// 4. Get table schema
	schema, err := s.getTableSchema(ctx, dataSource, tableName)
	if err != nil {
		// Continue without schema, will use parsed column names
		schema = &TableSchema{}
	}

	// 5. Generate preview based on type
	switch insertType {
	case InsertPreviewTypeValues:
		return s.previewInsertValues(queryText, tableName, schema)
	case InsertPreviewTypeSelect:
		return s.previewInsertSelect(ctx, queryText, tableName, schema, dataSource)
	default:
		return nil, fmt.Errorf("unsupported INSERT type")
	}
}
```

- [ ] **Step 3: Add previewInsertValues method**

```go
// previewInsertValues handles preview for INSERT...VALUES
func (s *QueryService) previewInsertValues(
	queryText string,
	tableName string,
	schema *TableSchema,
) (*InsertPreviewResult, error) {
	parser := &InsertParser{}
	
	parsed, err := parser.parseInsertValues(queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VALUES: %w", err)
	}

	// Use parsed columns or fallback to schema columns
	columns := parsed.Columns
	if len(columns) == 0 && len(schema.Columns) > 0 {
		columns = make([]string, len(schema.Columns))
		for i, col := range schema.Columns {
			columns[i] = col.Name
		}
	}

	// Convert columns to ColumnInfo
	columnInfos := make([]ColumnInfo, len(columns))
	for i, colName := range columns {
		columnInfos[i] = ColumnInfo{Name: colName, Type: "unknown", Nullable: true}
	}

	// Convert parsed rows to map format
	var rows []map[string]interface{}
	rowCount := len(parsed.Rows)
	if rowCount > MaxInsertPreviewRows {
		rowCount = MaxInsertPreviewRows
	}
	
	for i := 0; i < rowCount; i++ {
		row := make(map[string]interface{})
		for j, colName := range columns {
			if j < len(parsed.Rows[i]) {
				value := parsed.Rows[i][j]
				// Try to parse as different types
				row[colName] = parseValue(value)
			} else {
				row[colName] = nil
			}
		}
		rows = append(rows, row)
	}

	return &InsertPreviewResult{
		TableName:     tableName,
		Columns:       columnInfos,
		Rows:          rows,
		TotalRowCount: len(parsed.Rows),
		PreviewType:   InsertPreviewTypeValues,
	}, nil
}
```

- [ ] **Step 4: Add parseValue helper**

```go
// parseValue attempts to parse a SQL value string into appropriate Go type
func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)
	
	// Handle NULL
	if strings.ToUpper(value) == "NULL" {
		return nil
	}
	
	// Handle quoted strings
	if (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) ||
	   (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
		// Remove quotes and handle escaped quotes
		unquoted := value[1 : len(value)-1]
		unquoted = strings.ReplaceAll(unquoted, "''", "'")
		unquoted = strings.ReplaceAll(unquoted, "\"\"", "\"")
		return unquoted
	}
	
	// Try integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}
	
	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}
	
	// Try boolean
	lowerVal := strings.ToLower(value)
	if lowerVal == "true" {
		return true
	}
	if lowerVal == "false" {
		return false
	}
	
	// Return as string
	return value
}
```

- [ ] **Step 5: Add TableSchema type and getTableSchema method**

```go
// TableSchema represents table structure
type TableSchema struct {
	Columns []ColumnInfo
}

// getTableSchema fetches schema for a table
func (s *QueryService) getTableSchema(
	ctx context.Context,
	dataSource *models.DataSource,
	tableName string,
) (*TableSchema, error) {
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return nil, err
	}

	// Query information schema (PostgreSQL)
	var columns []ColumnInfo
	query := `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = ?
		ORDER BY ordinal_position
	`
	
	rows, err := dataSourceDB.Raw(query, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var nullable string
		if err := rows.Scan(&col.Name, &col.Type, &nullable); err != nil {
			continue
		}
		col.Nullable = (nullable == "YES")
		columns = append(columns, col)
	}

	return &TableSchema{Columns: columns}, nil
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/service/query.go
git commit -m "feat: add PreviewInsertQuery and previewInsertValues methods"
```

---

### Task 2.3: Add previewInsertSelect method

**Files:**
- Modify: `internal/service/query.go`

- [ ] **Step 1: Add previewInsertSelect method**

```go
// previewInsertSelect handles preview for INSERT...SELECT
func (s *QueryService) previewInsertSelect(
	ctx context.Context,
	queryText string,
	tableName string,
	schema *TableSchema,
	dataSource *models.DataSource,
) (*InsertPreviewResult, error) {
	// 1. Extract SELECT clause
	selectQuery, err := extractSelectFromInsert(queryText)
	if err != nil {
		return nil, err
	}

	// 2. Connect to data source
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// 3. Get total count (before adding LIMIT)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_table", selectQuery)
	var totalCount int
	if err := dataSourceDB.Raw(countQuery).Scan(&totalCount).Error; err != nil {
		// Continue even if count fails
		totalCount = -1
	}

	// 4. Add LIMIT for preview
	limitedSelectQuery := selectQuery
	if !strings.Contains(strings.ToUpper(selectQuery), "LIMIT") {
		limitedSelectQuery = fmt.Sprintf("%s LIMIT %d", selectQuery, MaxInsertPreviewRows)
	}

	// 5. Execute limited SELECT
	rows, err := dataSourceDB.Raw(limitedSelectQuery).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute SELECT: %w", err)
	}
	defer rows.Close()

	// 6. Get column names from result
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Convert to ColumnInfo
	columnInfos := make([]ColumnInfo, len(columns))
	for i, colName := range columns {
		columnInfos[i] = ColumnInfo{Name: colName, Type: "unknown", Nullable: true}
	}

	// 7. Scan rows
	var previewRows []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		previewRows = append(previewRows, row)
	}

	return &InsertPreviewResult{
		TableName:     tableName,
		Columns:       columnInfos,
		Rows:          previewRows,
		TotalRowCount: totalCount,
		PreviewType:   InsertPreviewTypeSelect,
		SelectQuery:   limitedSelectQuery,
	}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/service/query.go
git commit -m "feat: add previewInsertSelect method for INSERT...SELECT queries"
```

---

### Task 2.4: Write integration tests for preview methods

**Files:**
- Create: `internal/service/insert_preview_test.go`

- [ ] **Step 1: Create integration test file**

```go
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewInsertQuery_VALUES(t *testing.T) {
	// This is an integration test that requires a database
	// For now, test the parse logic only
	parser := &InsertParser{}
	query := "INSERT INTO users (name, email) VALUES ('John', 'john@test.com'), ('Jane', 'jane@test.com')"
	
	result, err := parser.parseInsertValues(query)
	
	require.NoError(t, err)
	assert.Equal(t, []string{"name", "email"}, result.Columns)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "'John'", result.Rows[0][0])
	assert.Equal(t, "'john@test.com'", result.Rows[0][1])
}

func TestPreviewInsertQuery_VALUES_Empty(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name) VALUES ('test')"
	
	result, err := parser.parseInsertValues(query)
	
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)
}

func TestDetectInsertType_VALUES(t *testing.T) {
	query := "INSERT INTO users (name) VALUES ('John')"
	result := detectInsertType(query)
	assert.Equal(t, InsertPreviewTypeValues, result)
}

func TestDetectInsertType_SELECT(t *testing.T) {
	query := "INSERT INTO archive SELECT * FROM logs WHERE created_at < '2024-01-01'"
	result := detectInsertType(query)
	assert.Equal(t, InsertPreviewTypeSelect, result)
}

func TestExtractSelectFromInsert(t *testing.T) {
	query := "INSERT INTO audit_log SELECT action, user_id FROM events WHERE status = 'pending'"
	selectQuery, err := extractSelectFromInsert(query)
	
	require.NoError(t, err)
	assert.Equal(t, "SELECT action, user_id FROM events WHERE status = 'pending'", selectQuery)
}

func TestExtractInsertTableName(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"INSERT INTO users (name) VALUES ('John')", "users"},
		{"INSERT INTO \"Users\" VALUES (1)", "Users"},
		{"INSERT INTO public.users VALUES (1)", "public.users"},
	}
	
	for _, tt := range tests {
		result := extractInsertTableName(tt.query)
		assert.Equal(t, tt.expected, result)
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"NULL", nil},
		{"'hello'", "hello"},
		{"\"world\"", "world"},
		{"123", int64(123)},
		{"45.67", 45.67},
		{"true", true},
		{"false", false},
		{"'O''Brien'", "O'Brien"},
		{"test", "test"},
	}
	
	for _, tt := range tests {
		result := parseValue(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd /Users/gatotsayogya/Project/querybase
go test -v ./internal/service/... -run TestPreviewInsert
```

Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add internal/service/insert_preview_test.go
git commit -m "test: add integration tests for INSERT preview functionality"
```

---

## Chunk 3: Backend API - Endpoint and Handler

### Task 3.1: Add DTO types

**Files:**
- Modify: `internal/api/dto/query.go`

- [ ] **Step 1: Add InsertPreviewRequest and InsertPreviewResponse**

Add at end of file:

```go
// InsertPreviewRequest represents a request to preview an INSERT query
type InsertPreviewRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
}

// InsertPreviewResponse represents the response for INSERT preview
type InsertPreviewResponse struct {
	TableName     string                 `json:"table_name"`
	Columns       []ColumnInfo           `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	TotalRowCount int                    `json:"total_row_count"`
	PreviewType   string                 `json:"preview_type"`
	SelectQuery   string                 `json:"select_query,omitempty"`
}

// ColumnInfo represents column metadata in the response
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/api/dto/query.go
git commit -m "feat: add InsertPreviewRequest and InsertPreviewResponse DTOs"
```

---

### Task 3.2: Add API handler

**Files:**
- Modify: `internal/api/handlers/query.go`

- [ ] **Step 1: Add PreviewInsertQuery handler method**

Add to QueryHandler:

```go
// PreviewInsertQuery handles preview requests for INSERT queries
func (h *QueryHandler) PreviewInsertQuery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.InsertPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse data source ID
	dataSourceID, err := uuid.Parse(req.DataSourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data source ID"})
		return
	}

	// Get data source
	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		return
	}

	// Generate preview
	result, err := h.queryService.PreviewInsertQuery(c.Request.Context(), req.QueryText, &dataSource)
	if err != nil {
		h.logger.Error("Failed to preview INSERT query", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTO
	columns := make([]dto.ColumnInfo, len(result.Columns))
	for i, col := range result.Columns {
		columns[i] = dto.ColumnInfo{
			Name:     col.Name,
			Type:     col.Type,
			Nullable: col.Nullable,
		}
	}

	response := dto.InsertPreviewResponse{
		TableName:     result.TableName,
		Columns:       columns,
		Rows:          result.Rows,
		TotalRowCount: result.TotalRowCount,
		PreviewType:   string(result.PreviewType),
		SelectQuery:   result.SelectQuery,
	}

	c.JSON(http.StatusOK, response)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/api/handlers/query.go
git commit -m "feat: add PreviewInsertQuery handler for INSERT preview endpoint"
```

---

### Task 3.3: Register route

**Files:**
- Modify: `internal/api/routes.go`

- [ ] **Step 1: Find query routes and add preview-insert endpoint**

Find the query routes section (around line with queries.PreviewWriteQuery) and add:

```go
// In setupQueryRoutes or similar function
queryRoutes.POST("/preview-insert", queryHandler.PreviewInsertQuery)
```

- [ ] **Step 2: Commit**

```bash
git add internal/api/routes.go
git commit -m "feat: register preview-insert endpoint"
```

---

## Chunk 4: Frontend - API Client and Types

### Task 4.1: Add TypeScript types

**Files:**
- Create: `web/src/lib/api/insert-preview.ts`

- [ ] **Step 1: Create type definitions**

```typescript
export interface ColumnInfo {
  name: string;
  type: string;
  nullable: boolean;
}

export interface InsertPreviewResult {
  table_name: string;
  columns: ColumnInfo[];
  rows: Record<string, unknown>[];
  total_row_count: number;
  preview_type: 'values' | 'select';
  select_query?: string;
}

export interface InsertPreviewRequest {
  data_source_id: string;
  query_text: string;
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/lib/api/insert-preview.ts
git commit -m "feat: add TypeScript types for INSERT preview"
```

---

### Task 4.2: Add API client method

**Files:**
- Modify: `web/src/lib/api-client.ts`

- [ ] **Step 1: Add previewInsertQuery method**

Add to ApiClient class:

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

Also add import at top:
```typescript
import type { InsertPreviewResult } from './api/insert-preview';
```

- [ ] **Step 2: Commit**

```bash
git add web/src/lib/api-client.ts
git commit -m "feat: add previewInsertQuery API client method"
```

---

## Chunk 5: Frontend - UI Components

### Task 5.1: Create InsertPreviewPanel component

**Files:**
- Create: `web/src/components/approvals/InsertPreviewPanel.tsx`

- [ ] **Step 1: Create component with basic structure**

```typescript
import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Alert } from '@/components/ui/Alert';
import type { InsertPreviewResult } from '@/lib/api/insert-preview';

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
  const [expandedRows, setExpandedRows] = useState<Set<number>>(new Set());

  const formatCellValue = (value: unknown): string => {
    if (value === null) return 'NULL';
    if (value === undefined) return '';
    if (typeof value === 'object') {
      const str = JSON.stringify(value);
      return str.length > 100 ? str.substring(0, 100) + '...' : str;
    }
    const str = String(value);
    return str.length > 100 ? str.substring(0, 100) + '...' : str;
  };

  const toggleRowExpand = (index: number) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedRows(newExpanded);
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Badge variant="info">INSERT</Badge>
          <span className="text-lg font-semibold">INTO {preview.table_name}</span>
          <span className="text-muted-foreground">
            {preview.total_row_count} row{preview.total_row_count !== 1 ? 's' : ''} to insert
          </span>
        </div>
      </div>

      {/* SELECT indicator */}
      {preview.preview_type === 'select' && (
        <Alert variant="info">
          Preview from SELECT query. Showing up to 50 rows.
          {preview.total_row_count > 50 && (
            <span> Total: {preview.total_row_count} rows.</span>
          )}
        </Alert>
      )}

      {/* Empty state */}
      {preview.rows.length === 0 && (
        <Alert variant="warning">
          No rows to insert. The query will not insert any data.
        </Alert>
      )}

      {/* Data table */}
      {preview.rows.length > 0 && (
        <div className="border rounded-md overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-2 text-left font-medium w-12">#</th>
                {preview.columns.map((col) => (
                  <th key={col.name} className="px-4 py-2 text-left font-medium">
                    <div className="flex flex-col">
                      <span>{col.name}</span>
                      <span className="text-xs text-muted-foreground">{col.type}</span>
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {preview.rows.map((row, idx) => (
                <tr key={idx} className="border-t hover:bg-muted/50">
                  <td className="px-4 py-2 text-muted-foreground">{idx + 1}</td>
                  {preview.columns.map((col) => (
                    <td key={col.name} className="px-4 py-2 font-mono text-xs">
                      {formatCellValue(row[col.name])}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          
          {preview.rows.length < preview.total_row_count && (
            <div className="px-4 py-2 text-sm text-muted-foreground border-t bg-muted/30">
              Showing {preview.rows.length} of {preview.total_row_count} rows
            </div>
          )}
        </div>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={onProceed}>
          Proceed to Transaction
        </Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/approvals/InsertPreviewPanel.tsx
git commit -m "feat: add InsertPreviewPanel component for INSERT query preview"
```

---

### Task 5.2: Update ApprovalDetail to route INSERT previews

**Files:**
- Modify: `web/src/components/approvals/ApprovalDetail.tsx`

- [ ] **Step 1: Add state and imports**

Add import:
```typescript
import InsertPreviewPanel from './InsertPreviewPanel';
import type { InsertPreviewResult } from '@/lib/api/insert-preview';
```

Add state:
```typescript
const [insertPreview, setInsertPreview] = useState<InsertPreviewResult | null>(null);
```

- [ ] **Step 2: Update handlePreviewQuery to handle INSERT**

Update the handlePreviewQuery function:

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
      setInsertPreview(preview);
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
```

- [ ] **Step 3: Add render case for INSERT preview**

Find where the phase === 'preview_ready' renders and add:

```typescript
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

- [ ] **Step 4: Commit**

```bash
git add web/src/components/approvals/ApprovalDetail.tsx
git commit -m "feat: integrate InsertPreviewPanel into ApprovalDetail"
```

---

## Chunk 6: Build and Verification

### Task 6.1: Build backend

- [ ] **Step 1: Build Go backend**

```bash
cd /Users/gatotsayogya/Project/querybase
make build
```

Expected: Build succeeds with no errors

- [ ] **Step 2: Run backend tests**

```bash
make test-short
```

Expected: All tests pass

- [ ] **Step 3: Commit (if any fixes needed)**

```bash
# Only if changes made during build/test
git add .
git commit -m "fix: resolve build and test issues"
```

---

### Task 6.2: Build frontend

- [ ] **Step 1: Build Next.js frontend**

```bash
cd /Users/gatotsayogya/Project/querybase/web
npm run build
```

Expected: Build succeeds with no errors

- [ ] **Step 2: Run TypeScript check**

```bash
npx tsc --noEmit
```

Expected: No TypeScript errors

- [ ] **Step 3: Commit (if any fixes needed)**

```bash
# Only if changes made during build
git add .
git commit -m "fix: resolve frontend build issues"
```

---

### Task 6.3: Push all changes

- [ ] **Step 1: Push to git**

```bash
git push origin feature/insert-preview
```

- [ ] **Step 2: Push beads data**

```bash
bd dolt push
```

---

## Verification Checklist

After completing all tasks, verify:

- [ ] VALUES parser handles single and multi-row INSERTs
- [ ] VALUES parser handles escaped quotes and JSON
- [ ] SELECT-based INSERT preview executes with LIMIT 50
- [ ] API endpoint returns correct response format
- [ ] Frontend displays preview table with headers
- [ ] Row count shows correctly ("Showing X of Y")
- [ ] Empty state handled (0 rows)
- [ ] Proceed/Cancel buttons work
- [ ] All tests pass
- [ ] Both backend and frontend build successfully

---

**Plan complete and saved to `docs/superpowers/plans/2026-03-23-insert-preview-implementation.md`. Ready to execute?**
