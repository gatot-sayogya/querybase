package service

import (
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/yourorg/querybase/internal/models"
)

// SQLParseResult holds the result of parsing a SQL query
type SQLParseResult struct {
	Valid         bool
	Error         string
	OperationType models.OperationType
	Tables        []string
	Columns       []string
	QueryType     string // SELECT, INSERT, UPDATE, DELETE, etc.
}

// ParseAndValidateSQL parses and validates SQL using dialect-specific parsers
func ParseAndValidateSQL(sql string, dialect models.DataSourceType) (*SQLParseResult, error) {
	switch dialect {
	case models.DataSourceTypePostgreSQL:
		return parsePostgreSQL(sql)
	case models.DataSourceTypeMySQL:
		return parseMySQL(sql)
	default:
		// Fallback to basic validation for unsupported dialects
		return parseGeneric(sql)
	}
}

// parsePostgreSQL uses pg_query_go (real PostgreSQL parser)
func parsePostgreSQL(sql string) (*SQLParseResult, error) {
	result := &SQLParseResult{
		Valid: true,
	}

	// Parse the SQL using PostgreSQL's actual parser
	parseResult, err := pg_query.Parse(sql)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("PostgreSQL syntax error: %v", err)
		return result, nil
	}

	// Check for parse errors
	if parseResult == nil || len(parseResult.Stmts) == 0 {
		result.Valid = false
		result.Error = "No valid statements found"
		return result, nil
	}

	// For now, use basic operation type detection
	// Full AST traversal would require more complex protobuf handling
	result.OperationType = DetectOperationType(sql)
	result.QueryType = string(result.OperationType)

	return result, nil
}

// parseMySQL uses TiDB parser (MySQL-compatible)
func parseMySQL(sql string) (*SQLParseResult, error) {
	result := &SQLParseResult{
		Valid: true,
	}

	// Create a new TiDB parser
	p := parser.New()

	// Parse the SQL
	stmts, _, err := p.Parse(sql, "", "")
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("MySQL syntax error: %v", err)
		return result, nil
	}

	if len(stmts) == 0 {
		result.Valid = false
		result.Error = "No valid statements found"
		return result, nil
	}

	// Analyze the first statement
	stmt := stmts[0]

	// Extract operation type
	result.OperationType = getOperationTypeFromTiDB(stmt)
	result.QueryType = getTiDBQueryType(stmt)

	// Extract tables and columns
	result.Tables = extractTablesFromTiDB(stmt)
	result.Columns = extractColumnsFromTiDB(stmt)

	return result, nil
}

// parseGeneric provides basic validation for unsupported dialects
func parseGeneric(sql string) (*SQLParseResult, error) {
	// Use the existing ValidateSQL for basic checks
	if err := ValidateSQL(sql); err != nil {
		return &SQLParseResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &SQLParseResult{
		Valid:         true,
		OperationType: DetectOperationType(sql),
		QueryType:     string(DetectOperationType(sql)),
	}, nil
}

// getOperationTypeFromTiDB extracts operation type from TiDB AST
func getOperationTypeFromTiDB(stmt ast.StmtNode) models.OperationType {
	switch stmt.(type) {
	case *ast.SelectStmt:
		return models.OperationSelect
	case *ast.InsertStmt:
		return models.OperationInsert
	case *ast.UpdateStmt:
		return models.OperationUpdate
	case *ast.DeleteStmt:
		return models.OperationDelete
	case *ast.CreateTableStmt:
		return models.OperationCreateTable
	case *ast.DropTableStmt:
		return models.OperationDropTable
	case *ast.AlterTableStmt:
		return models.OperationAlterTable
	default:
		return models.OperationSelect
	}
}

// getTiDBQueryType returns a human-readable query type name
func getTiDBQueryType(stmt ast.StmtNode) string {
	switch stmt.(type) {
	case *ast.SelectStmt:
		return "SELECT"
	case *ast.InsertStmt:
		return "INSERT"
	case *ast.UpdateStmt:
		return "UPDATE"
	case *ast.DeleteStmt:
		return "DELETE"
	case *ast.CreateTableStmt:
		return "CREATE TABLE"
	case *ast.DropTableStmt:
		return "DROP TABLE"
	case *ast.AlterTableStmt:
		return "ALTER TABLE"
	default:
		return "UNKNOWN"
	}
}

// extractTablesFromTiDB extracts table names from TiDB AST
// Note: This is a simplified implementation. Full AST traversal would
// be needed for complex queries with joins, subqueries, etc.
func extractTablesFromTiDB(stmt ast.StmtNode) []string {
	// TODO: Implement full AST traversal for table extraction
	// For now, return empty - validation is the primary goal
	return []string{}
}

// extractColumnsFromTiDB extracts column names from TiDB AST
func extractColumnsFromTiDB(stmt ast.StmtNode) []string {
	columns := []string{}
	
	switch n := stmt.(type) {
	case *ast.SelectStmt:
		if n.Fields != nil {
			for _, field := range n.Fields.Fields {
				if field != nil && field.AsName.String() != "" {
					columns = append(columns, field.AsName.String())
				}
			}
		}
	case *ast.UpdateStmt:
		if n.List != nil {
			for _, assignment := range n.List {
				if assignment != nil && assignment.Column != nil {
					columns = append(columns, assignment.Column.Name.String())
				}
			}
		}
	}
	
	return columns
}

// ValidateSQLWithDialect validates SQL using the appropriate parser for the dialect
func ValidateSQLWithDialect(sql string, dialect models.DataSourceType) error {
	result, err := ParseAndValidateSQL(sql, dialect)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	if !result.Valid {
		return fmt.Errorf("%s", result.Error)
	}
	
	return nil
}

// DetectOperationTypeWithDialect detects operation type using dialect-specific parser
func DetectOperationTypeWithDialect(sql string, dialect models.DataSourceType) models.OperationType {
	result, err := ParseAndValidateSQL(sql, dialect)
	if err != nil || !result.Valid {
		// Fallback to basic detection
		return DetectOperationType(sql)
	}
	return result.OperationType
}

// ExtractTables extracts table names from SQL using dialect-specific parser
func ExtractTables(sql string, dialect models.DataSourceType) ([]string, error) {
	result, err := ParseAndValidateSQL(sql, dialect)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return nil, fmt.Errorf("parse error: %s", result.Error)
	}
	return result.Tables, nil
}

// IsValidSQL checks if SQL is valid for the given dialect
func IsValidSQL(sql string, dialect models.DataSourceType) (bool, string) {
	result, err := ParseAndValidateSQL(sql, dialect)
	if err != nil {
		return false, err.Error()
	}
	return result.Valid, result.Error
}

// NormalizeSQLForDialect normalizes SQL for the specific dialect
func NormalizeSQLForDialect(sql string, dialect models.DataSourceType) (string, error) {
	// First validate the SQL
	result, err := ParseAndValidateSQL(sql, dialect)
	if err != nil {
		return "", err
	}
	if !result.Valid {
		return "", fmt.Errorf("parse error: %s", result.Error)
	}
	
	// For MySQL, we can use TiDB's pretty print functionality
	if dialect == models.DataSourceTypeMySQL {
		p := parser.New()
		stmts, _, err := p.Parse(sql, "", "")
		if err != nil {
			return sql, nil // Return original if we can't parse
		}
		if len(stmts) > 0 {
			// Use TiDB's restore to normalize
			var buf strings.Builder
			ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &buf)
			err = stmts[0].Restore(ctx)
			if err == nil {
				return buf.String(), nil
			}
		}
	}
	
	// For other dialects, just return the original normalized SQL
	return normalizeSQLForExecution(sql), nil
}