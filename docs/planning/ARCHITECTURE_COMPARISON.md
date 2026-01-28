# Architecture Comparison: Current vs Planned
**System Architecture Gap Analysis**

**Date:** January 28, 2025
**Purpose:** Compare current architecture with planned frontend requirements
**Status:** Analysis Complete

---

## Executive Summary

**Current Architecture:** REST API with basic CRUD, query execution, and approval workflow.
**Planned Architecture:** Enhanced REST API with organization features, analytics, and optional WebSocket for real-time updates.

**Key Findings:**
- ✅ **8 new services** needed
- ✅ **6 new handlers** needed
- ✅ **4 new database tables** needed
- ✅ **WebSocket can be deferred** to future phases (not critical)
- ✅ **Polling workaround** viable for Phase 1 frontend

---

## Architecture Comparison Matrix

| Layer | Current Components | Planned Components | Gap | Priority |
|-------|-------------------|-------------------|-----|----------|
| **Frontend** | Next.js (Planned) | Next.js + Monaco + shadcn/ui | ❌ None | - |
| **API Gateway** | Gin + Auth + RBAC | Same + CORS/Logging/Rate Limit | ⚠️ Middleware | 🟡 Medium |
| **Handlers** | 5 handlers | 11 handlers (6 new) | ❌ 6 handlers | 🔴 Critical |
| **Services** | 4 services | 12 services (8 new) | ❌ 8 services | 🔴 Critical |
| **Database** | 13 tables | 17 tables (4 new) | ❌ 4 tables | 🔴 Critical |
| **Real-time** | Google Chat webhooks | Optional WebSocket + Polling | ⚠️ WebSocket optional | 🟢 Low |

---

## Detailed Component Analysis

### 1. Frontend Layer

**Current Status:**
```yaml
Status: Planned
Technology: Next.js 14/15
Components: Not started
```

**Planned:**
```yaml
Technology: Next.js + TypeScript + Tailwind + Monaco + shadcn/ui
Components:
  - SQL Editor (Monaco)
  - Query Results Table
  - Schema Browser
  - Saved Queries
  - Approval Dashboard
  - Data Source Management
```

**Gap:** ❌ None - Frontend is planned from scratch

**Timeline:** 12 weeks (see [DASHBOARD_UI_PLAN.md](DASHBOARD_UI_PLAN.md))

---

### 2. API Gateway (Middleware)

**Current State:**
```go
// internal/api/middleware/
├── auth.go          ✅ JWT authentication
└── rbac.go          ✅ Role-based access control

// TODO
├── cors.go          ❌ Not implemented
└── logging.go       ❌ Not implemented
```

**Planned Additions:**
```go
// internal/api/middleware/
├── auth.go          ✅ Existing
├── rbac.go          ✅ Existing
├── cors.go          ⚠️ TODO (Phase 3)
├── logging.go       ⚠️ TODO (Phase 3)
└── ratelimit.go     ❌ NEW (Phase 3)
```

**Gap:** ⚠️ **Middleware incomplete** but not blocking
- CORS: Can be added later (browsers handle preflight)
- Logging: Can use Gin's built-in logger temporarily
- Rate Limiting: Can rely on infrastructure (NGINX, load balancer)

**Recommendation:** Defer to Phase 3 (polish), not critical for Phase 1

---

### 3. Handler Layer

**Current Handlers (5):**
```go
// internal/api/handlers/
├── auth.go          ✅ Login, users CRUD, change password
├── query.go         ✅ Execute, list, save, delete, history
├── approval.go      ✅ List, review, transaction management
├── datasource.go    ✅ CRUD, test connection, permissions
└── group.go         ✅ CRUD, user assignment
```

**Planned Handlers (+6):**
```go
// internal/api/handlers/
├── auth.go          ✅ Existing
├── query.go         ✅ Existing (add pagination)
├── approval.go      ✅ Existing (add comments)
├── datasource.go    ✅ Existing
├── group.go         ✅ Existing
├── schema.go        ❌ NEW (Critical - Schema API)
├── folder.go        ❌ NEW (Critical - Folder CRUD)
├── tag.go           ❌ NEW (Main - Tag CRUD)
├── export.go        ❌ NEW (Main - Export API)
├── analytics.go     ❌ NEW (Main - Metrics API)
└── websocket.go     ❌ NEW (Optional - Real-time)
```

**Gap Analysis:**

| Handler | Purpose | Can Frontend Launch Without? | Priority |
|---------|---------|----------------------------|----------|
| **schema.go** | Autocomplete, schema browser | ❌ NO - Blocks SQL Editor | 🔴 Critical |
| **folder.go** | Query organization | ❌ NO - Blocks Saved Queries | 🔴 Critical |
| **tag.go** | Tagging queries | ⚠️ YES - Can use folders only | 🟡 Medium |
| **export.go** | Download results | ⚠️ YES - Can view in browser | 🟡 Medium |
| **analytics.go** | Performance metrics | ⚠️ YES - Nice to have | 🟡 Medium |
| **websocket.go** | Real-time updates | ✅ YES - Can use polling | 🟢 Low |

**Critical Path:** schema.go, folder.go (must have for Phase 1)

---

### 4. Service Layer

**Current Services (4):**
```go
// internal/service/
├── query.go         ✅ Query execution, EXPLAIN, dry run
├── approval.go      ✅ Approval workflow, transactions
├── datasource.go    ✅ Connection management
└── notification.go  ✅ Google Chat webhooks
```

**Planned Services (+8):**
```go
// internal/service/
├── query.go         ✅ Existing (add pagination)
├── approval.go      ✅ Existing (add comments)
├── datasource.go    ✅ Existing
├── notification.go  ✅ Existing
├── schema.go        ❌ NEW (Critical - Schema introspection)
├── folder.go        ❌ NEW (Critical - Folder management)
├── tag.go           ❌ NEW (Main - Tag management)
├── export.go        ❌ NEW (Main - CSV/JSON export)
├── comment.go       ❌ NEW (Main - Approval comments)
├── analytics.go     ❌ NEW (Main - Performance metrics)
├── formatter.go     ❌ NEW (Optional - SQL formatting)
└── comparison.go    ❌ NEW (Optional - Query comparison)
```

**Service Dependency Graph:**

```
query.go
    │
    ├─> Uses: schema.go (for autocomplete)
    │
folder.go
    │
    └─> Contains: queries (one-to-many)

tag.go
    │
    └─> Tags: queries (many-to-many)

approval.go
    │
    └─> Has: comments.go (one-to-many)

export.go
    │
    └─> Uses: query.go (fetch results)

analytics.go
    │
    └─> Uses: query.go (aggregate data)
```

**Gap Analysis:**

| Service | Dependencies | Can Frontend Work Without? | Priority |
|---------|-------------|--------------------------|----------|
| **schema.go** | None | ❌ NO | 🔴 Critical |
| **folder.go** | None | ❌ NO | 🔴 Critical |
| **tag.go** | None | ⚠️ YES | 🟡 Medium |
| **comment.go** | approval.go | ⚠️ YES | 🟡 Medium |
| **export.go** | query.go | ⚠️ YES | 🟡 Medium |
| **analytics.go** | query.go | ⚠️ YES | 🟡 Medium |
| **formatter.go** | None | ⚠️ YES | 🟢 Low |
| **comparison.go** | query.go | ⚠️ YES | 🟢 Low |

**Critical Path:** schema.go, folder.go (no dependencies, can start immediately)

---

### 5. Database Layer

**Current Tables (13):**
```sql
-- Existing tables
users                       ✅
groups                      ✅
user_groups                 ✅
data_sources                ✅
data_source_permissions     ✅
queries                     ✅
query_results               ✅
query_history               ✅
approval_requests           ✅
approval_reviews            ✅
query_transactions          ✅
notification_configs        ✅
notifications               ✅
```

**Planned Tables (+4):**
```sql
-- New tables
folders                     ❌ NEW (Critical)
tags                        ❌ NEW (Main)
query_tags                  ❌ NEW (Main)
approval_comments           ❌ NEW (Main)

-- Updated tables
queries                     ✅ UPDATE (add folder_id, is_favorite)
```

**Schema Changes Required:**

**Migration 000005:**
```sql
-- Folders (Critical)
CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Tags (Main)
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(7),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE query_tags (
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (query_id, tag_id)
);

-- Update queries table
ALTER TABLE queries ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;
ALTER TABLE queries ADD COLUMN is_favorite BOOLEAN DEFAULT FALSE;
```

**Migration 000006:**
```sql
-- Comments (Main)
CREATE TABLE approval_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Gap Analysis:**

| Table | Purpose | Can Frontend Launch Without? | Priority |
|-------|---------|----------------------------|----------|
| **folders** | Query organization | ❌ NO - Blocks Saved Queries | 🔴 Critical |
| **tags** | Additional organization | ⚠️ YES - Can use folders | 🟡 Medium |
| **query_tags** | Tag relationships | ⚠️ YES | 🟡 Medium |
| **approval_comments** | Approval collaboration | ⚠️ YES - Can approve without comments | 🟡 Medium |

**Critical Path:** folders table (must have for Phase 1)

---

### 6. Real-time Communication

**Current State:**
```yaml
Implementation: Google Chat Webhooks (HTTP POST)
Direction: Backend → External (one-way)
Use Case: Approval notifications
```

**Planned:**
```yaml
Option 1: WebSocket (Full duplex)
Option 2: Server-Sent Events (SSE)
Option 3: Polling (HTTP GET repeated)
```

**Comparison:**

| Approach | Pros | Cons | Complexity | Timeline |
|----------|------|------|------------|----------|
| **WebSocket** | Real-time, efficient | Complex, stateful, scaling issues | High | 4-5 days |
| **SSE** | Simpler, one-way | One-way only, browser limitations | Medium | 2-3 days |
| **Polling** | Simple, stateless | Not real-time, higher server load | Low | 1 day (frontend) |

**Recommendation: ✅ Defer WebSocket to Phase 3**

**Reasoning:**

1. **Not Critical for MVP**
   - Query status can be polled every 2-3 seconds
   - Approvals don't update that frequently
   - Transaction status is short-lived (< 5 minutes)

2. **Polling Workaround is Simple**
   ```typescript
   // Frontend polling implementation
   useEffect(() => {
     const interval = setInterval(async () => {
       const status = await fetchQueryStatus(queryId);
       if (status === 'completed') {
         clearInterval(interval);
         // Refresh results
       }
     }, 3000); // Poll every 3 seconds

     return () => clearInterval(interval);
   }, [queryId]);
   ```

3. **WebSocket Adds Complexity**
   - Connection management (reconnect, heartbeat)
   - Stateful servers (harder to scale horizontally)
   - Load balancer configuration (sticky sessions)
   - Authentication per connection
   - Error handling

4. **Performance Impact is Minimal**
   - Query status polling: 1 request every 3 seconds
   - For 100 concurrent users: 33 requests/second (negligible)
   - For 1,000 concurrent users: 333 requests/second (manageable)
   - Can add rate limiting if needed

5. **Can Add WebSocket Later**
   - Frontend can be designed to switch between polling/websocket
   - No breaking API changes
   - WebSocket is an optimization, not a requirement

**Proposed Architecture for Phase 1-2:**

```
Frontend
    │
    ├─> Polling: GET /api/v1/queries/:id (every 3 seconds)
    ├─> Polling: GET /api/v1/approvals (every 5 seconds)
    └─> Polling: GET /api/v1/transactions/:id (every 2 seconds when active)
```

**Proposed Architecture for Phase 3:**

```
Frontend ←WebSocket→ Hub
                ↓
         Backend Events
    (broadcast updates)
```

**Decision:** Start with polling, add WebSocket in Phase 3 if performance metrics show it's needed.

---

## Architecture Gap Summary

### Critical Gaps (Block Frontend) 🔴

**Must implement before frontend Phase 2 (SQL Editor):**

1. **Schema Service + Handler** (3-4 days)
   - Files: `internal/service/schema.go`, `internal/api/handlers/schema.go`
   - Endpoints: 5 new endpoints
   - Purpose: SQL autocomplete, schema browser

2. **Query Pagination** (2-3 days)
   - Files: Update `internal/service/query.go`, `internal/api/handlers/query.go`
   - Endpoints: 1 updated endpoint
   - Purpose: Display large result sets

3. **Folder System** (3-4 days)
   - Files: `internal/models/folder.go`, `internal/service/folder.go`, `internal/api/handlers/folder.go`
   - Endpoints: 6 new endpoints
   - Database: 1 new table
   - Purpose: Organize saved queries

**Total Critical Effort:** 8-11 days (2 weeks)

---

### Main Gaps (Degrade UX) 🟡

**Should implement during frontend Phase 3-4:**

4. **Tag System** (2-3 days)
5. **Comment System** (2-3 days)
6. **Export API** (1-2 days)
7. **Analytics/Metrics API** (2-3 days)

**Total Main Effort:** 7-11 days (2 weeks)

---

### Optional Gaps (Nice to Have) 🟢

**Can defer to Phase 5-6:**

8. **WebSocket Support** (4-5 days) - **CAN BE DEFERRED**
9. **SQL Formatter** (1-2 days)
10. **Favorites** (1 day)
11. **Health Check API** (1 day)
12. **Usage Statistics** (1-2 days)
13. **Bulk Operations** (1-2 days)
14. **Query Comparison** (2-3 days)

**Total Optional Effort:** 11-16 days (3 weeks)

---

## Updated Implementation Plan

### Phase 1: Critical Foundation (Week 1-2) 🔴

**Backend Team:**
- Day 1-4: Schema Service + Handler (5 endpoints)
- Day 5-7: Query Pagination (1 endpoint)
- Day 8-11: Folder System (migration + 6 endpoints)

**Frontend Team:**
- Can start Phase 1 (Foundation) in parallel
- Week 1: Project setup, auth, layout
- Week 2: Start SQL Editor (without autocomplete)

**Deliverables:**
- ✅ 12 new API endpoints
- ✅ 1 new database table
- ✅ Frontend unblocked for SQL Editor, Schema Browser, Saved Queries

---

### Phase 2: Core Features (Week 3-4) 🟡

**Backend Team:**
- Week 3: Tag System + Comment System
- Week 4: Export API + Table Statistics

**Frontend Team:**
- Week 3: SQL Editor (with autocomplete), Query Results
- Week 4: Schema Browser, Saved Queries (with folders)

**Real-time:** Use polling (3-5 second intervals)

**Deliverables:**
- ✅ 10 more API endpoints
- ✅ 3 more database tables
- ✅ Enhanced UX with tags, comments, export

---

### Phase 3: Polish & Optimization (Week 5-6) 🟢

**Backend Team:**
- Week 5: Analytics API, Health Check, Usage Stats
- Week 6: **Evaluate WebSocket** - if polling shows performance issues, implement WebSocket

**Frontend Team:**
- Week 5: Query History, Filters, Search
- Week 6: Performance optimization, testing, polish

**WebSocket Decision Point:**

```
IF (concurrent users > 500) OR (polling requests > 1000/sec):
    Implement WebSocket
ELSE:
    Continue with polling
    Document architecture decision
```

**Deliverables:**
- ✅ Full feature parity with Bytebase CE
- ✅ Performance metrics to guide WebSocket decision
- ✅ Production-ready frontend

---

## Reduced Scope Architecture

### What If We Skip WebSocket Entirely?

**Pros:**
- Simpler architecture
- Easier to scale (stateless servers)
- Lower complexity
- Faster time to market

**Cons:**
- 2-5 second delay on status updates
- Higher server load from polling
- Not "real-time"

**Verdict:** ✅ **Acceptable for MVP**

Most database tools (DBeaver, DataGrip, phpMyAdmin) don't have real-time updates and work fine. QueryBase's approval workflow is inherently async (human approval), so real-time updates aren't critical.

---

## Revised Priority Matrix

### Before Starting Frontend (Week 1-2) 🔴

| Feature | Effort | Files | Database |
|---------|--------|-------|----------|
| Schema API | 3-4 days | 3 files | 0 tables |
| Query Pagination | 2-3 days | 2 files (update) | 0 tables |
| Folder System | 3-4 days | 3 files | 1 table |

**Total:** 8-11 days, 8 files, 1 migration

### During Frontend Development (Week 3-4) 🟡

| Feature | Effort | Files | Database |
|---------|--------|-------|----------|
| Tag System | 2-3 days | 3 files | 2 tables |
| Comment System | 2-3 days | 2 files | 1 table |
| Export API | 1-2 days | 2 files | 0 tables |
| Table Statistics | 1-2 days | 2 files | 0 tables |

**Total:** 6-10 days, 9 files, 3 migrations

### After Frontend MVP (Week 5-6+) 🟢

| Feature | Effort | Files | Database |
|---------|--------|-------|----------|
| Analytics API | 2-3 days | 2 files | 0 tables |
| SQL Formatter | 1-2 days | 2 files | 0 tables |
| Favorites | 1 day | 1 file (update) | 0 columns |
| Health Check | 1 day | 2 files | 0 tables |
| Usage Stats | 1-2 days | 2 files | 0 tables |
| Bulk Operations | 1-2 days | 1 file (update) | 0 tables |
| Query Comparison | 2-3 days | 2 files | 0 tables |
| **WebSocket** | **4-5 days** | **2 files** | **0 tables** |

**Total:** 15-21 days, 14 files, 0 migrations

**WebSocket Note:** Can implement if polling performance is insufficient.

---

## Final Recommendation

### Minimum Viable Backend (MVB) for Frontend Launch

**Must Have (Week 1-2):**
1. ✅ Schema Introspection API
2. ✅ Query Pagination API
3. ✅ Folder System

**Should Have (Week 3-4):**
4. ⚠️ Tag System
5. ⚠️ Comment System
6. ⚠️ Export API

**Nice to Have (Week 5-6+):**
7. 🔄 Analytics API
8. 🔄 Table Statistics
9. 🔄 WebSocket (only if polling shows issues)
10. 🔄 SQL Formatter
11. 🔄 Favorites
12. 🔄 Health Check

**Total MVB Effort:** 8-11 days (2 weeks)

**Total Full Feature Effort:** 29-44 days (6-9 weeks)

---

## WebSocket Decision Framework

### Implement WebSocket IF:

```
✅ Polling requests exceed 500/second sustained
✅ Frontend users complain about latency
✅ Need > 10 concurrent real-time features
✅ Server CPU > 70% from polling overhead
✅ Mobile app requires push notifications
```

### Stick with Polling IF:

```
✅ Concurrent users < 500
✅ Polling requests < 200/second
✅ Updates are infrequent (> 5 second intervals acceptable)
✅ Simple stateless architecture preferred
✅ Want faster time to market
```

### Current QueryBase Profile:

```
Expected concurrent users (Year 1): 50-200
Expected polling load: 50-200 requests/second
Update frequency: Low (approval workflow is async)
Conclusion: ✅ POLLING IS SUFFICIENT
```

### When to Reconsider:

```
Year 2: 500+ concurrent users
Year 2: Mobile app launch
Year 2: Real-time collaboration features
Action: Re-evaluate WebSocket at that time
```

---

## Architecture Evolution Timeline

### Current Architecture (v0.8) ✅

```
Frontend (Planned)
    ↓ REST
API Gateway (Gin + Auth + RBAC)
    ↓
Services (4 services)
    ↓
PostgreSQL + Redis (Queue)
    ↓
User Databases
```

### Phase 1 Architecture (v0.9) - Week 2 🔴

```
Frontend (MVP - Polling)
    ↓ REST
API Gateway (Gin + Auth + RBAC)
    ↓
Handlers (11 handlers)
    ↓
Services (12 services)
    ↓
PostgreSQL (17 tables) + Redis (Cache + Queue)
    ↓
User Databases
```

**New Components:**
- Schema Service (autocomplete)
- Folder Service (organization)
- Pagination (query results)

### Phase 2 Architecture (v1.0) - Week 4 🟡

```
Frontend (Full Features - Polling)
    ↓ REST
API Gateway (Gin + Auth + RBAC)
    ↓
Handlers (11 handlers)
    ↓
Services (12 services)
    ↓
PostgreSQL (17 tables) + Redis (Cache + Queue)
    ↓
User Databases
```

**New Components:**
- Tag Service, Comment Service
- Export Service, Analytics Service

### Phase 3 Architecture (v1.1 or v2.0) - Week 6+ 🟢

**Option A: Keep Polling (v1.1)**
```
Frontend (Optimized Polling)
    ↓ REST
API Gateway + Rate Limiting
    ↓
All Services
    ↓
PostgreSQL + Redis + CDN
```

**Option B: Add WebSocket (v2.0)**
```
Frontend (WebSocket + Polling fallback)
    ├─> WebSocket (real-time)
    └─> REST (fallback)
API Gateway + WebSocket Hub
    ↓
All Services + Broadcast Layer
    ↓
PostgreSQL + Redis + Message Queue
```

**Decision Point:** End of Week 5, based on performance metrics

---

## Summary

### Critical Gaps (Must Fix) 🔴

| Component | Current | Planned | Gap | Effort |
|-----------|---------|---------|-----|--------|
| **Schema API** | ❌ None | ✅ 5 endpoints | ❌ Missing | 3-4 days |
| **Pagination** | ❌ None | ✅ 1 endpoint | ❌ Missing | 2-3 days |
| **Folders** | ❌ None | ✅ 6 endpoints | ❌ Missing | 3-4 days |

### Main Gaps (Should Fix) 🟡

| Component | Current | Planned | Gap | Effort |
|-----------|---------|---------|-----|--------|
| **Tags** | ❌ None | ✅ 5 endpoints | ❌ Missing | 2-3 days |
| **Comments** | ❌ None | ✅ 4 endpoints | ❌ Missing | 2-3 days |
| **Export** | ❌ None | ✅ 2 endpoints | ❌ Missing | 1-2 days |
| **Analytics** | ❌ None | ✅ 2 endpoints | ❌ Missing | 2-3 days |

### Optional Gaps (Can Defer) 🟢

| Component | Current | Planned | Gap | Effort | Decision |
|-----------|---------|---------|-----|--------|----------|
| **WebSocket** | ❌ None | ✅ 4 topics | ❌ Missing | 4-5 days | ✅ **DEFER** - Use polling instead |
| **Formatter** | ❌ None | ✅ 1 endpoint | ❌ Missing | 1-2 days | Phase 3 |
| **Favorites** | ❌ None | ✅ 3 endpoints | ❌ Missing | 1 day | Phase 3 |
| **Health Check** | ❌ None | ✅ 1 endpoint | ❌ Missing | 1 day | Phase 3 |

### Recommendation: ✅ WebSocket to Future Plans

**Rationale:**
1. Polling is sufficient for 50-500 concurrent users
2. Adds 4-5 days of complexity
3. Harder to scale (stateful servers)
4. Can be added later if needed (no breaking changes)
5. Most DB tools don't use real-time updates

**Decision:** **Defer WebSocket to Phase 3 or later**

**Alternative:** Implement only if performance metrics show polling overhead > 30% CPU or > 500 requests/second.

---

**Last Updated:** January 28, 2025
**Status:** Ready for Implementation
**Next Step:** Begin Schema API implementation (Phase 1, Day 1)
