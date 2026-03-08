## 🧠 Brainstorm: Reducing Dashboard API Logs & DB Load

### Context

When a user opens the dashboard, multiple requests to `/api/v1/dashboard/stats` are logged simultaneously. While we've solved the backend database "Thundering Herd" with Singleflight and O(1) targeted Redis Cache invalidation, the frontend structure still mounts the `useDashboardStats` hook concurrently or multiple times, generating heavy network and logging noise.

---

### Option A: Silence `/api/v1/dashboard/stats` in Backend HTTP Logs

Since this is a high-traffic telemetry endpoint meant to maintain realtime metrics, we can configure the `gin.Logger` middleware to ignore or skip logging for this specific endpoint, while leaving the current React fetching pattern undisturbed.

✅ **Pros:**

- Cleans up the terminal/server logs immediately.
- Zero risk of breaking frontend react state.
- Extremely fast to implement.

❌ **Cons:**

- Doesn't stop the actual network requests hitting the server.
- The Singleflight protects the DB, but Gin still has to parse HTTP overhead across multiple requests.

📊 **Effort:** Low

---

### Option B: Lift `useDashboardStats` State to Global Zustand Store

Right now, if multiple components use `useDashboardStats()`, each component instance fetches the data on mount simultaneously. We can lift the data fetching responsibility into a global Zustand store (e.g. `useStatsStore`). The AppLayout would fetch it _once_ on load, and all children read from the reactive state.

**Realtime Behavior:** NO page refresh needed. The Zustand store will still listen to the WebSocket `"stats_changed"` event. When the event fires, the store triggers one background `fetch()`, updates its state, and all UI components blindly re-render instantly.

✅ **Pros:**

- Structurally solves the frontend issue.
- Guarantees exactly 1 network request per page load regardless of how many components need the stats.
- Clean architectural pattern.

❌ **Cons:**

- Requires refactoring `useDashboardStats.ts` into a Zustand store.
- Requires updating `dashboard/page.tsx` and anywhere else it's imported.

📊 **Effort:** Medium

---

### Option C: Add SWR / React Query for Deduplication

Replace the custom `useEffect` data fetching in `useDashboardStats` with a library like SWR (`useSWR`) or React Query. These libraries automatically deduplicate identical requests fired in the same render cycle and handle background revalidation caching cleanly.

**Realtime Behavior:** NO page refresh needed. Inside the custom hook, we would map the WebSocket `"stats_changed"` event to trigger SWR's `mutate()` method. This tells SWR to quietly fetch updated data in the background and instantly mutate the UI.

✅ **Pros:**

- Next.js standard approach for data fetching.
- Built-in request deduplication, caching, and retry logic.
- Minimal backend changes needed.

❌ **Cons:**

- Adds a new dependency to the project (`swr` or `@tanstack/react-query`) if not already present.
- Might require adjusting the global WebSocket invalidation logic to trigger an SWR revalidation instead of a manual refetch.

📊 **Effort:** Medium

---

## 💡 Recommendation

**Option B (Zustand Global Store)** is recommended because we already have `auth-store` and `theme-store` implemented using Zustand, keeping our state-management pattern consistent without adding new libraries. This guarantees only 1 request happens on the frontend.

Alternatively, if you simply want the console clean but don't mind the fast cached API hits, **Option A** is the quickest fix.

What direction would you like to explore?
