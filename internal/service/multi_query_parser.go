package service

import (
	"strings"

	"github.com/yourorg/querybase/internal/models"
)

// ParsedStatement represents a single parsed SQL statement
type ParsedStatement struct {
	Sequence      int                  `json:"sequence"`
	QueryText     string               `json:"query_text"`
	OperationType models.OperationType `json:"operation_type"`
}

// ParseError represents an error during query parsing
type ParseError struct {
	Sequence int    `json:"sequence"`
	Position int    `json:"position"`
	Message  string `json:"message"`
}

// MultiQueryParseResult contains parsed statements and any errors
type MultiQueryParseResult struct {
	Statements []ParsedStatement `json:"statements"`
	Errors     []ParseError      `json:"errors"`
}

// ParseMultipleQueries splits a semicolon-separated SQL string into individual statements.
// It properly handles semicolons inside string literals and comments.
func ParseMultipleQueries(queryText string) *MultiQueryParseResult {
	result := &MultiQueryParseResult{
		Statements: make([]ParsedStatement, 0),
		Errors:     make([]ParseError, 0),
	}

	if strings.TrimSpace(queryText) == "" {
		return result
	}

	// State machine for parsing
	// States: normal, single_quote, double_quote, line_comment, block_comment
	type parseState int
	const (
		stateNormal parseState = iota
		stateSingleQuote
		stateDoubleQuote
		stateLineComment
		stateBlockComment
		stateEscapeSingle
		stateEscapeDouble
	)

	state := stateNormal
	var currentStmt strings.Builder
	sequence := 0

	for i := 0; i < len(queryText); i++ {
		char := queryText[i]

		switch state {
		case stateNormal:
			switch char {
			case '\'':
				state = stateSingleQuote
				currentStmt.WriteByte(char)
			case '"':
				state = stateDoubleQuote
				currentStmt.WriteByte(char)
			case '-':
				if i+1 < len(queryText) && queryText[i+1] == '-' {
					state = stateLineComment
					currentStmt.WriteByte(char)
				} else {
					currentStmt.WriteByte(char)
				}
			case '/':
				if i+1 < len(queryText) && queryText[i+1] == '*' {
					state = stateBlockComment
					currentStmt.WriteByte(char)
				} else {
					currentStmt.WriteByte(char)
				}
			case ';':
				// Statement terminator - save current statement if not empty
				stmt := strings.TrimSpace(currentStmt.String())
				if stmt != "" {
					result.Statements = append(result.Statements, ParsedStatement{
						Sequence:      sequence,
						QueryText:     stmt,
						OperationType: DetectOperationType(stmt),
					})
					sequence++
				}
				currentStmt.Reset()
			default:
				currentStmt.WriteByte(char)
			}

		case stateSingleQuote:
			currentStmt.WriteByte(char)
			if char == '\\' && i+1 < len(queryText) {
				state = stateEscapeSingle
			} else if char == '\'' {
				state = stateNormal
			}

		case stateDoubleQuote:
			currentStmt.WriteByte(char)
			if char == '\\' && i+1 < len(queryText) {
				state = stateEscapeDouble
			} else if char == '"' {
				state = stateNormal
			}

		case stateLineComment:
			currentStmt.WriteByte(char)
			if char == '\n' {
				state = stateNormal
			}

		case stateBlockComment:
			currentStmt.WriteByte(char)
			if char == '*' && i+1 < len(queryText) && queryText[i+1] == '/' {
				currentStmt.WriteByte(queryText[i+1])
				i++ // Skip the '/'
				state = stateNormal
			}

		case stateEscapeSingle:
			currentStmt.WriteByte(char)
			state = stateSingleQuote

		case stateEscapeDouble:
			currentStmt.WriteByte(char)
			state = stateDoubleQuote
		}
	}

	// Don't forget the last statement (if no trailing semicolon)
	stmt := strings.TrimSpace(currentStmt.String())
	if stmt != "" {
		result.Statements = append(result.Statements, ParsedStatement{
			Sequence:      sequence,
			QueryText:     stmt,
			OperationType: DetectOperationType(stmt),
		})
	}

	return result
}

// IsMultiQuery checks if a query text contains multiple statements
func IsMultiQuery(queryText string) bool {
	result := ParseMultipleQueries(queryText)
	return len(result.Statements) > 1
}

// ValidateMultiQuery validates that all statements in a multi-query are valid
func ValidateMultiQuery(queryText string) *MultiQueryParseResult {
	result := ParseMultipleQueries(queryText)

	// Validate each statement
	for i, stmt := range result.Statements {
		// Check for empty statements
		if strings.TrimSpace(stmt.QueryText) == "" {
			result.Errors = append(result.Errors, ParseError{
				Sequence: i,
				Position: 0,
				Message:  "Empty statement",
			})
			continue
		}

		// Check for transaction control statements (not allowed in multi-query)
		upperQuery := strings.ToUpper(stmt.QueryText)
		if strings.HasPrefix(upperQuery, "BEGIN") ||
			strings.HasPrefix(upperQuery, "COMMIT") ||
			strings.HasPrefix(upperQuery, "ROLLBACK") ||
			strings.HasPrefix(upperQuery, "START TRANSACTION") {
			result.Errors = append(result.Errors, ParseError{
				Sequence: i,
				Position: 0,
				Message:  "Transaction control statements (BEGIN, COMMIT, ROLLBACK, START TRANSACTION) are not allowed in multi-query mode",
			})
		}
	}

	return result
}
