# QueryBase Advanced Query Features

This document describes advanced query testing and analysis features in QueryBase.

## Table of Contents
- [EXPLAIN Queries](#explain-queries)
- [Dry Run DELETE](#dry-run-delete)
- [API Reference](#api-reference)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)

## EXPLAIN Queries

### Overview
EXPLAIN queries show how the database will execute your query, including:
- **Execution Plan**: The order of operations and algorithms used
- **Index Usage**: Which indexes are being used (or not used)
- **Estimated Row Counts**: How many rows the database expects to process
- **Join Methods**: How tables are joined (Nested Loop, Hash Join, Merge Join)
- **Cost Estimates**: Query cost for optimization decisions

### EXPLAIN vs EXPLAIN ANALYZE

**EXPLAIN** (Default):
- Shows the **planned** execution plan
- Does **not actually execute** the query
- Fast and safe to run on any query
- Provides estimates only

**EXPLAIN ANALYZE** (With `analyze: true`):
- Shows the **actual** execution plan
- **Actually executes** the query
- Provides real timing and row counts
- Use for SELECT queries only (never for write operations)

### Supported Databases

| Database | EXPLAIN | EXPLAIN ANALYZE | Notes |
|----------|---------|-----------------|-------|
| PostgreSQL | ✅ | ✅ | Full support |
| MySQL | ✅ | ⚠️ | MySQL uses `EXPLAIN ANALYZE` (8.0.18+) or `EXPLAIN FORMAT=TREE` |

### When to Use EXPLAIN

1. **Before Running Expensive Queries**
   ```sql
   -- Check if your query will use indexes
   EXPLAIN SELECT * FROM users WHERE email = 'user@example.com'
   ```

2. **Debug Slow Queries**
   ```sql
   -- See actual execution time
   EXPLAIN ANALYZE SELECT * FROM orders WHERE status = 'pending'
   ```

3. **Optimize JOINs**
   ```sql
   -- Check join methods and order
   EXPLAIN SELECT * FROM orders JOIN users ON orders.user_id = users.id
   ```

4. **Verify Index Usage**
   ```sql
   -- Make sure indexes are being used
   EXPLAIN SELECT * FROM products WHERE category_id = 5 AND price < 100
   ```

### Reading EXPLAIN Output

**PostgreSQL Output Example:**
```json
{
  "plan": [
    {
      "QUERY PLAN": "Index Scan using users_email_idx on users (cost=0.28..8.30 rows=1 width=248)"
    }
  ]
}
```

**Key Metrics:**
- **cost**: Estimated cost (lower is better)
- **rows**: Estimated number of rows
- **Index Scan**: Using an index (good!)
- **Seq Scan**: Full table scan (consider adding an index)

**MySQL Output Example:**
```json
{
  "plan": [
    {
      "id": 1,
      "select_type": "SIMPLE",
      "table": "users",
      "type": "ref",
      "possible_keys": "email_idx",
      "key": "email_idx",
      "rows": 1,
      "Extra": "Using index"
    }
  ]
}
```

### Common Issues and Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| Seq Scan on large table | No index or index not used | Create appropriate index |
| High cost | Complex query or missing indexes | Simplify query or add indexes |
| Nested Loop Join | Missing join condition | Add proper WHERE clause |
| Filter on large result set | Query retrieves too many rows | Add more selective WHERE clause |

---

## Dry Run DELETE

### Overview
Dry Run DELETE converts a DELETE query to a SELECT query to show exactly which rows will be affected **before** you execute the DELETE. This is a critical safety feature for data deletion operations.

### How It Works

1. **Conversion**: `DELETE FROM table WHERE ...` → `SELECT * FROM table WHERE ...`
2. **Execution**: The SELECT query is executed to show affected rows
3. **Preview**: Returns the count and actual row data that would be deleted
4. **Safety**: No data is actually deleted

### When to Use Dry Run

1. **Before Deleting Data**
   ```sql
   -- Check what will be deleted
   DELETE FROM users WHERE last_login < '2024-01-01'
   ```

2. **Verify WHERE Clause**
   ```sql
   -- Make sure your conditions are correct
   DELETE FROM orders WHERE status = 'cancelled' AND created_at < '2024-01-01'
   ```

3. **Count Affected Rows**
   ```sql
   -- See how many rows will be deleted
   DELETE FROM logs WHERE level = 'DEBUG' AND created_at < '2024-01-01'
   ```

4. **Review Data Before Deletion**
   ```sql
   -- Preview the actual data that will be deleted
   DELETE FROM products WHERE discontinued = true
   ```

### Dry Run Output Example

**Request:**
```json
{
  "query_text": "DELETE FROM users WHERE status = 'inactive'",
  "data_source_id": "uuid-here"
}
```

**Response:**
```json
{
  "affected_rows": 3,
  "query": "SELECT * FROM users WHERE status = 'inactive'",
  "rows": [
    {
      "id": 1,
      "name": "Alice",
      "status": "inactive",
      "email": "alice@example.com"
    },
    {
      "id": 2,
      "name": "Bob",
      "status": "inactive",
      "email": "bob@example.com"
    },
    {
      "id": 3,
      "name": "Charlie",
      "status": "inactive",
      "email": "charlie@example.com"
    }
  ]
}
```

### Best Practices

1. **Always Dry Run Before Deleting**
   - Run dry run first
   - Review affected rows
   - Verify count matches expectations
   - Then proceed with actual DELETE (via approval workflow)

2. **Use Specific WHERE Clauses**
   ```sql
   -- Good: Specific condition
   DELETE FROM logs WHERE created_at < '2024-01-01' AND level = 'DEBUG'

   -- Bad: Deletes everything (dry run will show all rows!)
   DELETE FROM logs
   ```

3. **Add LIMIT in Dry Run for Large Tables**
   ```sql
   -- If dry run returns too many rows, add a LIMIT to preview
   SELECT * FROM large_table WHERE condition LIMIT 100
   ```

4. **Check Foreign Key Constraints**
   - Dry run won't show foreign key constraint violations
   - Use EXPLAIN on the DELETE to check constraints
   - Review dependent tables before deleting

---

## API Reference

### EXPLAIN Query

**Endpoint:** `POST /api/v1/queries/explain`

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "data_source_id": "uuid-of-data-source",
  "query_text": "SELECT * FROM users WHERE email = 'user@example.com'",
  "analyze": false
}
```

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| data_source_id | string | Yes | UUID of the data source |
| query_text | string | Yes | SQL query to explain (SELECT, UPDATE, DELETE) |
| analyze | boolean | No | Use EXPLAIN ANALYZE (default: false) |

**Response:**
```json
{
  "plan": [
    {
      "QUERY PLAN": "Index Scan using users_email_idx on users (cost=0.28..8.30 rows=1 width=248)"
    }
  ],
  "raw_output": "[\n  {\n    \"QUERY PLAN\": \"Index Scan using users_email_idx on users...\"\n  }\n]"
}
```

**Error Response:**
```json
{
  "error": "EXPLAIN query failed: relation \"users\" does not exist"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/v1/queries/explain \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "SELECT * FROM users WHERE email = \"user@example.com\"",
    "analyze": false
  }'
```

### Dry Run DELETE

**Endpoint:** `POST /api/v1/queries/dry-run`

**Authentication:** Required (JWT token + write permission)

**Request Body:**
```json
{
  "data_source_id": "uuid-of-data-source",
  "query_text": "DELETE FROM users WHERE status = 'inactive'"
}
```

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| data_source_id | string | Yes | UUID of the data source |
| query_text | string | Yes | DELETE query to dry run |

**Response:**
```json
{
  "affected_rows": 3,
  "query": "SELECT * FROM users WHERE status = 'inactive'",
  "rows": [
    {
      "id": 1,
      "name": "Alice",
      "status": "inactive",
      "email": "alice@example.com"
    }
  ]
}
```

**Error Response:**
```json
{
  "error": "dry run is only supported for DELETE queries"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/v1/queries/dry-run \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "DELETE FROM users WHERE status = \"inactive\""
  }'
```

---

## Usage Examples

### Example 1: Optimize Slow Query

**Problem:** Query is slow
```sql
SELECT * FROM orders WHERE customer_id = 123 AND status = 'pending'
```

**Solution:** Use EXPLAIN to check
```bash
curl -X POST http://localhost:8080/api/v1/queries/explain \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "SELECT * FROM orders WHERE customer_id = 123 AND status = \"pending\"",
    "analyze": true
  }'
```

**Response shows:** `Seq Scan on orders` (no index!)

**Fix:** Create index
```sql
CREATE INDEX idx_orders_customer_status ON orders(customer_id, status)
```

**Verify:** Run EXPLAIN again - should show `Index Scan`

### Example 2: Safe Data Deletion

**Scenario:** Delete old log entries

**Step 1: Dry Run**
```bash
curl -X POST http://localhost:8080/api/v1/queries/dry-run \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "DELETE FROM logs WHERE created_at < \"2024-01-01\""
  }'
```

**Response:**
```json
{
  "affected_rows": 15234,
  "query": "SELECT * FROM logs WHERE created_at < '2024-01-01'",
  "rows": [
    { "id": 1, "message": "...", "created_at": "2023-12-31" },
    ...
  ]
}
```

**Step 2: Review**
- 15,234 rows will be deleted
- Sample rows look correct
- Date is as expected

**Step 3: Execute via Approval Workflow**
```bash
curl -X POST http://localhost:8080/api/v1/queries \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "DELETE FROM logs WHERE created_at < \"2024-01-01\""
  }'
```

### Example 3: Verify JOIN Performance

**Query:**
```sql
SELECT *
FROM orders o
JOIN customers c ON o.customer_id = c.id
JOIN products p ON o.product_id = p.id
WHERE o.status = 'pending'
```

**Step 1: EXPLAIN**
```bash
curl -X POST http://localhost:8080/api/v1/queries/explain \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "SELECT * FROM orders o JOIN customers c ON o.customer_id = c.id JOIN products p ON o.product_id = p.id WHERE o.status = \"pending\"",
    "analyze": true
  }'
```

**Step 2: Review Plan**
- Check join order (should start with most selective table)
- Check join methods (Hash Join for large tables, Nested Loop for small)
- Check if indexes are used for join conditions

---

## Best Practices

### EXPLAIN Queries

1. **Use EXPLAIN Before Production Queries**
   - Always test slow queries with EXPLAIN
   - Use EXPLAIN ANALYZE for realistic metrics
   - Compare different query formulations

2. **Check for Sequential Scans**
   - Seq Scan on large tables = bad
   - Create indexes on frequently filtered columns
   - Use `ANALYZE table` to update statistics

3. **Review JOIN Strategies**
   - Hash Join: Good for large unsorted datasets
   - Merge Join: Good for sorted datasets
   - Nested Loop: Good for small tables with indexes

4. **Monitor Actual vs Estimated Rows**
   - Large differences indicate outdated statistics
   - Run `ANALYZE` on tables regularly
   - Consider increasing statistics target

### Dry Run DELETE

1. **Always Dry Run Before Deleting**
   - Never skip dry run for production data
   - Review the affected rows count
   - Sample the row data to verify

2. **Use Specific WHERE Conditions**
   - Avoid DELETE without WHERE
   - Use multiple conditions for safety
   - Consider created_at thresholds

3. **Check for Foreign Keys**
   - Review dependent tables
   - Use CASCADE if appropriate
   - Delete child records first if needed

4. **Consider Transaction Safety**
   - For large deletes, use batch operations
   - Delete in chunks (e.g., 1000 rows at a time)
   - Use WHERE ... LIMIT for safety

### General Query Safety

1. **Write-Query Workflow**
   ```
   1. EXPLAIN the query → Check performance
   2. Dry run DELETE → Preview affected rows
   3. Submit for approval → Get approval
   4. Execute → Commit transaction
   ```

2. **Query Review Checklist**
   - [ ] EXPLAIN shows good performance
   - [ ] Indexes are being used
   - [ ] Estimated row count is reasonable
   - [ ] For DELETE: Dry run shows correct data
   - [ ] For DELETE: Affected rows count is expected
   - [ ] WHERE clause is specific enough

3. **Common Pitfalls**
   - ❌ Forgetting WHERE clause in DELETE
   - ❌ Using OR instead of AND in conditions
   - ❌ Not checking NULL values
   - ❌ Forgetting about time zones in dates
   - ❌ Deleting parent records without handling children

---

## Troubleshooting

### EXPLAIN Fails

**Error:** `"EXPLAIN query failed: relation does not exist"`

**Cause:** Table doesn't exist or permission denied

**Solution:**
1. Check table name spelling
2. Verify data source connection
3. Check user permissions on table

### Dry Run Returns No Rows

**Error:** Dry run returns 0 affected rows but you expected rows

**Possible Causes:**
1. WHERE clause is too restrictive
2. Data doesn't match condition
3. Table is empty

**Solution:**
1. Run dry run without WHERE to see all data
2. Check data format (dates, case sensitivity)
3. Verify table has data

### Dry Run is Slow

**Problem:** Dry run takes too long on large table

**Solution:**
1. Add LIMIT to dry run manually
2. Add more specific WHERE conditions
3. Use EXPLAIN to check if indexes exist
4. Consider adding indexes for WHERE columns

---

## Performance Tips

### Indexing Strategies

1. **Single Column Indexes**
   ```sql
   CREATE INDEX idx_users_email ON users(email);
   ```

2. **Composite Indexes**
   ```sql
   CREATE INDEX idx_orders_customer_status ON orders(customer_id, status);
   ```

3. **Covering Indexes**
   ```sql
   CREATE INDEX idx_products_category_price ON products(category, price) INCLUDE (name, sku);
   ```

4. **Partial Indexes**
   ```sql
   CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';
   ```

### Query Optimization

1. **Use Specific Columns**
   ```sql
   -- Instead of SELECT *
   SELECT id, name, email FROM users WHERE status = 'active'
   ```

2. **Use LIMIT for Preview**
   ```sql
   -- Add LIMIT when testing
   SELECT * FROM large_table WHERE condition LIMIT 100
   ```

3. **Avoid Functions in WHERE**
   ```sql
   -- Bad: Function prevents index use
   WHERE LOWER(name) = 'alice'

   -- Good: Use proper collation
   WHERE name = 'alice' COLLATE NOCASE
   ```

---

## Security Considerations

### EXPLAIN Queries
- ✅ Safe: Read-only operations
- ✅ No data modification
- ⚠️ EXPLAIN ANALYZE: Actually executes the query
- ❌ Never use EXPLAIN ANALYZE on write operations

### Dry Run DELETE
- ✅ Safe: No data deletion
- ✅ Only reads data
- ✅ Requires write permission (for safety)
- ✅ All queries logged

### Access Control
- EXPLAIN: Requires `can_read` permission
- Dry Run: Requires `can_write` permission
- Both operations are logged in query history

---

## Related Features

- [Query Execution](#): Standard query execution
- [Approval Workflow](#): Required for write operations
- [Query History](#): Track all query executions
- [Data Source Permissions](#): Control access per data source

---

## API Endpoints Summary

| Feature | Endpoint | Method | Auth | Permissions |
|---------|----------|--------|------|-------------|
| EXPLAIN Query | `/api/v1/queries/explain` | POST | Required | `can_read` |
| Dry Run DELETE | `/api/v1/queries/dry-run` | POST | Required | `can_write` |

---

## See Also

- [CLAUDE.md](CLAUDE.md) - Project overview
- [API.md](API.md) - Full API documentation
- [TESTING.md](TESTING.md) - Testing guide
- [QUERY_SAFE_DELETION_GUIDE.md](QUERY_SAFE_DELETION_GUIDE.md) - Best practices for safe data deletion
