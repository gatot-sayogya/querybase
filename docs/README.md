# QueryBase Documentation

Comprehensive documentation for QueryBase - A database explorer with approval workflow.

## Quick Start

- **[README](../README.md)** - Project overview and quick start
- **[Getting Started](getting-started/)** - Setup and installation guide

## User Guides

- **[Query Features Guide](guides/query-features.md)** - EXPLAIN and Dry Run features
- **[Quick Reference](guides/quick-reference.md)** - Quick reference for daily use

## Architecture

- **[Flow Diagrams](architecture/flow.md)** - Visual flow diagrams
- **[Technical Flow](architecture/detailed-flow.md)** - Detailed technical flow documentation

## Development

- **[Testing Guide](development/testing.md)** - Testing strategies and guidelines
- **[Build Guide](development/build.md)** - Build instructions for all platforms
- **[Session Summary](development/session-summary.md)** - Development history and status

## Features

- **[Feature Implementation](features/explain-dryrun.md)** - EXPLAIN and Dry Run implementation details

## Project Documentation

- **[CLAUDE.md](../CLAUDE.md)** - Complete project guide for AI assistants
- **[TEST_FAILURES.md](../TEST_FAILURES.md)** - Test failure analysis (now all resolved!)

## Documentation Index

### By Audience

**For Users:**
- [README](../README.md) - Overview and quick start
- [Query Features Guide](guides/query-features.md) - How to use EXPLAIN and Dry Run
- [Quick Reference](guides/quick-reference.md) - Quick lookup guide

**For Developers:**
- [Architecture - Flow](architecture/flow.md) - System architecture and flow
- [Architecture - Detailed Flow](architecture/detailed-flow.md) - Technical implementation details
- [Testing Guide](development/testing.md) - How to test the system
- [Build Guide](development/build.md) - How to build for different platforms

**For Contributors:**
- [CLAUDE.md](../CLAUDE.md) - Complete project context
- [Session Summary](development/session-summary.md) - Development history

### By Topic

**Query Execution:**
- [Flow Diagrams](architecture/flow.md) - How queries are executed
- [Technical Flow](architecture/detailed-flow.md) - Detailed step-by-step flow
- [Query Features](guides/query-features.md) - EXPLAIN and Dry Run features

**Development:**
- [Testing](development/testing.md) - Unit tests, integration tests, coverage
- [Building](development/build.md) - Multi-platform build guide
- [Features](features/explain-dryrun.md) - Feature implementation

**API & Endpoints:**
- See [CLAUDE.md](../CLAUDE.md) for complete API reference
- All endpoints documented with examples

---

## Project Status

**Backend:** ✅ ~85% Complete
- All infrastructure (database, models, auth, config)
- Query execution engine with SQL parser
- Approval workflow system
- Data source management with encryption
- Redis queue + background worker
- Google Chat notifications
- User & Group Management
- EXPLAIN and Dry Run features
- All API endpoints implemented

**Frontend:** 🚧 To Be Implemented
- Next.js + Tailwind CSS
- SQL editor with Monaco
- Query results display
- Approval dashboard
- Data source management UI

**Testing:** ✅ 90/90 Tests Passing (100%)
- Auth tests: 18/18 PASS
- Parser tests: 30/30 PASS
- Query Service tests: 21/21 PASS
- Models tests: 21/21 PASS

**Next Steps:**
1. Integration tests with real databases
2. CORS and logging middleware
3. Rate limiting middleware
4. Frontend development

---

## Quick Links

- **GitHub Repository:** [querybase](https://github.com/yourorg/querybase)
- **API Documentation:** See [CLAUDE.md](../CLAUDE.md)
- **Issue Tracker:** [GitHub Issues](https://github.com/yourorg/querybase/issues)
- **Contributing:** See [CLAUDE.md](../CLAUDE.md) for contribution guidelines

---

**Last Updated:** January 27, 2025
