# API Endpoints Reference

Complete reference for all QueryBase API endpoints.

**Base URL:** `http://localhost:8080/api/v1`

**Authentication:** All endpoints (except `/auth/login` and `/health`) require JWT token via `Authorization: Bearer <token>` header

---

## Authentication

### POST /auth/login

Login with email and password.

**Request:**

```json
{
  "email": "admin@querybase.local",
  "password": "admin123"
}
```

**Response (200):**

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

---

### GET /auth/me

Get current user information.

**Response (200):**

```json
{
  "id": "uuid",
  "email": "admin@querybase.local",
  "username": "admin",
  "role": "admin",
  "created_at": "2026-01-29T00:00:00Z",
  "last_login": "2026-01-29T12:00:00Z"
}
```

---

### POST /auth/change-password

Change current user's password.

**Request:**

```json
{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}
```

**Response (200):**

```json
{
  "message": "Password changed successfully"
}
```

---

### POST /auth/users/:id/reset-password

Reset a user's password (admin only).

**Request:**

```json
{
  "new_password": "newpassword123"
}
```

**Response (200):**

```json
{
  "message": "Password reset successfully"
}
```

**Permissions Required:** Admin only

**Restrictions:**

- Cannot reset own password via this endpoint (use `/auth/change-password` instead)
- Minimum password length: 8 characters

**Security Notes:**

- Password is hashed using bcrypt before storage
- Admin action should be logged for audit purposes
- Consider invalidating user sessions after password reset

---

## Queries

### POST /queries

Execute a query (SELECT or write operation).

**Request:**

```json
{
  "data_source_id": "uuid",
  "query_text": "SELECT * FROM users LIMIT 10"
}
```

**Response - SELECT (200):**

```json
{
  "id": "query-uuid",
  "status": "completed",
  "row_count": 10,
  "columns": ["id", "email", "username"],
  "data": [{ "id": 1, "email": "user@example.com", "username": "user1" }],
  "execution_time_ms": 45,
  "created_at": "2026-01-29T12:00:00Z"
}
```

**Response - Write Operation (202):**

```json
{
  "approval_id": "approval-uuid",
  "status": "pending_approval",
  "requires_approval": true,
  "message": "Query submitted for approval"
}
```

**Permissions Required:**

- SELECT: `can_read` on data source
- Write: `can_write` on data source

**Rate Limited:** Yes (60 req/min)

---

### GET /queries

List queries for current user.

**Query Parameters:**

- `page` (integer, default: 1)
- `page_size` (integer, default: 50)

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "query_text": "SELECT * FROM users",
      "status": "completed",
      "row_count": 100,
      "created_at": "2026-01-29T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_pages": 5,
    "total_rows": 250
  }
}
```

---

### GET /queries/:id

Get query details.

**Response (200):**

```json
{
  "id": "uuid",
  "data_source_id": "uuid",
  "query_text": "SELECT * FROM users",
  "operation_type": "SELECT",
  "status": "completed",
  "row_count": 100,
  "execution_time_ms": 45,
  "created_at": "2026-01-29T12:00:00Z"
}
```

---

### DELETE /queries/:id

Delete a query and its results.

**Response (200):**

```json
{
  "message": "Query deleted successfully"
}
```

**Permissions Required:** Query owner or admin

---

### GET /queries/:id/results

Get paginated query results.

**Query Parameters:**

- `page` (integer, default: 1)
- `page_size` (integer, default: 50)
- `sort_by` (string, optional)
- `sort_order` (string: "asc" or "desc", default: "asc")

**Response (200):**

```json
{
  "query_id": "uuid",
  "columns": ["id", "email", "username"],
  "data": [{ "id": 1, "email": "user@example.com", "username": "user1" }],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_pages": 2,
    "total_rows": 100
  }
}
```

---

### POST /queries/save

Save a query for later use.

**Request:**

```json
{
  "data_source_id": "uuid",
  "query_text": "SELECT * FROM users WHERE active = true",
  "name": "Active Users Query"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "Active Users Query",
  "query_text": "SELECT * FROM users WHERE active = true",
  "created_at": "2026-01-29T12:00:00Z"
}
```

---

## Approvals

### GET /approvals

List approval requests.

**Query Parameters:**

- `status` (string: "pending", "approved", "rejected", "executed", optional)
- `data_source_id` (uuid, optional)

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "query_text": "INSERT INTO users ...",
      "operation_type": "INSERT",
      "status": "pending",
      "requested_by": {
        "id": "uuid",
        "email": "user@example.com"
      },
      "data_source_id": "uuid",
      "created_at": "2026-01-29T12:00:00Z"
    }
  ]
}
```

**Permissions Required:** `can_approve` on data source or admin

---

### GET /approvals/counts

Get approval counts grouped by status.

**Response (200):**

```json
{
  "all": 10,
  "pending": 3,
  "approved": 5,
  "rejected": 2
}
```

**Permissions Required:** Any authenticated user (admins and approvers see global/accessible counts, regular users see counts for their requests)

---

### GET /approvals/:id

Get approval details.

**Response (200):**

```json
{
  "id": "uuid",
  "query_text": "INSERT INTO users ...",
  "operation_type": "INSERT",
  "status": "pending",
  "requested_by": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "requester"
  },
  "data_source_id": "uuid",
  "data_source": {
    "name": "Production DB"
  },
  "reviews": [],
  "created_at": "2026-01-29T12:00:00Z"
}
```

---

### POST /approvals/:id/review

Review (approve/reject) an approval request.

**Request:**

```json
{
  "decision": "approved",
  "comments": "Looks good, proceed"
}
```

**Response (200):**

```json
{
  "id": "uuid",
  "decision": "approved",
  "comments": "Looks good, proceed",
  "reviewed_by": "uuid",
  "reviewed_at": "2026-01-29T12:05:00Z"
}
```

**Permissions Required:** `can_approve` on data source

---

### POST /transactions/:id/commit

Commit a transaction (after preview).

**Response (200):**

```json
{
  "message": "Transaction committed successfully",
  "row_count": 5
}
```

---

### POST /transactions/:id/rollback

Rollback a transaction.

**Response (200):**

```json
{
  "message": "Transaction rolled back successfully"
}
```

---

## Data Sources

### GET /datasources

List data sources (filtered by user's permissions).

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Production Database",
      "type": "postgresql",
      "host": "db.example.com",
      "port": 5432,
      "database_name": "production",
      "status": "connected",
      "permissions": {
        "can_read": true,
        "can_write": true,
        "can_approve": false
      }
    }
  ]
}
```

---

### POST /datasources

Create a new data source.

**Request:**

```json
{
  "name": "Production Database",
  "type": "postgresql",
  "host": "db.example.com",
  "port": 5432,
  "database_name": "production",
  "username": "querybase",
  "password": "encrypted_password",
  "ssl_mode": "require"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "Production Database",
  "type": "postgresql",
  "host": "db.example.com",
  "created_at": "2026-01-29T12:00:00Z"
}
```

**Permissions Required:** Admin

---

### GET /datasources/:id

Get data source details.

**Response (200):**

```json
{
  "id": "uuid",
  "name": "Production Database",
  "type": "postgresql",
  "host": "db.example.com",
  "port": 5432,
  "database_name": "production",
  "username": "querybase",
  "ssl_mode": "require",
  "status": "connected",
  "last_health_check": "2026-01-29T12:00:00Z",
  "created_at": "2026-01-29T12:00:00Z"
}
```

**Permissions Required:** Access to data source

---

### PUT /datasources/:id

Update data source configuration.

**Request:**

```json
{
  "name": "Production Database (Updated)",
  "host": "new-db.example.com"
}
```

**Response (200):**

```json
{
  "id": "uuid",
  "name": "Production Database (Updated)",
  "updated_at": "2026-01-29T12:05:00Z"
}
```

**Permissions Required:** Admin

---

### DELETE /datasources/:id

Delete a data source.

**Response (200):**

```json
{
  "message": "Data source deleted successfully"
}
```

**Permissions Required:** Admin

---

### POST /datasources/:id/test

Test data source connection.

**Response (200):**

```json
{
  "success": true,
  "message": "Connection successful",
  "latency_ms": 45
}
```

**Response (400):**

```json
{
  "success": false,
  "message": "Connection failed: invalid credentials"
}
```

---

### GET /datasources/:id/schema

Get database schema (tables and columns).

**Query Parameters:**

- `table_name` (string, optional - filter by table)

**Response (200):**

```json
{
  "data_source_id": "uuid",
  "schema": [
    {
      "table_name": "users",
      "columns": [
        {
          "name": "id",
          "type": "integer",
          "nullable": false,
          "primary_key": true
        },
        {
          "name": "email",
          "type": "varchar(255)",
          "nullable": false
        }
      ]
    }
  ],
  "last_synced": "2026-01-29T12:00:00Z"
}
```

**Permissions Required:** `can_read` on data source

---

### POST /datasources/:id/sync

Force schema synchronization.

**Response (200):**

```json
{
  "message": "Schema synchronization started",
  "status": "syncing"
}
```

**Permissions Required:** `can_read` on data source

---

### GET /datasources/:id/permissions

Get data source permissions.

**Response (200):**

```json
{
  "data_source_id": "uuid",
  "permissions": [
    {
      "group_id": "uuid",
      "group_name": "Developers",
      "can_read": true,
      "can_write": true,
      "can_approve": false
    }
  ]
}
```

**Permissions Required:** Access to data source

---

### PUT /datasources/:id/permissions

Set data source permissions.

**Request:**

```json
{
  "permissions": [
    {
      "group_id": "uuid",
      "can_read": true,
      "can_write": true,
      "can_approve": false
    }
  ]
}
```

**Response (200):**

```json
{
  "message": "Permissions updated successfully"
}
```

**Permissions Required:** Admin

---

### GET /datasources/:id/health

Get data source health status.

**Response (200):**

```json
{
  "data_source_id": "uuid",
  "status": "healthy",
  "last_check": "2026-01-29T12:00:00Z",
  "latency_ms": 45,
  "error": null
}
```

---

## Users

### GET /auth/users

List all users.

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "username": "user",
      "role": "user",
      "active": true,
      "created_at": "2026-01-29T12:00:00Z"
    }
  ]
}
```

**Permissions Required:** Admin

---

### POST /auth/users

Create a new user.

**Request:**

```json
{
  "email": "newuser@example.com",
  "username": "newuser",
  "password": "password123",
  "role": "user"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "email": "newuser@example.com",
  "username": "newuser",
  "role": "user",
  "active": true,
  "created_at": "2026-01-29T12:00:00Z"
}
```

**Permissions Required:** Admin

---

### GET /auth/users/:id

Get user details.

**Response (200):**

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "user",
  "role": "user",
  "active": true,
  "created_at": "2026-01-29T12:00:00Z",
  "last_login": "2026-01-29T12:00:00Z",
  "groups": [
    {
      "id": "uuid",
      "name": "Developers"
    }
  ]
}
```

---

### PUT /auth/users/:id

Update user information.

**Request:**

```json
{
  "email": "updated@example.com",
  "role": "admin"
}
```

**Response (200):**

```json
{
  "id": "uuid",
  "email": "updated@example.com",
  "role": "admin",
  "updated_at": "2026-01-29T12:05:00Z"
}
```

**Permissions Required:** Admin or self (for limited fields)

---

### DELETE /auth/users/:id

Delete a user.

**Response (200):**

```json
{
  "message": "User deleted successfully"
}
```

**Permissions Required:** Admin

---

## Groups

### GET /groups

List all groups.

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Developers",
      "description": "Development team",
      "member_count": 5,
      "created_at": "2026-01-29T12:00:00Z"
    }
  ]
}
```

---

### POST /groups

Create a new group.

**Request:**

```json
{
  "name": "Developers",
  "description": "Development team"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "Developers",
  "description": "Development team",
  "created_at": "2026-01-29T12:00:00Z"
}
```

**Permissions Required:** Admin

---

### GET /groups/:id

Get group details.

**Response (200):**

```json
{
  "id": "uuid",
  "name": "Developers",
  "description": "Development team",
  "members": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "username": "user"
    }
  ],
  "created_at": "2026-01-29T12:00:00Z"
}
```

---

### PUT /groups/:id

Update group information.

**Request:**

```json
{
  "name": "Development Team",
  "description": "Updated description"
}
```

**Response (200):**

```json
{
  "id": "uuid",
  "name": "Development Team",
  "description": "Updated description",
  "updated_at": "2026-01-29T12:05:00Z"
}
```

**Permissions Required:** Admin

---

### DELETE /groups/:id

Delete a group.

**Response (200):**

```json
{
  "message": "Group deleted successfully"
}
```

**Permissions Required:** Admin

---

### POST /groups/:id/users

Add user to group.

**Request:**

```json
{
  "user_id": "uuid"
}
```

**Response (200):**

```json
{
  "message": "User added to group successfully"
}
```

**Permissions Required:** Admin

---

### DELETE /groups/:id/users

Remove user from group.

**Request:**

```json
{
  "user_id": "uuid"
}
```

**Response (200):**

```json
{
  "message": "User removed from group successfully"
}
```

**Permissions Required:** Admin

---

## Health

### GET /health

Health check endpoint.

**Response (200):**

```json
{
  "status": "ok",
  "timestamp": "2026-01-29T12:00:00Z"
}
```

**No authentication required.**

---

## Error Responses

All endpoints may return error responses:

### 400 Bad Request

```json
{
  "error": "bad request",
  "details": "invalid request body"
}
```

### 401 Unauthorized

```json
{
  "error": "unauthorized",
  "details": "missing or invalid authorization token"
}
```

### 403 Forbidden

```json
{
  "error": "forbidden",
  "details": "insufficient permissions"
}
```

### 404 Not Found

```json
{
  "error": "not found",
  "details": "resource not found"
}
```

### 422 Unprocessable Entity

```json
{
  "error": "validation failed",
  "details": {
    "field": "email",
    "message": "email is required"
  }
}
```

### 429 Too Many Requests

```json
{
  "error": "rate limit exceeded",
  "details": "maximum 60 requests per minute"
}
```

### 500 Internal Server Error

```json
{
  "error": "internal server error",
  "details": "an error occurred while processing your request"
}
```

---

**Total Endpoints:** 42

**See Also:**

- [API Overview](README.md)
- [CORS Setup](CORS_SETUP.md)
- [Main Documentation](../README.md)
