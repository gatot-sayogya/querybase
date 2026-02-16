# QueryBase Documentation

**Last Updated:** January 29, 2026

Comprehensive documentation for QueryBase - A database explorer with approval workflow.

## Quick Navigation

### 👥 For Users

- **[Getting Started](getting-started/)** - Setup and installation
- **[User Guide](user-guide/)** - How to use QueryBase features

### 👨‍💻 For Developers

- **[Development Guide](development/)** - Local development setup
- **[Architecture](architecture/)** - System architecture and design
- **[API Reference](api/)** - API endpoints and usage

### 🎯 Project Status

- **[Implementation Status](planning/IMPLEMENTATION_STATUS.md)** - Current progress
- **[Planning](planning/)** - Roadmap and implementation plans

---

## Documentation Index

### Getting Started 🚀

- **[Installation Guide](getting-started/README.md)**
  - Prerequisites
  - 5-minute quick start
  - Docker setup
  - Configuration

### User Guide 📖

- **[Query Features](user-guide/query-features.md)**
  - SQL editor with autocomplete
  - Query execution (SELECT, INSERT, UPDATE, DELETE)
  - EXPLAIN and Dry Run features
  - Query history and export

- **[Approval Workflow](user-guide/approval-workflow.md)**
  - How write operations work
  - Reviewing and approving queries
  - Transaction management

- **[Schema Browser](user-guide/schema-browser.md)**
  - Exploring database schemas
  - Viewing tables and columns
  - Schema synchronization

- **[Admin Panel](user-guide/admin-panel.md)**
  - Managing users and groups
  - Data source configuration
  - Permission management

### Development 💻

- **[Development Overview](development/README.md)**
  - Development workflow
  - Project structure
  - Code organization

- **[Local Development Setup](development/setup.md)**
  - Setting up your development environment
  - Running servers (API, Worker, Frontend)
  - Hot reload and debugging

- **[Backend Development](development/backend.md)**
  - Go project structure
  - Adding new API endpoints
  - Database migrations
  - Testing

- **[Frontend Development](development/frontend.md)**
  - Next.js project structure
  - Adding new pages
  - State management with Zustand
  - Component development

- **[Testing Guide](development/testing.md)**
  - Unit tests (backend)
  - Unit tests (frontend)
  - Integration tests
  - E2E tests with Playwright

- **[Build Guide](development/build.md)**
  - Building for production
  - Multi-architecture builds
  - Deployment

- **[Session Summary](development/session-summary.md)**
  - Development history
  - Recent changes

### Architecture 🏗️

- **[Architecture Overview](architecture/README.md)**
  - System design
  - Technology stack
  - Key architectural decisions

- **[Flow Diagrams](architecture/flow.md)**
  - Query execution flow
  - Approval workflow
  - Authentication flow
  - Visual diagrams

- **[Technical Flow](architecture/detailed-flow.md)**
  - Detailed step-by-step flows
  - Security layers
  - Performance considerations

### API Reference 🔌

- **[API Overview](api/README.md)**
  - Authentication
  - Response formats
  - Error handling

- **[Endpoints](api/endpoints.md)**
  - All 41 API endpoints
  - Request/response examples
  - Permission requirements

- **[CORS Setup](api/CORS_SETUP.md)**
  - Configuring CORS
  - Allowed origins
  - Environment variables

### Planning 📋

- **[Planning Overview](planning/README.md)**
  - Planning documents
  - Implementation priorities

- **[Implementation Status](planning/IMPLEMENTATION_STATUS.md)**
  - Current progress (✅ Backend & Frontend Complete!)
  - Completed features
  - Next steps

- **[Historical Plans](planning/)**
  - CORE_WORKFLOW_PLAN.md - Backend completion plan
  - DASHBOARD_UI_CURRENT_WORKFLOW.md - Frontend implementation
  - IMPLEMENTATION_TESTING_PLAN.md - Testing roadmap

### Features ✨

- **[Features Overview](features/README.md)** - Feature documentation
- **[EXPLAIN & Dry Run](features/explain-dryrun.md)** - Query analysis features
- **[Encrypted Communication](features/ENCRYPTED_COMMUNICATION.md)** - Security documentation

### Project Documentation

- **[CLAUDE.md](development/CLAUDE.md)** - Complete project guide for AI assistants
- **[Main README](../README.md)** - Project overview
- **[CONTRIBUTING](../CONTRIBUTING.md)** - How to contribute to this project
- **[CHANGELOG](../CHANGELOG.md)** - Version history and recent changes
- **[LICENSE](../LICENSE)** - MIT License

---

## Project Status Summary

### Backend: ✅ Complete (~95%)

**Completed Features:**

- ✅ All infrastructure (database, models, auth, config)
- ✅ Query execution engine with SQL parser
- ✅ Approval workflow system with transaction preview
- ✅ Data source management with encryption
- ✅ Redis queue + background worker
- ✅ Google Chat notifications
- ✅ User & Group Management
- ✅ EXPLAIN and Dry Run features
- ✅ Query results pagination with sorting
- ✅ Query export (CSV/JSON)
- ✅ Approval comments system
- ✅ Data source health check API
- ✅ Error handling improvements
- ✅ Request logging & panic recovery middleware

**API Endpoints:** 41 endpoints implemented

### Frontend: ✅ Complete (~90%)

**Completed Features:**

- ✅ Next.js 15+ with App Router
- ✅ SQL editor with Monaco Editor
- ✅ Intelligent autocomplete (tables/columns)
- ✅ Query results viewer (pagination, sorting)
- ✅ Approval dashboard
- ✅ Admin panel (users, groups, data sources)
- ✅ Schema browser with polling
- ✅ Permission-based UI filtering
- ✅ Authentication and authorization
- ✅ Query history and export

**Next Step:** Polish and optimization

### Testing: ✅ 90/90 Tests Passing (100%)

- Auth tests: 18/18 PASS
- Parser tests: 30/30 PASS
- Query Service tests: 21/21 PASS
- Models tests: 21/21 PASS

**Integration Tests:** 37 test cases ready

---

## Quick Links

- **[GitHub Repository](https://github.com/yourorg/querybase)**
- **[Main README](../README.md)** - Project overview
- **[CONTRIBUTING](../CONTRIBUTING.md)** - Contribution guidelines
- **[CHANGELOG](../CHANGELOG.md)** - Version history
- **[CLAUDE.md](development/CLAUDE.md)** - Complete development guide

---

## Technology Stack

### Backend

- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL 15 (primary)
- **Queue:** Redis 7 (Asynq)
- **Auth:** JWT (golang-jwt/jwt)

### Frontend

- **Framework:** Next.js 15+ (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Editor:** Monaco Editor
- **State:** Zustand

---

## Next Steps

1. ✅ **Backend complete!** (All core features implemented)
2. ✅ **Frontend complete!** (All UI components implemented)
3. ⏳ **Polish and optimization** (Performance, UX improvements)
4. ⏳ **Comprehensive testing** (Integration tests, E2E tests)
5. ⏳ **Production deployment** (Infrastructure, monitoring)

---

**Looking for something specific?** Use the search in your IDE or check the index above.
