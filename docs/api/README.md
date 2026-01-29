# API Overview

QueryBase REST API documentation and usage guide.

## Base URL

- **Development:** `http://localhost:8080`
- **Production:** `https://api.querybase.example.com`

All endpoints are prefixed with `/api/v1`

## Authentication

### JWT Token-Based Authentication

QueryBase uses JSON Web Tokens (JWT) for authentication.

#### Login Endpoint

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@querybase.local",
  "password": "admin123"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "admin@querybase.local",
    "username": "admin",
    "role": "admin"
  }
}
```

#### Using the Token

Include the token in the Authorization header for subsequent requests:

```http
GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Token Expiration

- **Default:** 24 hours
- **Configurable:** `JWT_EXPIRE_HOURS` environment variable

When the token expires, you'll receive a `401 Unauthorized` response. Login again to get a new token.

---

## Response Format

### Success Response

```json
{
  "data": { ... },
  "message": "Success"
}
```

### Error Response

```json
{
  "error": "Error message",
  "details": "Additional error details"
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created |
| 202 | Accepted - Request accepted for processing |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation error |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |

---

## Error Handling

### Common Error Types

#### Validation Errors (422)

```json
{
  "error": "validation failed",
  "details": {
    "field": "email",
    "message": "email is required"
  }
}
```

#### Authentication Errors (401)

```json
{
  "error": "unauthorized",
  "details": "invalid or expired token"
}
```

#### Authorization Errors (403)

```json
{
  "error": "forbidden",
  "details": "insufficient permissions for this operation"
}
```

#### Rate Limit Errors (429)

```json
{
  "error": "rate limit exceeded",
  "details": "maximum 60 requests per minute"
}
```

---

## Pagination

Query results support pagination:

### Request Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-indexed) |
| `page_size` | integer | 50 | Items per page (max 1000) |
| `sort_by` | string | - | Column to sort by |
| `sort_order` | string | asc | Sort direction (asc, desc) |

### Example

```http
GET /api/v1/queries/uuid/results?page=2&page_size=100&sort_by=id&sort_order=desc
```

### Response

```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "page_size": 100,
    "total_pages": 5,
    "total_rows": 500
  }
}
```

---

## Rate Limiting

QueryBase implements rate limiting for query execution only.

### Rate Limited Endpoints

- `POST /api/v1/queries` - Execute query

### Rate Limit Rules

- **Limit:** 60 requests per minute
- **Algorithm:** Token bucket
- **Headers:**
  - `X-RateLimit-Limit`: 60
  - `X-RateLimit-Remaining`: 45
  - `X-RateLimit-Reset`: Unix timestamp

**Note:** Schema endpoints, authentication, and data sources are NOT rate limited.

---

## Permissions

QueryBase uses Role-Based Access Control (RBAC) with data source permissions.

### User Roles

| Role | Permissions |
|------|-------------|
| `admin` | Full access to all features |
| `user` | Can execute queries, submit approvals, manage own data sources |
| `viewer` | Read-only access, SELECT queries only |

### Data Source Permissions

| Permission | Description |
|------------|-------------|
| `can_read` | Execute SELECT queries |
| `can_write` | Submit write operation requests |
| `can_approve` | Approve/reject write operations |

### Permission Errors

If a user lacks permissions:

```json
{
  "error": "forbidden",
  "details": "missing 'can_read' permission for data source"
}
```

---

## CORS

Cross-Origin Resource Sharing (CORS) is configurable.

### Configuration Priority

1. Environment variable: `CORS_ALLOWED_ORIGINS`
2. Config file: `config/config.yaml`
3. Code default: `*` (allow all)

### Setup

See [CORS Setup Guide](CORS_SETUP.md) for detailed configuration.

---

## API Endpoints Summary

### Authentication (3 endpoints)
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/change-password` - Change password

### Queries (6 endpoints)
- `POST /api/v1/queries` - Execute query
- `GET /api/v1/queries` - List queries
- `GET /api/v1/queries/:id` - Get query details
- `DELETE /api/v1/queries/:id` - Delete query
- `GET /api/v1/queries/:id/results` - Get results (paginated)
- `POST /api/v1/queries/save` - Save query

### Approvals (5 endpoints)
- `GET /api/v1/approvals` - List approvals
- `GET /api/v1/approvals/:id` - Get approval details
- `POST /api/v1/approvals/:id/review` - Review approval
- `POST /api/v1/transactions/:id/commit` - Commit transaction
- `POST /api/v1/transactions/:id/rollback` - Rollback transaction

### Data Sources (11 endpoints)
- `GET /api/v1/datasources` - List data sources
- `POST /api/v1/datasources` - Create data source
- `GET /api/v1/datasources/:id` - Get data source
- `PUT /api/v1/datasources/:id` - Update data source
- `DELETE /api/v1/datasources/:id` - Delete data source
- `POST /api/v1/datasources/:id/test` - Test connection
- `GET /api/v1/datasources/:id/schema` - Get schema
- `POST /api/v1/datasources/:id/sync` - Sync schema
- `GET /api/v1/datasources/:id/permissions` - Get permissions
- `PUT /api/v1/datasources/:id/permissions` - Set permissions
- `GET /api/v1/datasources/:id/health` - Health check

### Users (5 endpoints)
- `GET /api/v1/auth/users` - List users
- `POST /api/v1/auth/users` - Create user
- `GET /api/v1/auth/users/:id` - Get user
- `PUT /api/v1/auth/users/:id` - Update user
- `DELETE /api/v1/auth/users/:id` - Delete user

### Groups (6 endpoints)
- `GET /api/v1/groups` - List groups
- `POST /api/v1/groups` - Create group
- `GET /api/v1/groups/:id` - Get group
- `PUT /api/v1/groups/:id` - Update group
- `DELETE /api/v1/groups/:id` - Delete group
- `POST /api/v1/groups/:id/users` - Add user to group
- `DELETE /api/v1/groups/:id/users` - Remove user from group

### Health (1 endpoint)
- `GET /health` - Health check

**Total:** 41 endpoints

---

## Request/Response Examples

### Execute SELECT Query

**Request:**

```http
POST /api/v1/queries
Authorization: Bearer <token>
Content-Type: application/json

{
  "data_source_id": "uuid",
  "query_text": "SELECT * FROM users LIMIT 10"
}
```

**Response:**

```json
{
  "id": "query-uuid",
  "status": "completed",
  "row_count": 10,
  "columns": ["id", "email", "username"],
  "data": [
    {"id": 1, "email": "user@example.com", "username": "user1"}
  ],
  "execution_time_ms": 45
}
```

### Execute Write Query (Approval Workflow)

**Request:**

```http
POST /api/v1/queries
Authorization: Bearer <token>
Content-Type: application/json

{
  "data_source_id": "uuid",
  "query_text": "INSERT INTO users (email, username) VALUES ('test@example.com', 'testuser')"
}
```

**Response:**

```json
{
  "approval_id": "approval-uuid",
  "status": "pending_approval",
  "requires_approval": true,
  "message": "Query submitted for approval"
}
```

---

## SDKs and Clients

### Official Clients

- **JavaScript/TypeScript:** `@querybase/client` (planned)
- **Go:** `github.com/querybase/go-client` (planned)
- **Python:** `querybase-python` (planned)

### Using with cURL

```bash
# Login
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@querybase.local","password":"admin123"}' \
  | jq -r '.token')

# Execute query
curl -X POST http://localhost:8080/api/v1/queries \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"data_source_id":"uuid","query_text":"SELECT 1"}'
```

### Using with Axios (JavaScript)

```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
});

// Login
const { data } = await api.post('/auth/login', {
  email: 'admin@querybase.local',
  password: 'admin123'
});

const token = data.token;

// Set auth header
api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

// Execute query
const result = await api.post('/queries', {
  data_source_id: 'uuid',
  query_text: 'SELECT * FROM users LIMIT 10'
});

console.log(result.data);
```

---

## Versioning

The API uses URL versioning: `/api/v1/`

Future versions will be released as `/api/v2/`, etc.

### Backward Compatibility

- v1 APIs will be maintained for 12 months after v2 release
- Deprecated endpoints will return a `Deprecation` header
- Breaking changes will result in a new major version

---

## Changelog

### v1.0.0 (January 2026)
- Initial release
- 41 endpoints
- Authentication, queries, approvals, data sources, users, groups
- Permission system
- Rate limiting
- CORS support

---

## Support

- **Documentation:** [docs/](../README.md)
- **Issues:** [GitHub Issues](https://github.com/yourorg/querybase/issues)
- **Email:** support@querybase.example.com

---

**Next:** See [Endpoints Reference](endpoints.md) for detailed API documentation.
