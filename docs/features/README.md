# Feature Documentation

Implementation details and technical documentation for QueryBase features.

## Features

### 1. [EXPLAIN and Dry Run](explain-dryrun.md)
**Query Analysis and Safe Deletion**

Comprehensive implementation documentation for:
- EXPLAIN query execution plans
- Dry run DELETE for safe data deletion
- API endpoints
- Service layer methods
- Testing strategies
- Usage examples

**What's Included:**
- Feature overview and use cases
- API reference
- Implementation details
- Code examples
- Test coverage
- Security considerations
- Performance impact

**When to Read:**
- Understanding feature implementation
- Maintaining or extending features
- Writing tests for features
- Debugging feature issues

## Feature Overview

### EXPLAIN Query
**Purpose:** Analyze query performance before execution

**Key Capabilities:**
- Show execution plans
- Display index usage
- Reveal join strategies
- Provide cost estimates
- Support EXPLAIN ANALYZE (actual execution metrics)

**API Endpoint:**
```
POST /api/v1/queries/explain
```

**Benefits:**
- Optimize slow queries
- Verify index usage
- Debug query performance
- Compare query formulations

### Dry Run DELETE
**Purpose:** Preview affected rows before deletion

**Key Capabilities:**
- Convert DELETE to SELECT
- Show affected row count
- Return sample of affected data
- Support complex WHERE clauses
- Handle subqueries and JOINs

**API Endpoint:**
```
POST /api/v1/queries/dry-run
```

**Benefits:**
- Prevent accidental data loss
- Verify WHERE clause correctness
- Count affected rows
- Review data before deletion

## Implementation Summary

### Files Modified

**Service Layer:**
- `internal/service/query.go`
  - Added `ExplainQuery()` method
  - Added `DryRunDelete()` method
  - Added `convertDeleteToSelect()` helper

**API Layer:**
- `internal/api/handlers/query.go`
  - Added `ExplainQuery()` handler
  - Added `DryRunDelete()` handler

**DTOs:**
- `internal/api/dto/query.go`
  - Added `ExplainQueryRequest`
  - Added `DryRunRequest`

**Routes:**
- `internal/api/routes/routes.go`
  - Added `/queries/explain` route
  - Added `/queries/dry-run` route

**Tests:**
- `internal/service/query_test.go`
  - Added 13 new test functions
  - 100% test pass rate

### API Endpoints

| Feature | Method | Endpoint | Permission |
|---------|--------|----------|------------|
| EXPLAIN | POST | `/api/v1/queries/explain` | `can_read` |
| Dry Run | POST | `/api/v1/queries/dry-run` | `can_write` |

### Test Coverage

**Total Tests:** 90/90 (100%)
- Auth tests: 18/18
- Parser tests: 30/30
- Query Service tests: 21/21 (including new tests)
- Models tests: 21/21

**New Tests:**
- ConvertDeleteToSelect: 10 test cases
- Dry Run Delete: 3 validation tests + 4 integration tests
- EXPLAIN Query: 3 integration tests

## Usage Examples

### EXPLAIN Query

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

```bash
curl -X POST http://localhost:8080/api/v1/queries/dry-run \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_source_id": "uuid-here",
    "query_text": "DELETE FROM users WHERE status = \"inactive\""
  }'
```

## Security Considerations

### EXPLAIN Queries
- ✅ Safe: Read-only operations
- ✅ No data modification
- ⚠️ EXPLAIN ANALYZE: Executes query (use with SELECT only)
- ❌ Never use EXPLAIN ANALYZE on write operations

### Dry Run DELETE
- ✅ Safe: Only reads data
- ✅ Requires `can_write` permission (safety measure)
- ✅ Validates operation is DELETE
- ✅ All operations logged

## Performance Impact

### EXPLAIN Queries
- **Overhead:** Minimal (planning only)
- **EXPLAIN ANALYZE:** Executes query (use with caution)
- **Recommendation:** Use EXPLAIN without ANALYZE for initial checks

### Dry Run DELETE
- **Overhead:** Executes SELECT query
- **Large Tables:** May be slow if many rows match
- **Recommendation:** Add LIMIT for preview if needed

## Future Enhancements

### Potential Improvements
1. **EXPLAIN for MySQL 8.0.18+:** Full `EXPLAIN ANALYZE` support
2. **Batch Dry Run:** Multiple DELETE queries
3. **Visual Explain Plans:** Graphical representation
4. **Cost Analysis:** Query optimization suggestions
5. **Auto-fix Suggestions:** Query improvements based on EXPLAIN

### Nice-to-Have
1. **Query History with EXPLAIN:** Store execution plans
2. **Performance Alerts:** Warn if cost exceeds threshold
3. **Index Recommendations:** Suggest indexes automatically

## Related Documentation

- **[User Guides](../guides/)** - How to use features
- **[Architecture](../architecture/)** - System design
- **[Development](../development/)** - Testing and building
- **[Query Features Guide](../guides/query-features.md)** - Comprehensive user guide

## See Also

- [QUERY_FEATURES.md](../guides/query-features.md) - Complete feature guide
- [FLOW_DIAGRAM.md](../architecture/flow.md) - Visual flow diagrams
- [CLAUDE.md](../../CLAUDE.md) - Complete project documentation

---

**Last Updated:** January 27, 2025
