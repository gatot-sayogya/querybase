export interface ParsedQuery {
  sequence: number;
  queryText: string;
  operationType: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'CREATE_TABLE' | 'DROP_TABLE' | 'ALTER_TABLE';
}

export interface ParseError {
  sequence: number;
  position: number;
  message: string;
}

export interface MultiQueryParseResult {
  statements: ParsedQuery[];
  errors: ParseError[];
}

type ParseState = 
  | 'normal' 
  | 'singleQuote' 
  | 'doubleQuote' 
  | 'lineComment' 
  | 'blockComment'
  | 'escapeSingle'
  | 'escapeDouble';

/**
 * Splits a semicolon-separated SQL string into individual statements.
 * Properly handles semicolons inside string literals and comments.
 */
export function parseMultipleQueries(queryText: string): MultiQueryParseResult {
  const result: MultiQueryParseResult = {
    statements: [],
    errors: []
  };

  if (!queryText || queryText.trim() === '') {
    return result;
  }

  // Normalize line endings so \r\n and \r don't end up inside statement text
  queryText = queryText.replace(/\r\n/g, '\n').replace(/\r/g, '\n');

  let state: ParseState = 'normal';
  let currentStmt = '';
  let sequence = 0;

  for (let i = 0; i < queryText.length; i++) {
    const char = queryText[i];
    const nextChar = queryText[i + 1];

    switch (state) {
      case 'normal':
        if (char === "'") {
          state = 'singleQuote';
          currentStmt += char;
        } else if (char === '"') {
          state = 'doubleQuote';
          currentStmt += char;
        } else if (char === '-' && nextChar === '-') {
          state = 'lineComment';
          currentStmt += char;
        } else if (char === '/' && nextChar === '*') {
          state = 'blockComment';
          currentStmt += char;
        } else if (char === ';') {
          // Statement terminator
          const stmt = currentStmt.trim();
          if (stmt) {
            result.statements.push({
              sequence,
              queryText: stmt + ';',
              operationType: detectOperationType(stmt)
            });
            sequence++;
          }
          currentStmt = '';
        } else {
          currentStmt += char;
        }
        break;

      case 'singleQuote':
        currentStmt += char;
        if (char === '\\' && i + 1 < queryText.length) {
          state = 'escapeSingle';
        } else if (char === "'") {
          state = 'normal';
        }
        break;

      case 'doubleQuote':
        currentStmt += char;
        if (char === '\\' && i + 1 < queryText.length) {
          state = 'escapeDouble';
        } else if (char === '"') {
          state = 'normal';
        }
        break;

      case 'lineComment':
        currentStmt += char;
        if (char === '\n') {
          state = 'normal';
        }
        break;

      case 'blockComment':
        currentStmt += char;
        if (char === '*' && nextChar === '/') {
          currentStmt += nextChar;
          i++; // Skip the '/'
          state = 'normal';
        }
        break;

      case 'escapeSingle':
        currentStmt += char;
        state = 'singleQuote';
        break;

      case 'escapeDouble':
        currentStmt += char;
        state = 'doubleQuote';
        break;
    }
  }

  // Don't forget the last statement (may have no trailing semicolon — add one)
  const stmt = currentStmt.trim();
  if (stmt) {
    result.statements.push({
      sequence,
      queryText: stmt + ';',
      operationType: detectOperationType(stmt)
    });
  }

  return result;
}

/**
 * Detects the operation type from a SQL query
 */
function detectOperationType(query: string): ParsedQuery['operationType'] {
  const trimmed = query.trim().toUpperCase();
  
  if (trimmed.startsWith('SELECT')) return 'SELECT';
  if (trimmed.startsWith('INSERT')) return 'INSERT';
  if (trimmed.startsWith('UPDATE')) return 'UPDATE';
  if (trimmed.startsWith('DELETE')) return 'DELETE';
  if (trimmed.startsWith('CREATE TABLE')) return 'CREATE_TABLE';
  if (trimmed.startsWith('DROP TABLE')) return 'DROP_TABLE';
  if (trimmed.startsWith('ALTER TABLE')) return 'ALTER_TABLE';
  
  return 'SELECT'; // Default
}

/**
 * Checks if a query text contains multiple statements
 */
export function isMultiQuery(queryText: string): boolean {
  const result = parseMultipleQueries(queryText);
  return result.statements.length > 1;
}

/**
 * Validates multiple queries and returns any errors
 */
export function validateMultiQuery(queryText: string): MultiQueryParseResult {
  const result = parseMultipleQueries(queryText);

  // Validate each statement
  result.statements.forEach((stmt, index) => {
    // Check for empty statements
    if (!stmt.queryText.trim()) {
      result.errors.push({
        sequence: index,
        position: 0,
        message: 'Empty statement'
      });
      return;
    }

    // Check for transaction control statements
    const upperQuery = stmt.queryText.toUpperCase().trim();
    if (
      upperQuery.startsWith('BEGIN') ||
      upperQuery.startsWith('COMMIT') ||
      upperQuery.startsWith('ROLLBACK') ||
      upperQuery.startsWith('START TRANSACTION')
    ) {
      result.errors.push({
        sequence: index,
        position: 0,
        message: 'Transaction control statements (BEGIN, COMMIT, ROLLBACK, START TRANSACTION) are not allowed in multi-query mode'
      });
    }
  });

  return result;
}

/**
 * Gets a summary of the multi-query
 */
export function getMultiQuerySummary(queryText: string): {
  statementCount: number;
  operations: Record<string, number>;
} {
  const result = parseMultipleQueries(queryText);
  const operations: Record<string, number> = {};

  result.statements.forEach(stmt => {
    operations[stmt.operationType] = (operations[stmt.operationType] || 0) + 1;
  });

  return {
    statementCount: result.statements.length,
    operations
  };
}
