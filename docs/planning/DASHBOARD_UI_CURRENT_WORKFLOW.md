# Dashboard UI Plan - Current Workflow Focus
**Build Frontend for Existing Backend (No New Backend APIs Needed)**

**Date:** January 28, 2025
**Status:** Ready to Start
**Approach:** Build UI for current backend capabilities
**Timeline:** 6-8 weeks to production-ready frontend

---

## Executive Summary

**Strategic Decision:** Build Dashboard UI for **current backend capabilities**, not planned features.

**Current Backend Status:**
- ✅ Authentication (JWT, login, password change)
- ✅ Query execution (SELECT, INSERT, UPDATE, DELETE, EXPLAIN, dry run)
- ✅ Query results storage (JSONB, with history)
- ✅ Approval workflow (transaction-based preview)
- ✅ Data source management (PostgreSQL, MySQL)
- ✅ User & group management
- ✅ RBAC (permissions, roles)
- ✅ Google Chat notifications

**What's NOT Needed (Deferred):**
- ❌ Schema API (use simple table list)
- ❌ Folder System (use search/filter)
- ❌ Tag System (use search/filter)
- ❌ WebSocket (use polling)

**Frontend Features to Build:**
1. **Authentication UI** (login, password change)
2. **SQL Editor** (Monaco, basic features)
3. **Query Results** (table view, pagination)
4. **Saved Queries** (flat list with search)
5. **Query History** (searchable list)
6. **Approval Dashboard** (review, approve/reject)
7. **Data Source Management** (admin only)
8. **User/Group Management** (admin only)

**Timeline:** 6-8 weeks

---

## Feature Prioritization

### Phase 1: Foundation (Week 1-2)

**Goal:** Set up project and authentication

**Features:**
1. Project setup (Next.js, Tailwind, shadcn/ui)
2. Authentication flow (login, JWT handling)
3. Layout components (sidebar, header)
4. API client setup
5. Protected routes

**Deliverables:**
- ✅ Users can login
- ✅ JWT token stored and used
- ✅ Protected pages work

---

### Phase 2: SQL Editor & Results (Week 3-4)

**Goal:** Core query functionality

**Features:**
1. SQL Editor (Monaco integration)
2. Data source selector
3. Run query button
4. Query results table
5. Result pagination (frontend-side)
6. Status indicators (running, completed, failed)
7. Error display

**APIs Used:**
```
POST /api/v1/queries              # Execute query
GET  /api/v1/datasources           # List data sources
GET  /api/v1/queries/:id           # Get query with results
GET  /api/v1/queries/history       # Get history (paginated)
```

**Workarounds for Missing Features:**
- **No Schema API:** Show simple table list from data source list
- **No Pagination API:** Client-side pagination (fetch all, paginate in browser)
- **No Autocomplete:** Text-based search only

**Deliverables:**
- ✅ Users can write and execute SQL queries
- ✅ Users can view query results
- ✅ Users can see query history

---

### Phase 3: Approval Dashboard (Week 5)

**Goal:** Approval workflow UI

**Features:**
1. Approval request list (with filters)
2. Approval detail view
3. SQL query display
4. Dry run results (for DELETE queries)
5. Approve/Reject buttons
6. Transaction status display
7. Comment discussion (if backend ready)

**APIs Used:**
```
GET    /api/v1/approvals                    # List approvals
GET    /api/v1/approvals/:id                # Get approval details
POST   /api/v1/approvals/:id/review         # Approve/reject
POST   /api/v1/approvals/:id/transaction-start  # Start transaction
POST   /api/v1/transactions/:id/commit      # Commit transaction
POST   /api/v1/transactions/:id/rollback    # Rollback transaction
```

**Polling Strategy:**
```typescript
// Poll transaction status every 2 seconds
const pollTransactionStatus = async (transactionId: string) => {
  const interval = setInterval(async () => {
    const status = await fetchTransactionStatus(transactionId);
    if (status === 'completed' || status === 'rolled_back') {
      clearInterval(interval);
      // Refresh UI
    }
  }, 2000);
};
```

**Deliverables:**
- ✅ Approvers can view pending approvals
- ✅ Approvers can review and approve/reject
- ✅ Transaction preview works
- ✅ Real-time status updates (via polling)

---

### Phase 4: Data Source & User Management (Week 6)

**Goal:** Admin features

**Features:**
1. Data source list (admin only)
2. Add/edit/delete data source (admin only)
3. Test connection button
4. Permission management (admin only)
5. User list (admin only)
6. Create/edit/delete users (admin only)
7. Group list (admin only)
8. Group management (admin only)

**APIs Used:**
```
# Data Sources
GET    /api/v1/datasources                   # List data sources
POST   /api/v1/datasources                   # Create (admin)
PUT    /api/v1/datasources/:id               # Update (admin)
DELETE /api/v1/datasources/:id               # Delete (admin)
POST   /api/v1/datasources/:id/test          # Test connection
PUT    /api/v1/datasources/:id/permissions   # Set permissions (admin)

# Users
GET    /api/v1/auth/users                    # List users (admin)
POST   /api/v1/auth/users                    # Create user (admin)
PUT    /api/v1/auth/users/:id                # Update user (admin)
DELETE /api/v1/auth/users/:id                # Delete user (admin)

# Groups
GET    /api/v1/groups                        # List groups (admin)
POST   /api/v1/groups                        # Create group (admin)
PUT    /api/v1/groups/:id                    # Update group (admin)
DELETE /api/v1/groups/:id                    # Delete group (admin)
POST   /api/v1/groups/:id/users              # Add user to group
DELETE /api/v1/groups/:id/users              # Remove user from group
```

**Deliverables:**
- ✅ Admins can manage data sources
- ✅ Admins can manage users
- ✅ Admins can manage groups

---

### Phase 5: Polish & Optimization (Week 7-8)

**Goal:** Production-ready quality

**Features:**
1. Error handling and display
2. Loading states
3. Empty states
4. Keyboard shortcuts
5. Responsive design (tablet support)
6. Performance optimization
7. Browser testing
8. Accessibility audit

**Deliverables:**
- ✅ Production-ready frontend
- ✅ Good UX (loading, errors, empty states)
- ✅ Performance < 2s page load
- ✅ Accessibility WCAG 2.1 AA

---

## Technology Stack

### Frontend Framework
**Next.js 14/15** (App Router)
```bash
npx create-next-app@latest querybase-web --typescript --tailwind --app
```

### UI Component Library
**shadcn/ui** + **Radix UI**
```bash
npx shadcn-ui@latest init
npx shadcn-ui@latest add button
npx shadcn-ui@latest add input
npx shadcn-ui@latest add table
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add dropdown-menu
# etc.
```

### Editor
**Monaco Editor** (VS Code's editor)
```bash
npm install @monaco-editor/react monaco-editor
```

### Data Fetching
**TanStack Query** (React Query)
```bash
npm install @tanstack/react-query
```

### State Management
**Zustand**
```bash
npm install zustand
```

### Additional Libraries
```json
{
  "dependencies": {
    "axios": "^1.6.0",
    "date-fns": "^3.0.0",
    "react-table": "^8.11.0",
    "recharts": "^2.10.0"
  }
}
```

---

## Component Structure

### Page Structure (Next.js App Router)

```
app/
├── (auth)/
│   ├── login/
│   │   └── page.tsx              # Login page
│   └── layout.tsx               # Auth layout
│
├── (dashboard)/
│   ├── page.tsx                  # Dashboard home
│   ├── layout.tsx                # Dashboard layout (with sidebar)
│   │
│   ├── editor/
│   │   ├── page.tsx              # SQL editor
│   │   └── [id]/
│   │       └── page.tsx          # Saved query edit
│   │
│   ├── queries/
│   │   ├── page.tsx              # Saved queries
│   │   └── [id]/
│   │       └── page.tsx          # Query detail
│   │
│   ├── history/
│   │   └── page.tsx              # Query history
│   │
│   ├── approvals/
│   │   ├── page.tsx              # Approval dashboard
│   │   └── [id]/
│   │       └── page.tsx          # Approval detail
│   │
│   ├── datasources/
│   │   ├── page.tsx              # Data source list (admin)
│   │   └── [id]/
│   │       └── page.tsx          # Data source detail
│   │
│   ├── users/
│   │   └── page.tsx              # User management (admin)
│   │
│   └── groups/
│       └── page.tsx              # Group management (admin)
│
└── api/                          # API routes (BFF pattern)
    └── proxy/[...path]/route.ts  # Proxy to backend
```

### Component Hierarchy

```
components/
├── layout/
│   ├── Sidebar.tsx              # Left navigation
│   ├── Header.tsx               # Top bar
│   └── Footer.tsx               # Footer
│
├── editor/
│   ├── SQLEditor.tsx           # Monaco editor wrapper
│   ├── EditorToolbar.tsx        # Run, Save, Explain buttons
│   └── QueryStatusBar.tsx       # Status indicators
│
├── results/
│   ├── ResultsTable.tsx         # Query results display
│   ├── Pagination.tsx           # Client-side pagination
│   └── EmptyState.tsx          # Empty results display
│
├── queries/
│   ├── QueryCard.tsx            # Saved query card
│   ├── QueryList.tsx            # Saved queries list
│   └── SearchFilter.tsx        # Search and filter controls
│
├── approvals/
│   ├── ApprovalCard.tsx         # Approval request card
│   ├── ReviewDialog.tsx         # Approve/reject dialog
│   ├── DryRunPreview.tsx       # DELETE preview results
│   └── TransactionStatus.tsx    # Transaction state display
│
├── datasources/
│   ├── DataSourceCard.tsx       # Data source card
│   ├── DataSourceForm.tsx       # Add/edit form
│   └── PermissionMatrix.tsx     # Permission grid
│
├── users/
│   ├── UserTable.tsx            # User list
│   └── UserForm.tsx             # Create/edit user
│
├── common/
│   ├── Button.tsx               # shadcn/ui button wrapper
│   ├── Input.tsx                # shadcn/ui input wrapper
│   ├── Dialog.tsx               # Dialog modal
│   ├── Dropdown.tsx             # Dropdown menu
│   ├── Badge.tsx                # Status badges
│   ├── Spinner.tsx              # Loading states
│   └── ErrorDisplay.tsx         # Error display
│
└── lib/
    ├── api-client.ts            # Axios API client
    ├── auth.ts                  # Auth utilities (JWT handling)
    └── hooks.ts                 # Custom React hooks
```

---

## Detailed Feature Implementation

### 1. Authentication UI (Week 1)

**Login Page:**
```typescript
// app/(auth)/login/page.tsx
export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const router = useRouter();

  const handleLogin = async (e: FormEvent) => {
    e.preventDefault();

    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    if (response.ok) {
      const { token } = await response.json();
      localStorage.setItem('token', token);
      router.push('/dashboard/editor');
    }
  };

  return (
    <form onSubmit={handleLogin}>
      <Input value={email} onChange={(e) => setEmail(e.target.value)} />
      <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
      <Button type="submit">Login</Button>
    </form>
  );
}
```

**Protected Route Component:**
```typescript
// components/layout/ProtectedRoute.tsx
export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token');
  const router = useRouter();

  if (!token) {
    router.push('/login');
    return null;
  }

  return <>{children}</>;
}
```

---

### 2. SQL Editor (Week 3)

**Monaco Integration:**
```typescript
// components/editor/SQLEditor.tsx
import Editor from '@monaco-editor/react';

export function SQLEditor({ dataSourceId, onSave, onExecute }: SQLEditorProps) {
  const [query, setQuery] = useState('');

  return (
    <div className="border rounded-lg">
      <Editor
        height="400px"
        defaultLanguage="sql"
        value={query}
        onChange={(value) => setQuery(value || '')}
        theme="vs-dark"
        options={{
          minimap: { enabled: false },
          fontSize: 14,
          scrollBeyondLastLine: false,
          automaticLayout: true,
        }}
      />
      <EditorToolbar
        query={query}
        onExecute={onExecute}
        onSave={onSave}
      />
    </div>
  );
}
```

**Execute Query:**
```typescript
const executeQuery = async (queryText: string) => {
  const response = await fetch('/api/v1/queries', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      dataSourceId: selectedDataSource,
      queryText: queryText,
      save: true,
    }),
  });

  const result = await response.json();

  if (result.requiresApproval) {
    // Redirect to approval detail
    router.push(`/approvals/${result.approvalId}`);
  } else {
    // Show results
    setQueryResult(result);
  }
};
```

---

### 3. Query Results (Week 3-4)

**Results Table with Client-Side Pagination:**
```typescript
// components/results/ResultsTable.tsx
export function ResultsTable({ result }: { result: QueryResult }) {
  const [page, setPage] = useState(1);
  const pageSize = 100;

  // Client-side pagination
  const paginatedRows = useMemo(() => {
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return result.rows.slice(start, end);
  }, [result.rows, page]);

  const totalPages = Math.ceil(result.rows.length / pageSize);

  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            {result.columns.map((col) => (
              <TableHead key={col}>{col}</TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {paginatedRows.map((row, i) => (
            <TableRow key={i}>
              {result.columns.map((col) => (
                <TableCell key={col}>
                  {formatCellValue(row[col])}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>

      <Pagination
        page={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />
    </div>
  );
}
```

**NOTE:** This uses client-side pagination. When backend pagination API is ready (in future), this component can be updated to use server-side pagination instead.

---

### 4. Saved Queries (Week 4)

**Simple List with Search (No Folders):**
```typescript
// app/(dashboard)/queries/page.tsx
export default function SavedQueriesPage() {
  const [queries, setQueries] = useState<Query[]>([]);
  const [search, setSearch] = useState('');

  useEffect(() => {
    fetchQueries();
  }, []);

  const filteredQueries = useMemo(() => {
    return queries.filter(q =>
      q.name.toLowerCase().includes(search.toLowerCase()) ||
      q.queryText.toLowerCase().includes(search.toLowerCase())
    );
  }, [queries, search]);

  return (
    <div>
      <Input
        placeholder="Search queries..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />

      <div className="grid gap-4">
        {filteredQueries.map((query) => (
          <QueryCard
            key={query.id}
            query={query}
            onExecute={handleExecuteQuery}
            onDelete={handleDeleteQuery}
          />
        ))}
      </div>
    </div>
  );
}
```

**NOTE:** This uses a flat list with search/filter. When Folder System is ready (in future), this can be updated to use folder tree structure.

---

### 5. Approval Dashboard (Week 5)

**Approval List with Polling:**
```typescript
// app/(dashboard)/approvals/page.tsx
export default function ApprovalsPage() {
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const [filter, setFilter] = useState<'pending' | 'all'>('pending');

  // Poll for updates every 5 seconds
  useEffect(() => {
    fetchApprovals();

    const interval = setInterval(fetchApprovals, 5000);
    return () => clearInterval(interval);
  }, [filter]);

  return (
    <div>
      <Tabs value={filter} onValueChange={setFilter}>
        <TabsList>
          <TabsTrigger value="pending">Pending</TabsTrigger>
          <TabsTrigger value="all">All</TabsTrigger>
        </TabsList>
      </Tabs>

      <div className="grid gap-4">
        {approvals.map((approval) => (
          <ApprovalCard
            key={approval.id}
            approval={approval}
            onReview={handleReview}
          />
        ))}
      </div>
    </div>
  );
}
```

**Transaction Status Polling:**
```typescript
// Poll transaction status when active
const useTransactionPolling = (transactionId: string) => {
  const [status, setStatus] = useState<TransactionStatus>('pending');

  useEffect(() => {
    if (!transactionId) return;

    const poll = async () => {
      const response = await fetch(`/api/v1/transactions/${transactionId}`);
      const data = await response.json();

      setStatus(data.status);

      // Stop polling when transaction is complete
      if (data.status === 'committed' || data.status === 'rolled_back') {
        clearInterval(interval);
      }
    };

    const interval = setInterval(poll, 2000);
    poll(); // Initial fetch

    return () => clearInterval(interval);
  }, [transactionId]);

  return status;
};
```

---

### 6. Data Source Management (Week 6)

**Data Source List (Admin Only):**
```typescript
// app/(dashboard)/datasources/page.tsx
export default function DataSourcesPage() {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);

  return (
    <div className="container">
      <div className="flex justify-between items-center">
        <h1>Data Sources</h1>
        <Button onClick={() => setShowCreateDialog(true)}>
          Add Data Source
        </Button>
      </div>

      <div className="grid gap-4 mt-6">
        {dataSources.map((ds) => (
          <DataSourceCard
            key={ds.id}
            dataSource={ds}
            onEdit={() => handleEdit(ds)}
            onDelete={() => handleDelete(ds.id)}
            onTest={() => handleTestConnection(ds.id)}
          />
        ))}
      </div>
    </div>
  );
}
```

---

## Workarounds for Missing Features

### 1. No Schema API → Simple Table List

**Problem:** Can't show database schema (tables, columns, types)

**Solution:** Show simple list from data source connection

**Implementation:**
```typescript
// When user selects data source, fetch basic info
const dataSourceInfo = await fetch(`/api/v1/datasources/${id}`);

// Show in sidebar as simple list
<div className="text-sm">
  <div className="font-semibold">Tables:</div>
  <div className="text-gray-500">
    (Schema browser coming soon)
  </div>
  <div className="mt-2 text-xs text-gray-400">
    Tip: Start your query with "SELECT * FROM table_name"
  </div>
</div>
```

### 2. No Autocomplete → Text Search

**Problem:** Can't provide table/column autocomplete in SQL editor

**Solution:** User types manually, show search in saved queries

**Future Enhancement:**
- Add simple text search in saved queries
- Show "Recently used tables" based on query history
- When Schema API is ready, add Monaco autocomplete provider

### 3. No Folder System → Search + Filter

**Problem:** Can't organize queries into folders

**Solution:** Flat list with powerful search and filters

**Implementation:**
```typescript
// Search by name, SQL, date range, data source
const filteredQueries = queries.filter(q => {
  const matchesSearch = searchQuery === '' ||
    q.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    q.queryText.toLowerCase().includes(searchQuery.toLowerCase());

  const matchesDataSource = !dataSourceFilter || q.dataSourceId === dataSourceFilter;
  const matchesDateRange = isWithinDateRange(q.createdAt, dateRange);

  return matchesSearch && matchesDataSource && matchesDateRange;
});
```

### 4. No Pagination API → Client-Side Pagination

**Problem:** Backend returns all rows at once

**Solution:** Paginate in browser (works for < 10,000 rows)

**Implementation:**
```typescript
// Fetch all results once
const { rows } = queryResult;

// Paginate client-side
const [page, setPage] = useState(1);
const pageSize = 100;

const paginatedRows = rows.slice((page - 1) * pageSize, page * pageSize);

// Future: When backend pagination API is ready, replace with:
// const { rows, pagination } = await fetch(`/api/v1/queries/${id}/results?page=${page}`);
```

### 5. No WebSocket → Polling

**Problem:** Can't get real-time updates

**Solution:** Poll every 3-5 seconds

**Performance Analysis:**
- 100 users × 1 poll/3 sec = 33 requests/second (negligible)
- 500 users × 1 poll/3 sec = 166 requests/second (manageable)

**Implementation:** See approval dashboard example above

---

## API Client Setup

**Axios with JWT Interceptor:**
```typescript
// lib/api-client.ts
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 30000,
});

// Add JWT token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 errors (unauthorized)
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

---

## State Management (Zustand)

**Query Store:**
```typescript
// lib/store/queryStore.ts
import create from 'zustand';

interface QueryStore {
  selectedDataSource: string | null;
  setSelectedDataSource: (id: string) => void;

  currentQuery: string;
  setCurrentQuery: (query: string) => void;

  queryResult: QueryResult | null;
  setQueryResult: (result: QueryResult) => void;

  isExecuting: boolean;
  setIsExecuting: (executing: boolean) => void;

  executeQuery: async (queryText: string) => Promise<void>;
}

export const useQueryStore = create<QueryStore>((set, get) => ({
  selectedDataSource: null,
  setSelectedDataSource: (id) => set({ selectedDataSource: id }),

  currentQuery: '',
  setCurrentQuery: (query) => set({ currentQuery: query }),

  queryResult: null,
  setQueryResult: (result) => set({ queryResult: result }),

  isExecuting: false,
  setIsExecuting: (executing) => set({ isExecuting: executing }),

  executeQuery: async (queryText) => {
    const { selectedDataSource } = get();
    set({ isExecuting: true });

    try {
      const response = await apiClient.post('/api/v1/queries', {
        dataSourceId: selectedDataSource,
        queryText,
        save: true,
      });

      set({
        queryResult: response.data,
        isExecuting: false,
      });
    } catch (error) {
      set({ isExecuting: false });
      throw error;
    }
  },
}));
```

---

## Testing Strategy

### Unit Tests

```typescript
// Components/__tests__/QueryCard.test.tsx
describe('QueryCard', () => {
  it('renders query name and SQL', () => {
    const query = {
      id: '1',
      name: 'Active Users',
      queryText: 'SELECT * FROM users WHERE active = true',
    };

    render(<QueryCard query={query} />);

    expect(screen.getByText('Active Users')).toBeInTheDocument();
    expect(screen.getByText(/SELECT \* FROM users/)).toBeInTheDocument();
  });
});
```

### Integration Tests (Playwright)

```typescript
// tests/e2e/query-flow.spec.ts
test('execute query flow', async ({ page }) => {
  // Login
  await page.goto('/login');
  await page.fill('[name="email"]', 'admin@querybase.local');
  await page.fill('[name="password"]', 'admin123');
  await page.click('button[type="submit"]');

  // Navigate to editor
  await page.goto('/editor');
  await page.waitForURL('/editor');

  // Select data source
  await page.selectOption('[name="dataSource"]', '1');

  // Type query
  await page.fill('[name="query"]', 'SELECT * FROM users LIMIT 10');

  // Click run
  await page.click('button:has-text("Run")');

  // Wait for results
  await page.waitForSelector('[data-testid="results-table"]');

  // Verify results
  const rows = await page.locator('[data-testid="results-table"] tbody tr').count();
  expect(rows).toBeGreaterThan(0);
});
```

---

## Performance Considerations

### Client-Side Pagination Limit

**Current Approach:** Paginate in browser
- Works for: < 10,000 rows
- Breakdown at: > 50,000 rows (browser memory)

**Solution:** Show warning for large results
```typescript
if (result.rows.length > 10000) {
  return (
    <Alert>
      <AlertTitle>Large Result Set</AlertTitle>
      <AlertDescription>
        This query returned {result.rows.length} rows.
        Showing first {pageSize} rows.
        Please add LIMIT clause or WHERE filter for better performance.
      </AlertDescription>
    </Alert>
  );
}
```

### Polling Optimization

**Smart Polling:** Only poll when necessary
```typescript
// Only poll approvals on approval pages
// Only poll transaction status when transaction is active
// Stop polling when user navigates away

useEffect(() => {
  if (!isActiveTransaction) return;  // Don't poll if no transaction

  const interval = setInterval(pollStatus, 2000);
  return () => clearInterval(interval);
}, [isActiveTransaction]);
```

---

## Implementation Timeline

### Week 1-2: Foundation
- Project setup
- Authentication
- Layout components
- API client

### Week 3-4: SQL Editor & Results
- Monaco integration
- Query execution
- Results display
- Query history

### Week 5: Approvals
- Approval dashboard
- Transaction preview
- Review workflow
- Polling for status

### Week 6: Admin Features
- Data source management
- User management
- Group management

### Week 7-8: Polish
- Error handling
- Loading states
- Performance optimization
- Testing
- Documentation

---

## Success Metrics

### User Experience
- ✅ Login to first query: < 30 seconds
- ✅ Query execution perceived latency: < 500ms
- ✅ Page load time: < 2 seconds
- ✅ No page crashes during normal usage

### Technical
- ✅ Lighthouse Performance score: > 90
- ✅ Lighthouse Accessibility score: > 90
- ✅ Bundle size: < 500KB (gzipped)
- ✅ First Contentful Paint: < 1.5s

### Adoption
- ✅ User satisfaction: 4.5/5 stars (survey)
- ✅ Weekly Active Users: 80% of target within 3 months

---

## Future Enhancements (When Backend APIs Are Ready)

### When Schema API Is Ready:
- Add Monaco autocomplete provider
- Build schema browser tree view
- Show column types and constraints
- Add inline schema documentation

### When Folder System Is Ready:
- Replace flat list with folder tree
- Add drag-and-drop organization
- Show folder breadcrumbs
- Add folder management UI

### When Pagination API Is Ready:
- Replace client-side pagination with server-side
- Handle 100,000+ row result sets
- Add server-side sorting
- Reduce initial page load time

### When WebSocket Is Ready:
- Replace polling with real-time updates
- Add live query status
- Show transaction progress in real-time
- Add collaboration features (live cursors)

---

## Summary

### Current Backend Capabilities ✅
- Authentication & RBAC
- Query execution (all operations)
- Approval workflow (transaction-based)
- Data source management
- User & group management
- Query history

### Frontend Features to Build (6-8 weeks)
1. Authentication UI
2. SQL Editor (Monaco)
3. Query Results (table + pagination)
4. Saved Queries (flat list + search)
5. Query History (searchable)
6. Approval Dashboard (with polling)
7. Data Source Management
8. User/Group Management

### Workarounds for Missing Features
- ❌ Schema API → Simple table list
- ❌ Autocomplete → Text search
- ❌ Folders → Search + filter
- ❌ Backend pagination → Client-side pagination
- ❌ WebSocket → Polling (3-5 seconds)

### Result
- ✅ Production-ready frontend in 6-8 weeks
- ✅ Works with current backend (no new APIs needed)
- ✅ Can be enhanced incrementally when new backend APIs are ready
- ✅ Focus on user experience and quality

---

**Last Updated:** January 28, 2025
**Status:** Ready to Start
**Next Step:** Initialize Next.js project and set up authentication
