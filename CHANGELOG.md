# Changelog

All notable changes to QueryBase will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
