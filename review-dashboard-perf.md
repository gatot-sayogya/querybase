# Dashboard and Status UI Performance Review

## Findings

During the review of the dashboard and status components for high Database and Redis load, two primary issues were discovered in the Backend's `StatsService`:

### 1. Redis O(N) `SCAN` Operations

Every time a query was executed (which triggers a history update) or an approval request changed status (which changes the queue counts), `s.statsService.TriggerStatsChanged()` was invoked.

- **The Issue:** `TriggerStatsChanged` used a Redis `SCAN` command matching `"stats:dashboard:user:*"` to find and delete every single user's cached dashboard stats. Because Redis is single-threaded, a global `SCAN` on every query execution blocks Redis and adds severe, unnecessary latency, essentially degrading to O(N) performance where N is the number of active users.
- **The Solution:** The method signature was changed to `TriggerStatsChanged(userID string)`. The backend now only deletes the specific executing `userID`'s cache along with the `global` admin cache. This brings cache invalidation to an O(1) targeted operation.

### 2. Database "Thundering Herd" via WebSockets

When `TriggerStatsChanged` was invoked, the backend immediately broadcasted a `"stats_changed"` WebSocket event to all connected dashboard clients.

- **The Issue:** Upon receiving this broadcast, every single client immediately hit the `/api/v1/dashboard/stats` endpoint. Because the cache was just completely emptied by the `SCAN` operation above, all API requests resulted in a cache miss. This forced the Database to concurrently compute `COUNT()` aggregations across multiple tables (`query_history`, `approval_requests`, etc.) for every connected user at the exact same millisecond. This combination creates a classic "thundering herd" problem capable of bringing down the primary database under load.
- **The Solution:** We introduced the `golang.org/x/sync/singleflight` pattern to the `GetDashboardStats` execution flow. Now, if 100 users hit the `/stats` endpoint concurrently during a cache miss, the `CalculateStats` database aggregations will only run **once** per cache key, and the resulting payload will be shared with all 100 requesting goroutines, drastically limiting the peak simultaneous load on the database.

## Validation

With the targeted cache invalidation and Singleflight implemented:

1. Normal users writing queries will only invalidate their own cache, leaving other active operators' cached dashboards untouched.
2. The Database is protected from concurrent metric-aggregate sweeps even during mass cache invalidations or server restarts.

No heavy pooling issues were observed on the Next.js Frontend — the web application correctly aggregates API hits into a single `useDashboardStats` hook relying on reactive WebSocket invalidation rather than blindly `setInterval` polling.

The application compiles perfectly following these core logic refactoring optimizations.
