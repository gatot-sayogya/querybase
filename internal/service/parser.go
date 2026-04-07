package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yourorg/querybase/internal/models"
)

// DetectOperationType detects the type of SQL operation (SELECT vs write operations)
func DetectOperationType(sql string) models.OperationType {
	// Trim whitespace and convert to uppercase for matching
	trimmedSQL := strings.TrimSpace(sql)
	upperSQL := strings.ToUpper(trimmedSQL)

	// Check for SET statements first (for user variables)
	if matchRegex(upperSQL, `^\s*SET\s+@`) {
		return models.OperationSet
	}

	// Check for write operations first (they take precedence)

	// Common DDL operations
	if matchRegex(upperSQL, `^\s*(CREATE|ALTER|DROP)\s+(TABLE|INDEX|VIEW|DATABASE|SCHEMA|TRIGGER|FUNCTION|PROCEDURE)\s+`) {
		return getOperationType(upperSQL)
	}

	// DML operations
	if matchRegex(upperSQL, `^\s*(INSERT|UPDATE|DELETE|REPLACE|TRUNCATE)\s+`) {
		return getOperationType(upperSQL)
	}

	// Transaction control
	if matchRegex(upperSQL, `^\s*(BEGIN|COMMIT|ROLLBACK|START\s+TRANSACTION)\s*`) {
		return models.OperationUpdate // Treat as write operation
	}

	// Grant/Revoke privileges
	if matchRegex(upperSQL, `^\s*(GRANT|REVOKE)\s+`) {
		return models.OperationUpdate
	}

	// Default to SELECT for read operations
	return models.OperationSelect
}

// getOperationType returns the specific operation type based on SQL keyword
func getOperationType(sql string) models.OperationType {
	upperSQL := strings.ToUpper(sql)

	switch {
	case strings.HasPrefix(upperSQL, "SELECT"):
		return models.OperationSelect
	case strings.HasPrefix(upperSQL, "INSERT"):
		return models.OperationInsert
	case strings.HasPrefix(upperSQL, "UPDATE"):
		return models.OperationUpdate
	case strings.HasPrefix(upperSQL, "DELETE"):
		return models.OperationDelete
	case strings.HasPrefix(upperSQL, "CREATE TABLE"):
		return models.OperationCreateTable
	case strings.HasPrefix(upperSQL, "DROP TABLE"):
		return models.OperationDropTable
	case strings.HasPrefix(upperSQL, "ALTER TABLE"):
		return models.OperationAlterTable
	default:
		// Default to update for other write operations
		return models.OperationUpdate
	}
}

// matchRegex checks if the SQL matches a regex pattern
func matchRegex(sql, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, sql)
	return matched
}

// SanitizeSQL removes comments and extra whitespace for better parsing
func SanitizeSQL(sql string) string {
	// Remove SQL comments (-- and /* */)

	// Remove single-line comments
	re := regexp.MustCompile(`--[^\n]*\n`)
	sql = re.ReplaceAllString(sql, " ")

	// Remove multi-line comments
	re = regexp.MustCompile(`/\*.*?\*/`)
	sql = re.ReplaceAllString(sql, " ")

	// Remove extra whitespace
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(sql, " ")

	return strings.TrimSpace(sql)
}

// normalizeSQLForExecution fixes common non-standard SQL syntax before sending
// to the database driver. Handles the most common user mistakes:
//   - Multi-line queries: collapses newlines/tabs/extra spaces into single spaces
//   - SQL comments: strips -- and /* */ comments
//   - DELETE without FROM: "DELETE table WHERE ..." → "DELETE FROM table WHERE ..."
//   - Trailing semicolons are stripped (some drivers choke on them)
func normalizeSQLForExecution(sql string) string {
	// Sanitize first: remove comments, collapse all whitespace (newlines, tabs, etc.)
	trimmed := SanitizeSQL(sql)

	// Strip trailing semicolon
	trimmed = strings.TrimRight(trimmed, "; \t\n\r")

	upper := strings.ToUpper(trimmed)

	// Fix: DELETE <table> WHERE ... → DELETE FROM <table> WHERE ...
	// MySQL (and standard SQL) requires FROM. Users often omit it.
	if strings.HasPrefix(upper, "DELETE ") && !strings.HasPrefix(upper, "DELETE FROM ") {
		// Prepend FROM after DELETE keyword
		trimmed = "DELETE FROM " + trimmed[7:]
	}

	return trimmed
}

// RequiresApproval returns true if the operation type requires approval
func RequiresApproval(operationType models.OperationType) bool {
	switch operationType {
	case models.OperationSelect, models.OperationSet:
		return false
	case models.OperationInsert,
		models.OperationUpdate,
		models.OperationDelete,
		models.OperationCreateTable,
		models.OperationDropTable,
		models.OperationAlterTable:
		return true
	default:
		return true
	}
}

// validateUpdateSetClause checks for common syntax errors in UPDATE SET clause
// specifically detecting when users use AND instead of comma between assignments
func validateUpdateSetClause(sql string) error {
	upperSQL := strings.ToUpper(sql)

	// Find the SET clause
	setIdx := strings.Index(upperSQL, " SET ")
	if setIdx == -1 {
		return nil // No SET clause, let other validations handle it
	}

	// Find WHERE clause (if exists) to limit the SET portion
	whereIdx := strings.Index(upperSQL, " WHERE ")

	// Extract the SET portion
	setStart := setIdx + 5 // len(" SET ")
	var setPortion string
	if whereIdx == -1 {
		setPortion = sql[setStart:]
	} else {
		setPortion = sql[setStart:whereIdx]
	}

	// Check for AND between assignments (indicating likely syntax error)
	// We need to distinguish between:
	// - SET col1 = 'val' AND col2 = 'val'  (incorrect - AND in SET)
	// - SET col1 = 'val' WHERE col2 = 'val' AND col3 = 'val'  (correct - AND in WHERE)

	// Simple heuristic: if there's an AND in the SET portion before any comparison
	// that looks like an assignment pattern, it's likely a mistake
	upperSetPortion := strings.ToUpper(setPortion)

	// Look for pattern: = ... AND ... =
	// This suggests: col1 = val AND col2 = val (should use comma)
	andIdx := strings.Index(upperSetPortion, " AND ")
	if andIdx != -1 {
		// Check if there's an = after the AND
		afterAnd := upperSetPortion[andIdx+5:]
		if strings.Contains(afterAnd, "=") {
			// This looks like: col1 = val AND col2 = val
			return fmt.Errorf("invalid UPDATE syntax: use comma (,) not AND to separate column assignments in SET clause")
		}
	}

	return nil
}

// ValidateSQL performs basic validation of SQL syntax before execution
// Returns an error if the SQL has obvious syntax issues
func ValidateSQL(sql string) error {
	trimmedSQL := strings.TrimSpace(sql)

	// Check if SQL is empty
	if trimmedSQL == "" {
		return fmt.Errorf("SQL query cannot be empty")
	}

	// Collapse newlines / tabs / extra spaces so that keyword checks like
	// " SET " work regardless of whether the user wrote the query on one line
	// or across multiple lines (e.g. "UPDATE t\nSET col = 1").
	normalizedSQL := SanitizeSQL(trimmedSQL)

	// Check for balanced parentheses on the normalised form
	openCount := strings.Count(normalizedSQL, "(")
	closeCount := strings.Count(normalizedSQL, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses in SQL query")
	}

	// All keyword checks use the normalised, uppercased form
	upperSQL := strings.ToUpper(normalizedSQL)

	// Check for SELECT without FROM (unless it's SELECT 1, SELECT variable, etc.)
	if strings.HasPrefix(upperSQL, "SELECT") {
		// Allow SELECT without FROM for simple expressions
		if !strings.Contains(upperSQL, " FROM ") {
			// Check if it's a simple expression like "SELECT 1" or "SELECT NOW()"
			selectPart := strings.TrimPrefix(trimmedSQL, "SELECT")
			selectPart = strings.TrimPrefix(selectPart, "select")
			selectPart = strings.TrimSpace(selectPart)

			// If it's just a simple expression, it's valid
			if !strings.Contains(strings.ToUpper(selectPart), " FROM") {
				return nil // Valid simple SELECT
			}
		}
	}

	// Check for INSERT without VALUES or SELECT
	if strings.HasPrefix(upperSQL, "INSERT") {
		if !strings.Contains(upperSQL, " VALUES") && !strings.Contains(upperSQL, " SELECT") && !strings.Contains(upperSQL, "SET ") {
			return fmt.Errorf("INSERT statement must include VALUES, SELECT, or SET clause")
		}
	}

	// Check for UPDATE without SET
	if strings.HasPrefix(upperSQL, "UPDATE") {
		if !strings.Contains(upperSQL, " SET ") {
			return fmt.Errorf("UPDATE statement must include SET clause")
		}

		// Check for common mistake: using AND instead of comma between column assignments
		// Example: UPDATE table SET col1 = 'val' AND col2 = 'val' (incorrect)
		// Should be: UPDATE table SET col1 = 'val', col2 = 'val' (correct)
		if err := validateUpdateSetClause(normalizedSQL); err != nil {
			return err
		}
	}

	// Check for DELETE without FROM (or TABLE in some SQL dialects)
	// Note: PostgreSQL, SQL Server, and Oracle allow DELETE without FROM (e.g., "DELETE table WHERE ...").
	// Therefore, we don't strictly enforce FROM or TABLE for DELETE statements.
	if strings.HasPrefix(upperSQL, "DELETE") {
		// Valid in many SQL dialects
	}

	// Check for CREATE TABLE without table name
	if strings.HasPrefix(upperSQL, "CREATE TABLE") {
		parts := strings.Fields(trimmedSQL)
		if len(parts) < 3 {
			return fmt.Errorf("CREATE TABLE statement must specify table name")
		}
	}

	// Check for DROP TABLE without table name
	if strings.HasPrefix(upperSQL, "DROP TABLE") {
		parts := strings.Fields(trimmedSQL)
		if len(parts) < 3 {
			return fmt.Errorf("DROP TABLE statement must specify table name")
		}
	}

	// Check for unterminated string literals
	// Count single quotes - should be even for properly terminated strings
	singleQuoteCount := 0
	inEscape := false
	for i := 0; i < len(normalizedSQL); i++ {
		char := normalizedSQL[i]
		if char == '\\' && inEscape {
			inEscape = true
			continue
		}
		if char == '\'' && !inEscape {
			singleQuoteCount++
		}
		inEscape = false
	}

	if singleQuoteCount%2 != 0 {
		return fmt.Errorf("unterminated string literal in SQL query")
	}

	return nil
}
