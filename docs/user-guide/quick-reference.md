# QueryBase Quick Reference: EXPLAIN & Dry Run

## TL;DR

- **EXPLAIN**: See how your query will execute (index usage, cost, row estimates)
- **Dry Run DELETE**: Preview which rows will be deleted before actually deleting them

---

## EXPLAIN Query

### What It Does
Shows the database execution plan for your query before running it.

### When to Use
- ✅ Query is slow and you want to know why
- ✅ Want to check if indexes are being used
- ✅ Want to optimize query performance
- ✅ Debugging JOIN performance

### Quick Example

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/queries/explain \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "your-datasource-uuid",
    "query_text": "SELECT * FROM users WHERE email = \"user@example.com\"",
    "analyze": false
  }'
```

**Response:**
```json
{
  "plan": [
    {
      "QUERY PLAN": "Index Scan using users_email_idx on users (cost=0.28..8.30 rows=1 width=248)"
    }
  ],
  "raw_output": "..."
}
```

### Key Indicators
- ✅ **Index Scan**: Using an index (good!)
- ⚠️ **Seq Scan**: Full table scan (need an index?)
- 📊 **cost**: Lower is better
- 📈 **rows**: Estimated row count

### EXPLAIN vs EXPLAIN ANALYZE

| Feature | EXPLAIN | EXPLAIN ANALYZE |
|---------|---------|-----------------|
| Executes query | ❌ No | ✅ Yes |
| Shows plan | ✅ Yes | ✅ Yes |
| Shows actual time | ❌ No | ✅ Yes |
| Shows actual rows | ❌ No | ✅ Yes |
| Use for | Planning | Real metrics |

**⚠️ Warning:** Only use EXPLAIN ANALYZE on SELECT queries!

---

## Dry Run DELETE

### What It Does
Converts DELETE to SELECT and shows you exactly which rows will be deleted.

### When to Use
- ✅ ALWAYS before running DELETE in production
- ✅ Want to verify WHERE clause is correct
- ✅ Want to see how many rows will be affected
- ✅ Want to preview the data that will be deleted

### Quick Example

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/queries/dry-run \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "your-datasource-uuid",
    "query_text": "DELETE FROM users WHERE status = \"inactive\""
  }'
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

### What to Check
1. **Affected Rows**: Is the count what you expect?
2. **Sample Data**: Do the rows look correct?
3. **WHERE Clause**: Are you deleting the right data?

---

## Typical Workflow

### For Slow Queries
```
1. Run EXPLAIN → See execution plan
2. Check for Seq Scan → Need index?
3. Create index (if needed)
4. Run EXPLAIN again → Verify index usage
5. Execute query
```

### For DELETE Operations
```
1. Run Dry Run → See affected rows
2. Review count and data → Is it correct?
3. Run EXPLAIN on DELETE → Check performance
4. Submit DELETE via approval workflow
5. Get approval
6. Execute DELETE
```

---

## Common Issues

### EXPLAIN Shows Seq Scan
**Problem:** Full table scan on large table
**Solution:** Create an index on the WHERE columns
```sql
CREATE INDEX idx_table_column ON table(column);
```

### Dry Run Returns 0 Rows
**Problem:** Expected rows but got 0
**Solution:**
- Check WHERE clause spelling
- Verify data format (case-sensitive?)
- Check for NULL values
- Try dry run without WHERE to see all data

### Dry Run is Slow
**Problem:** Too many rows affected
**Solution:**
- Add more specific WHERE conditions
- Consider adding a LIMIT for preview
- Check if you need all those rows deleted

---

## Quick Tips

### EXPLAIN
- Start with EXPLAIN (not ANALYZE) for safety
- Use EXPLAIN ANALYZE only on SELECT queries
- Look for "Index Scan" (good) vs "Seq Scan" (may need index)
- Check the "cost" value (lower is better)

### Dry Run DELETE
- ALWAYS dry run before DELETE in production
- Check the affected_rows count first
- Review sample rows to verify
- Make sure WHERE clause is specific enough
- Remember: foreign keys may prevent deletion

---

## API Endpoints

| Feature | Method | Endpoint | Permission |
|---------|--------|----------|------------|
| EXPLAIN | POST | `/api/v1/queries/explain` | `can_read` |
| Dry Run | POST | `/api/v1/queries/dry-run` | `can_write` |

---

## See Also

- [QUERY_FEATURES.md](QUERY_FEATURES.md) - Detailed documentation
- [CLAUDE.md](CLAUDE.md) - Project overview
- [API.md](API.md) - Full API reference
