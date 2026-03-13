# Changelog

All notable changes to QueryBase will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Enhanced Query Result Handling**:
  - **Empty Result Display**: SELECT queries returning 0 rows now show table structure with column headers and helpful empty state message
  - **Query Status Indicator**: Visual status component showing execution state (running, completed, empty, failed, no_match, pending_approval) with distinct colors and icons
  - **useQueryStatus Hook**: React hook for managing query execution state throughout the component lifecycle

- **Early Validation for Write Queries**:
  - **Automatic Validation**: UPDATE and DELETE queries are validated before creating approval requests
  - **Zero-Row Detection**: Queries that would affect 0 rows show validation modal immediately without creating approval
  - **QueryValidationModal Component**: New modal component displaying validation results with query preview and suggestions
  - **PreviewAndValidateWriteQuery Service**: Backend service method to validate write queries and count affected rows
  - **ValidationResult DTO**: New API response structure for validation results

### Changed

- **QueryResults Component**: Updated to handle empty data arrays while preserving column header display
- **QueryExecutor Component**: Integrated status indicator and validation handling
- **ExecuteQuery Handler**: Modified to validate write queries before approval creation
- **Query Status Type**: Added 'no_match' status to Query interface

### Fixed

- **Null Data Errors**: Fixed "Cannot read properties of null (reading 'length')" errors in QueryResults
- **Empty State UX**: Improved user experience for queries with no results

## [0.4.0] - 2026-02-28

### Added

- **Hardened Security Architecture**:
  - **Dual-Token Authentication**: Implemented short-lived access tokens (memory-only) and long-lived `HttpOnly/Secure/SameSite=Strict` refresh tokens (cookie-based).
  - **Token Revocation System**: Added Redis-based `TokenBlacklistService` for immediate session invalidation and access token revocation (JTI-based).
  - **Security Middleware Suite**: Added `SecurityHeadersMiddleware` (HSTS, CSP, XFO, etc.) and `SanitizationMiddleware` for automatic input cleansing.
  - **Intelligent Rate Limiting**: Upgraded rate limiter with prefix matching and generous burst support (300 req/min, burst 30) for normal navigation, while keeping strict limits (5 req/min) for authentication endpoints.
  - **Automatic Token Refresh**: Frontend `ApiClient` now handles 401 errors by seamlessly refreshing tokens using the rotated refresh token cookie.

### Changed

- **Frontend Auth Storage**: Removed `localStorage` persistence for sensitive tokens to mitigate persistent XSS vulnerabilities.
- **API CORS Configuration**: Strict origin reflection instead of wildcards when credentials are enabled.
- **Middleware Chain**: Reordered middleware for optimal security-first processing.

### Fixed

- **CORS Preflight Issues**: Resolved 403/Blocked-by-CORS errors during browser login by correctly reflecting origins and setting credential headers before preflight abort.
- **Navigation Throttling (429 Regression)**: Fixed aggressive rate limiting that blocked rapid menu switching by implementing prefix-based skip paths and increasing general burst capacity.
- **Token Leakage**: Prevented token exposure in browser history and local storage.

## [0.3.0] - 2026-02-21

### Added

- Live dashboard data implementation for Recent Activity, Pending Approvals, and Data Source Health
- Accurate approval status counts via new backend endpoint: `GET /api/v1/approvals/counts`
- Robustness improvements: API timeouts (5s/3s) and guaranteed loading state cleanup for dashboard widgets
- Admin password reset functionality - admins can reset any user's password
- Password reset API endpoint: `POST /api/v1/auth/users/:id/reset-password`
- Reset token fields to User model for future email-based password reset
- Comprehensive backend API test suite
- Repository structure documentation and cleanup
- CONTRIBUTING.md with development guidelines
- LICENSE file (MIT)

### Changed

- Revamped Dashboard UI with premium aesthetics (gradient headers, shimmering buttons)
- Replaced hardcoded dashboard metrics with live API data
- Updated `ApprovalList` to use real-time backend counts instead of local filtering
- Moved build scripts to `scripts/` directory for better organization
- Updated `.gitignore` to prevent compiled binaries from being committed
- Improved repository organization and documentation

### Fixed

- Approval count inconsistency between tabs in the approvals dashboard
- Dashboard stuck loading icon issue via timeout and `finally` blocks
- TypeScript property mismatches (`query_text`, `status` enums) in dashboard widgets
- Datasource password encryption issues with AES-256 key length
- Password decryption errors after JWT_SECRET changes
- Duplicate type declarations in DTO files

### Security

- Password reset requires admin privileges
- Prevents admins from resetting their own password via admin endpoint
- Password validation enforced (minimum 8 characters)
- Bcrypt hashing for all password storage

## [0.1.0] - Initial Release

### Added

- Database explorer with SQL query execution
- Support for PostgreSQL and MySQL datasources
- Approval workflow for write operations (INSERT, UPDATE, DELETE)
- Role-based access control (Admin, User, Viewer)
- JWT authentication
- Schema browser with table/column exploration
- Query history and saved queries
- Real-time query execution via WebSocket
- Background worker for async query execution
- Google Chat webhook notifications
- Docker Compose setup for local development
- Comprehensive documentation

### Features

- **Frontend**: Next.js with TypeScript, Monaco SQL editor
- **Backend**: Go with Gin framework
- **Database**: PostgreSQL (primary), MySQL (target datasources)
- **Cache**: Redis for job queue
- **Authentication**: JWT with role-based access
- **API**: RESTful endpoints with OpenAPI documentation

---

## Version History

- **Unreleased** - Current development
- **0.3.0** - Live Dashboard, real-time metrics, and robustness fixes
- **0.1.0** - Initial release with core features

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.
