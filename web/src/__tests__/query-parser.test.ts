import {
  parseMultipleQueries,
  isMultiQuery,
  validateMultiQuery,
  getMultiQuerySummary,
} from '@/lib/query-parser';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Return the queryText array from a parse result */
const stmts = (q: string) =>
  parseMultipleQueries(q).statements.map((s) => s.queryText);

/** Return the operationType array from a parse result */
const types = (q: string) =>
  parseMultipleQueries(q).statements.map((s) => s.operationType);

// ---------------------------------------------------------------------------
// parseMultipleQueries — JSON data in row values
// ---------------------------------------------------------------------------

describe('parseMultipleQueries — JSON in row values', () => {
  it('does not split on a semicolon inside a JSON string value', () => {
    const query = `INSERT INTO event_logs (payload) VALUES ('{"message":"Hello; World","status":"ok"}')`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].queryText).toBe(
      `INSERT INTO event_logs (payload) VALUES ('{"message":"Hello; World","status":"ok"}');`
    );
    expect(result.errors).toHaveLength(0);
  });

  it('does not split on multiple semicolons inside a JSON string value', () => {
    const query = `INSERT INTO event_logs (payload) VALUES ('{"sql":"SELECT 1; SELECT 2; SELECT 3","ok":true}')`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.errors).toHaveLength(0);
  });

  it('handles a JSON array value with semicolons in strings', () => {
    const query = `INSERT INTO products (tags) VALUES ('["tag;one","tag;two","tag;three"]')`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe('INSERT');
  });

  it('handles nested JSON objects in INSERT', () => {
    const query = `INSERT INTO customers (name, metadata) VALUES ('Jane Doe', '{"tier":"gold","address":{"city":"New York","zip":"10001"},"tags":["vip","loyal"]}')`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe('INSERT');
    expect(result.errors).toHaveLength(0);
  });

  it('handles JSONB merge operator in UPDATE with semicolon in value', () => {
    const query = `UPDATE customers SET metadata = metadata || '{"note":"VIP; Priority","updated":true}' WHERE id = 1`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe('UPDATE');
  });

  it('does not split a SELECT with JSONB ->> operator and semicolon in compared value', () => {
    const query = `SELECT * FROM customers WHERE metadata->>'note' = 'VIP; Priority'`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe('SELECT');
  });

  it('correctly splits two INSERT statements that each contain JSON values', () => {
    const query = [
      `INSERT INTO event_logs (event_type, payload) VALUES ('order_created', '{"order_id":1,"total":99.99}');`,
      `INSERT INTO event_logs (event_type, payload) VALUES ('payment_processed', '{"order_id":1,"amount":99.99,"gateway":"stripe"}')`,
    ].join('\n');

    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(2);
    expect(result.statements[0].operationType).toBe('INSERT');
    expect(result.statements[1].operationType).toBe('INSERT');
    expect(result.errors).toHaveLength(0);
  });

  it('correctly splits SELECT + UPDATE where UPDATE has a JSON value', () => {
    const query = [
      `SELECT id FROM customers WHERE email = 'test@example.com';`,
      `UPDATE customers SET metadata = '{"tier":"silver","migrated":true}' WHERE email = 'test@example.com'`,
    ].join('\n');

    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(2);
    expect(result.statements[0].operationType).toBe('SELECT');
    expect(result.statements[1].operationType).toBe('UPDATE');
  });

  it('handles a multi-line INSERT with a JSON object value', () => {
    const query = `INSERT INTO customers (
  first_name,
  last_name,
  metadata
) VALUES (
  'John',
  'Doe',
  '{"tier":"gold","tags":["vip","loyal"],"credit_limit":5000}'
)`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe('INSERT');
    expect(result.errors).toHaveLength(0);
  });

  it('handles a multi-line UPDATE with JSONB || merge and newlines in query', () => {
    const query = `UPDATE customers
SET
  metadata = metadata || '{"tier":"gold","upgraded_at":"2024-02-01"}',
  updated_at = NOW()
WHERE email = 'bob.johnson@example.com'`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe('UPDATE');
  });
});

// ---------------------------------------------------------------------------
// parseMultipleQueries — semicolons auto-added to all statements
// ---------------------------------------------------------------------------

describe('parseMultipleQueries — semicolons in output', () => {
  it('adds a semicolon to a single query that has no trailing semicolon', () => {
    const query = `SELECT * FROM customers WHERE id = 1`;
    const result = parseMultipleQueries(query);
    expect(result.statements[0].queryText).toBe(
      `SELECT * FROM customers WHERE id = 1;`
    );
  });

  it('adds a semicolon to a query that already has a trailing semicolon (normalised to one)', () => {
    const query = `SELECT 1;`;
    const result = parseMultipleQueries(query);
    // statement delimiter is consumed, then ';' appended — result is "SELECT 1;"
    expect(result.statements[0].queryText).toBe(`SELECT 1;`);
  });

  it('adds semicolons to every statement in a multi-query', () => {
    const query = `SELECT 1; SELECT 2; SELECT 3`;
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(3);
    result.statements.forEach((s) => {
      expect(s.queryText.endsWith(';')).toBe(true);
    });
  });

  it('last statement without trailing semicolon gets one added', () => {
    const query = `SELECT 1;\nSELECT 2`;
    const result = parseMultipleQueries(query);
    expect(result.statements[1].queryText).toBe(`SELECT 2;`);
  });
});

// ---------------------------------------------------------------------------
// parseMultipleQueries — line ending normalisation
// ---------------------------------------------------------------------------

describe('parseMultipleQueries — line ending normalisation', () => {
  it('handles \\r\\n (Windows) line endings without embedding \\r in statement text', () => {
    const query = 'SELECT *\r\nFROM customers\r\nWHERE id = 1';
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].queryText).not.toContain('\r');
    expect(result.statements[0].queryText).toBe(
      'SELECT *\nFROM customers\nWHERE id = 1;'
    );
  });

  it('handles \\r\\n in a multi-query', () => {
    const query =
      'INSERT INTO event_logs (event_type) VALUES (\'login\');\r\nSELECT * FROM event_logs';
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(2);
    result.statements.forEach((s) => {
      expect(s.queryText).not.toContain('\r');
    });
  });

  it('normalises \\r\\n inside a JSON value', () => {
    const query =
      "INSERT INTO event_logs (payload) VALUES ('{\"msg\":\"hello\"}')";
    // simulate Windows editor: replace \n with \r\n
    const windowsQuery = query.replace(/\n/g, '\r\n');
    const result = parseMultipleQueries(windowsQuery);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].queryText).not.toContain('\r');
  });
});

// ---------------------------------------------------------------------------
// parseMultipleQueries — sequence numbers
// ---------------------------------------------------------------------------

describe('parseMultipleQueries — sequence numbers', () => {
  it('assigns correct sequence numbers to three statements', () => {
    const query = `SELECT 1;\nSELECT 2;\nSELECT 3`;
    const result = parseMultipleQueries(query);
    expect(result.statements.map((s) => s.sequence)).toEqual([0, 1, 2]);
  });
});

// ---------------------------------------------------------------------------
// isMultiQuery — with JSON data
// ---------------------------------------------------------------------------

describe('isMultiQuery — JSON in row values', () => {
  it('returns false for a single INSERT with a JSON value', () => {
    expect(
      isMultiQuery(
        `INSERT INTO event_logs (payload) VALUES ('{"event":"login","user_id":1}')`
      )
    ).toBe(false);
  });

  it('returns false for a single UPDATE with a JSONB merge', () => {
    expect(
      isMultiQuery(
        `UPDATE customers SET metadata = metadata || '{"tier":"gold"}' WHERE id = 1`
      )
    ).toBe(false);
  });

  it('returns false for a single SELECT with a JSONB @> operator', () => {
    expect(
      isMultiQuery(`SELECT * FROM products WHERE tags @> '["laptop"]'`)
    ).toBe(false);
  });

  it('returns true for two INSERT statements each containing JSON values', () => {
    expect(
      isMultiQuery(
        `INSERT INTO event_logs (payload) VALUES ('{"type":"a"}'); INSERT INTO event_logs (payload) VALUES ('{"type":"b"}')`
      )
    ).toBe(true);
  });

  it('returns false for a single query with a JSON value containing multiple semicolons', () => {
    expect(
      isMultiQuery(
        `INSERT INTO notes (body) VALUES ('{"text":"step1; step2; step3","done":false}')`
      )
    ).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// validateMultiQuery — JSON data
// ---------------------------------------------------------------------------

describe('validateMultiQuery — JSON in row values', () => {
  it('returns no errors for a valid multi-INSERT with JSON', () => {
    const query = [
      `INSERT INTO event_logs (event_type, payload) VALUES ('login', '{"user_id":1}');`,
      `INSERT INTO event_logs (event_type, payload) VALUES ('logout', '{"user_id":1}')`,
    ].join('\n');
    const result = validateMultiQuery(query);
    expect(result.errors).toHaveLength(0);
  });

  it('returns an error when BEGIN is present even in a JSON-heavy multi-query', () => {
    const query = [
      `INSERT INTO event_logs (payload) VALUES ('{"type":"login"}');`,
      `BEGIN`,
    ].join('\n');
    const result = validateMultiQuery(query);
    expect(result.errors.length).toBeGreaterThan(0);
  });
});

// ---------------------------------------------------------------------------
// getMultiQuerySummary — JSON data
// ---------------------------------------------------------------------------

describe('getMultiQuerySummary — JSON in row values', () => {
  it('counts INSERT, UPDATE, SELECT correctly when queries contain JSON', () => {
    const query = [
      `INSERT INTO event_logs (payload) VALUES ('{"type":"a"}');`,
      `INSERT INTO event_logs (payload) VALUES ('{"type":"b"}');`,
      `UPDATE customers SET metadata = '{"tier":"gold"}' WHERE id = 1;`,
      `SELECT * FROM event_logs WHERE payload->>'type' = 'a'`,
    ].join('\n');

    const summary = getMultiQuerySummary(query);
    expect(summary.statementCount).toBe(4);
    expect(summary.operations['INSERT']).toBe(2);
    expect(summary.operations['UPDATE']).toBe(1);
    expect(summary.operations['SELECT']).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// parseMultipleQueries — operation type detection with JSON
// ---------------------------------------------------------------------------

describe('parseMultipleQueries — operationType with JSON queries', () => {
  it.each([
    [
      'INSERT with JSON',
      `INSERT INTO customers (metadata) VALUES ('{"tier":"gold"}')`,
      'INSERT',
    ],
    [
      'UPDATE with JSONB merge',
      `UPDATE customers SET metadata = metadata || '{"active":true}' WHERE id = 1`,
      'UPDATE',
    ],
    [
      'DELETE with JSONB filter',
      `DELETE FROM event_logs WHERE payload->>'event_type' = 'test'`,
      'DELETE',
    ],
    [
      'SELECT with JSONB ->>',
      `SELECT id, metadata->>'tier' AS tier FROM customers`,
      'SELECT',
    ],
    [
      'SELECT with JSONB @> containment',
      `SELECT * FROM products WHERE tags @> '["laptop","portable"]'`,
      'SELECT',
    ],
  ])('%s', (_, query, expectedType) => {
    const result = parseMultipleQueries(query);
    expect(result.statements).toHaveLength(1);
    expect(result.statements[0].operationType).toBe(expectedType);
  });
});
