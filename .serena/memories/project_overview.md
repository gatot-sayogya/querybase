# QueryBase Project Overview

## Purpose
QueryBase is a web-based database exploration platform that allows users to execute SQL queries on PostgreSQL and MySQL databases with an approval workflow for write operations.

## Tech Stack

### Backend
- **Language:** Go 1.22+
- **Framework:** Gin (HTTP router)
- **ORM:** GORM
- **Database:** PostgreSQL 15 (primary)
- **Cache/Queue:** Redis 7 (Asynq job queue)
- **Auth:** JWT (golang-jwt/jwt v5)
- **Password Hashing:** bcrypt
- **Config:** Viper (YAML + env vars)

### Frontend
- **Framework:** Next.js 15+ (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS 3.x
- **Editor:** Monaco Editor (SQL autocomplete)
- **State Management:** Zustand
- **HTTP Client:** Axios

### DevOps
- **Containerization:** Docker & Docker Compose
- **Process Management:** Makefiles
- **Migrations:** Manual SQL migrations
- **Testing:** Go testing + Jest + Playwright (E2E)

## Architecture
- **API Server:** Go/Gin at localhost:8080
- **Background Worker:** Processes jobs from Redis queue
- **Frontend:** Next.js at localhost:3000
- **Primary DB:** PostgreSQL for application data
- **Queue:** Redis for background jobs
- **Data Sources:** PostgreSQL and MySQL databases to query

## Key Features
1. Query execution with SELECT running immediately
2. Write operations (INSERT/UPDATE/DELETE) with 3-phase approval workflow
3. Schema browser with auto-refresh
4. User/group management with RBAC
5. Google Chat notifications for approvals
6. Query history and export (CSV/JSON)
7. Monaco Editor with SQL autocomplete

## Project Structure
- `cmd/api/` - API server entry point
- `cmd/worker/` - Background worker entry point
- `internal/` - Private Go code (api, auth, config, database, models, queue, service, validation)
- `web/` - Next.js frontend
- `migrations/` - Database migrations
- `tests/` - Test files
