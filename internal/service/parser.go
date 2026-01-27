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

// RequiresApproval returns true if the operation type requires approval
func RequiresApproval(operationType models.OperationType) bool {
	switch operationType {
	case models.OperationSelect:
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

// ValidateSQL performs basic validation of SQL syntax before execution
// Returns an error if the SQL has obvious syntax issues
func ValidateSQL(sql string) error {
	trimmedSQL := strings.TrimSpace(sql)

	// Check if SQL is empty
	if trimmedSQL == "" {
		return fmt.Errorf("SQL query cannot be empty")
	}

	// Check for balanced parentheses
	openCount := strings.Count(trimmedSQL, "(")
	closeCount := strings.Count(trimmedSQL, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses in SQL query")
	}

	// Check for common syntax errors
	upperSQL := strings.ToUpper(trimmedSQL)

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
		if !strings.Contains(upperSQL, " SET ") && !strings.Contains(upperSQL, "\nSET ") {
			return fmt.Errorf("UPDATE statement must include SET clause")
		}
	}

	// Check for DELETE without FROM (or TABLE in some SQL dialects)
	if strings.HasPrefix(upperSQL, "DELETE") {
		if !strings.Contains(upperSQL, " FROM ") && !strings.Contains(upperSQL, " TABLE ") && !strings.Contains(upperSQL, "\nFROM ") {
			return fmt.Errorf("DELETE statement must include FROM or TABLE clause")
		}
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
	for i := 0; i < len(trimmedSQL); i++ {
		char := trimmedSQL[i]
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
