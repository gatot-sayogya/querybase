# QueryBase Dashboard UI Plan
**Similar to Bytebase Community Edition**

**Status:** Planning Phase
**Target:** Modern database explorer UI with collaboration features
**Timeline:** 8-12 weeks

---

## Table of Contents

1. [Overview](#overview)
2. [Bytebase CE Analysis](#bytebase-ce-analysis)
3. [Technology Stack](#technology-stack)
4. [UI Architecture](#ui-architecture)
5. [Feature Implementation Plan](#feature-implementation-plan)
6. [Component Structure](#component-structure)
7. [Design System](#design-system)
8. [Development Roadmap](#development-roadmap)
9. [API Integration](#api-integration)
10. [Testing Strategy](#testing-strategy)

---

## Overview

### Vision

Build a modern, intuitive database explorer UI for QueryBase with:
- SQL query editor with autocomplete and syntax highlighting
- Real-time query results display
- Database schema browser
- Saved queries and folders
- Query history with filters
- Approval workflow UI for write operations
- Team collaboration features
- Responsive design (desktop & tablet)

### Goals

1. **User Experience**: Match Bytebase CE's clean, professional interface
2. **Performance**: Fast query execution and result rendering
3. **Collaboration**: Team-based workflow with approval UI
4. **Security**: Encrypted communication with backend
5. **Accessibility**: WCAG 2.1 AA compliant

### Success Metrics

- Time to first query: < 2 minutes from login
- Query execution latency: < 500ms perceived
- Page load time: < 2 seconds
- User satisfaction: 4.5/5 stars
- Adoption rate: 80% of target users within 3 months

---

## Bytebase CE Analysis

### Key Features to Replicate

**Core Features:**
1. ✅ **SQL Editor** - Monaco-based with autocomplete
2. ✅ **Query Results** - Table view with export
3. ✅ **Schema Browser** - Interactive database schema
4. ✅ **Saved Queries** - Organize in folders
5. ✅ **Query History** - Searchable history
6. ✅ **Tabbed Interface** - Multiple queries simultaneously
7. ✅ **Dark Mode** - Eye-friendly dark theme
8. ✅ **Keyboard Shortcuts** - Power user features
9. ✅ **Query Status** - Real-time status updates

**Unique QueryBase Features:**
1. ✅ **Approval Dashboard** - Review/Approve/Reject write operations
2. ✅ **EXPLAIN Visualization** - Query plan display
3. ✅ **Dry Run DELETE** - Preview affected rows
4. ✅ **Transaction Preview** - See changes before committing

### Bytebase UI Strengths

**Layout:**
- Left sidebar: Navigation (Databases, Queries, History)
- Main area: Tabbed query editor + results
- Right panel: Context-sensitive help (schema, history)

**Design Patterns:**
- Clean, minimal interface
- High information density
- Consistent color coding
- Intuitive icons and labels

---

## Technology Stack

### Frontend Framework

**Next.js 14/15** (App Router)
```bash
# Why Next.js?
- Server-side rendering (SEO, performance)
- API routes for BFF pattern
- Built-in optimization
- Great TypeScript support
- Easy deployment (Vercel, Docker)
```

### UI Component Library

**shadcn/ui** + **Radix UI**
```bash
# Why shadcn/ui?
- Modern, accessible components
- Built on Radix UI (primitives)
- Full TypeScript support
- Customizable via Tailwind
- Copy-paste components (no npm install bloat)
- Excellent dark mode support
```

### Editor

**Monaco Editor**
```typescript
// Monaco Editor (VS Code's editor)
- SQL syntax highlighting
- Autocomplete (can be customized)
- Multi-cursor support
- Keyboard shortcuts
- Split view
```

### Data Fetching

**TanStack Query** (React Query)
```typescript
// Why TanStack Query?
- Server state management
- Caching and revalidation
- Optimistic updates
- Background refetch
- Query invalidation
```

### State Management

**Zustand**
```typescript
// Why Zustand?
- Simple API
- No boilerplate
- TypeScript first
- Small bundle size
- DevTools integration
```

### Styling

**Tailwind CSS 4**
```css
/* Why Tailwind CSS? */
/* - Utility-first approach */
/* - Highly customizable */
/* - Great dark mode */
/* - Consistent design system */
/* - Small production bundle */
```

### Additional Libraries

```json
{
  "dependencies": {
    "monaco-editor": "^0.45.0",
    "@monaco-editor/react": "^4.6.0",
    "@tanstack/react-query": "^5.17.0",
    "zustand": "^4.4.0",
    "axios": "^1.6.0",
    "date-fns": "^3.0.0",
    "recharts": "^2.10.0",
    "react-table": "^8.11.0",
    "react-hot-toast": "^2.4.0",
    "zod": "^3.22.0"
  }
}
```

---

## UI Architecture

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
│   ├── databases/
│   │   ├── page.tsx              # Database list
│   │   ├── [id]/
│   │   │   └── page.tsx          # Database detail
│   │   │   └── schema/
│   │   │       └── page.tsx      # Schema browser
│   │   │   └── tables/
│   │   │       └── page.tsx      # Tables list
│   │   │
│   ├── queries/
│   │   ├── page.tsx              # Saved queries
│   │   ├── folders/
│   │   │   └── page.tsx          # Query folders
│   │   ├── [id]/
│   │   │   └── page.tsx          # Query detail/edit
│   │   │   └── edit/
│   │   │       └── page.tsx      # Edit query
│   │   │
│   ├── history/
│   │   └── page.tsx              # Query history
│   │   │
│   ├── approvals/
│   │   ├── page.tsx              # Approval dashboard
│   │   ├── [id]/
│   │   │   └── page.tsx          # Approval detail
│   │   │   └── review/
│   │   │       └── page.tsx      # Review approval
│   │   │
│   └── editor/
│       ├── page.tsx              # SQL editor
│       └── [id]/
│           └── page.tsx          # Query tab (saved state)
│
├── api/                          # API routes (BFF pattern)
│   ├── auth/
│   │   └── route.ts             # Auth proxy
│   ├── queries/
│   │   └── route.ts             # Query proxy
│   └── ...
│
└── layout.tsx                    # Root layout
```

### Component Hierarchy

```
components/
├── layout/
│   ├── Sidebar.tsx              # Left navigation
│   ├── Header.tsx               # Top bar
│   └── QueryTabs.tsx            # Tab manager
│
├── editor/
│   ├── SQLEditor.tsx           # Monaco editor wrapper
│   ├── EditorToolbar.tsx        # Format, run, save
│   ├── QueryStatusBar.tsx       # Status indicators
│   └── AutocompleteProvider.tsx # SQL autocomplete
│
├── results/
│   ├── ResultsTable.tsx         # Query results display
│   ├── ExportButton.tsx         # CSV, JSON export
│   ├── Pagination.tsx           # Results pagination
│   └── CellViewer.tsx           # Long content viewer
│
├── schema/
│   ├── SchemaTree.tsx           # Database schema tree
│   ├── TableViewer.tsx          # Table detail
│   ├── ColumnInfo.tsx           # Column information
│   └── IndexViewer.tsx          # Index viewer
│
├── queries/
│   ├── QueryCard.tsx            # Saved query card
│   ├── FolderTree.tsx           # Folder organization
│   ├── QueryHistoryList.tsx     # History list
│   └── QueryFilters.tsx         # Search/filter controls
│
├── approvals/
│   ├── ApprovalCard.tsx         # Approval request card
│   ├── ReviewDialog.tsx         # Review/approve/reject
│   ├── DryRunPreview.tsx       # DELETE preview
│   ├── TransactionStatus.tsx    # Transaction state
│   └── ApproverList.tsx         # Approver avatars
│
├── common/
│   ├── Button.tsx                # shadcn/ui button wrapper
│   ├── Input.tsx                 # shadcn/ui input wrapper
│   ├── Modal.tsx                 # Dialog modal
│   ├── Dropdown.tsx              # Dropdown menu
│   ├── Badge.tsx                 # Status badges
│   ├── Spinner.tsx               # Loading states
│   ├── EmptyState.tsx            # Empty state displays
│   ├── ErrorBoundary.tsx         # Error handling
│   └── Toast.tsx                 # Notifications
│
└── features/
    ├── ExplainPlan.tsx          # EXPLAIN visualization
    ├── ExplainCompare.tsx       # Before/after comparison
    └── QueryPerformance.tsx     # Performance metrics
```

---

## Feature Implementation Plan

### Phase 1: Foundation (Week 1-2)

**Goal:** Set up project structure and core infrastructure

**Tasks:**
1. Initialize Next.js 14 project with TypeScript
2. Install and configure shadcn/ui
3. Set up Tailwind CSS with custom theme
4. Configure ESLint, Prettier
5. Create base layout components
6. Set up routing structure
7. Implement authentication (JWT)
8. Create API client with axios
9. Set up Zustand stores
10. Configure Monaco Editor

**Deliverables:**
- Project scaffolding
- Authentication flow
- Base UI components
- API integration layer

**Success Criteria:**
- Can login and see dashboard
- Navigation works
- API calls succeed

---

### Phase 2: SQL Editor (Week 3-4)

**Goal:** Build feature-rich SQL editor

**Tasks:**
1. Monaco Editor integration
2. SQL syntax highlighting
3. SQL autocomplete (basic keywords, table names)
4. Editor toolbar (Run, Format, Save, Explain)
5. Query status indicators
6. Multi-tab support
7. Keyboard shortcuts
8. Error handling and display

**Components:**
- `SQLEditor.tsx` - Main editor component
- `EditorToolbar.tsx` - Run, Format, Save buttons
- `QueryStatusBar.tsx` - Status, execution time
- `QueryTabs.tsx` - Tab management
- `AutocompleteProvider.tsx` - Auto-complete logic

**Features:**
```typescript
// Editor Features
interface SQLEditorProps {
  dataSourceId: string;
  initialQuery?: string;
  readOnly?: boolean;
  onExecute?: (query: string) => void;
  onSave?: (query: string) => void;
}

// Autocomplete
const autoCompleteProvider = {
  provideCompletionItems: (model, position) => {
    // SQL keywords
    // Table names from schema
    // Column names from selected table
  }
};
```

**Success Criteria:**
- Can write and execute SQL queries
- Autocomplete suggests tables and columns
- Formatting works
- Tabs can be created/managed

---

### Phase 3: Query Results (Week 4-5)

**Goal:** Display query results effectively

**Tasks:**
1. Results table component (react-table)
2. Cell rendering (strings, numbers, JSON, dates)
3. Pagination (100, 500, 1000 rows per page)
4. Export functionality (CSV, JSON, Excel)
5. Column resizing and sorting
6. Row selection (for DELETE preview)
7. Loading states
8. Empty states
9. Error states
10. Long text truncation with viewer

**Components:**
- `ResultsTable.tsx` - Main results display
- `Pagination.tsx` - Pagination controls
- `ExportButton.tsx` - Export options
- `CellViewer.tsx` - Modal for long content
- `LoadingTable.tsx` - Skeleton loader

**Features:**
```typescript
interface ResultsProps {
  data: Array<Record<string, any>>;
  columns: ColumnInfo[];
  rowCount: number;
  executionTime: number;
  status: 'loading' | 'success' | 'error';
  onExport?: (format: 'csv' | 'json') => void;
}
```

**Success Criteria:**
- Results display correctly for all data types
- Pagination works smoothly
- Export downloads files correctly
- Performance handles 10,000+ rows

---

### Phase 4: Schema Browser (Week 5-6)

**Goal:** Interactive database schema visualization

**Tasks:**
1. Database connection test
2. Schema tree component
3. Table list with icons
4. Column details panel
5. Index viewer
6. Foreign key relationships
7. Refresh schema button
8. Search/filter tables
9. Collapse/expand tree
10. Schema statistics

**Components:**
- `SchemaTree.tsx` - Hierarchical tree view
- `TableViewer.tsx` - Table detail panel
- `ColumnInfo.tsx` - Column information
- `IndexViewer.tsx` - Index display

**Data Structure:**
```typescript
interface SchemaNode {
  type: 'database' | 'schema' | 'table' | 'column' | 'index';
  name: string;
  children?: SchemaNode[];
  metadata?: {
    type?: string;      // Column type
    nullable?: boolean;
    key?: boolean;       // Primary key
    indexed?: boolean;   // Has index
  };
}
```

**Success Criteria:**
- Can browse entire database schema
- Search finds tables quickly
- Performance handles 100+ tables

---

### Phase 5: Saved Queries (Week 6-7)

**Goal:** Query organization and management

**Tasks:**
1. Query list with search
2. Folder tree organization
3. Create/edit/delete queries
4. Move queries between folders
5. Query tags/labels
6. Favorite queries
7. Query sharing (team feature)
8. Bulk operations

**Components:**
- `QueryCard.tsx` - Query card in list view
- `FolderTree.tsx` - Folder sidebar
- `QueryEditor.tsx` - Create/edit form

**Data Model:**
```typescript
interface SavedQuery {
  id: string;
  name: string;
  description?: string;
  dataSourceId: string;
  queryText: string;
  folderId?: string;
  tags: string[];
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
}
```

**Success Criteria:**
- Can save and organize queries
- Search works by name/content
- Folders can be nested

---

### Phase 6: Query History (Week 7)

**Goal:** Searchable query execution history

**Tasks:**
1. History list with pagination
2. Advanced search (date, user, data source)
3. Filter by operation type
4. Status indicators (success, failed)
5. Re-run queries from history
6. History statistics
7. Clear history option
8. Export history

**Components:**
- `QueryHistoryList.tsx` - History list
- `QueryFilters.tsx` - Search/filter controls
- `HistoryStats.tsx` - Usage charts

**Features:**
```typescript
interface HistoryFilters {
  dateRange: [Date, Date];
  users: string[];
  dataSources: string[];
  operationTypes: OperationType[];
  status: 'success' | 'failed' | 'all';
}
```

**Success Criteria:**
- History loads quickly
- Search is fast
- Can re-run queries from history

---

### Phase 7: Approval Dashboard (Week 8-9)

**Goal:** UI for write operation approval workflow

**Tasks:**
1. Approval request list with filters
2. Approval detail view
3. Dry run DELETE preview
4. Review/approve/reject actions
5. Transaction status display
6. Approver assignments
7. Notification badges
8. Bulk approve/reject
9. Comment system
10. Approval history

**Components:**
- `ApprovalCard.tsx` - Request card
- `ReviewDialog.tsx` - Review modal
- `DryRunPreview.tsx` - Affected rows table
- `TransactionStatus.tsx` - Transaction state
- `ApproverList.tsx` - Approver avatars

**Workflow:**
```
Pending Approvals → Click → Detail View
                                ↓
                        [Dry Run Preview] → [Review Dialog]
                                ↓
                    Approve/Reject → Update Status
```

**Success Criteria:**
- Approvers can review requests easily
- Dry run preview works smoothly
- Transaction status updates in real-time

---

### Phase 8: EXPLAIN Visualization (Week 9)

**Goal:** Display query execution plans visually

**Tasks:**
1. Parse EXPLAIN output
2. Tree view visualization
3. Cost/row count highlighting
4. Index usage indicators
5. Before/after comparison
6. Performance tips
7. Explain anomalies
8. Export plan as image

**Components:**
- `ExplainPlan.tsx` - Plan tree visualization
- `ExplainCompare.tsx` - Before/after comparison
- `PerformanceTips.tsx` - Optimization suggestions

**Features:**
```typescript
interface ExplainNode {
  id: string;
  type: 'scan' | 'join' | 'sort' | 'aggregate';
  table?: string;
  index?: string;
  cost: number;
  rows: number;
  width: number;
  actualCost?: number;  // From EXPLAIN ANALYZE
  actualRows?: number;
  children: ExplainNode[];
}
```

**Success Criteria:**
- Plans render correctly
- Slow operations are highlighted
- Tips are actionable

---

### Phase 9: Data Source Management (Week 10)

**Goal:** UI for managing database connections

**Tasks:**
1. Data source list with status
2. Add/edit/delete data source (admin)
3. Connection test
4. Permission management UI
5. Health check status
6. Usage statistics
7. Last query time

**Components:**
- `DataSourceCard.tsx` - Connection card
- `DataSourceForm.tsx` - Add/edit form
- `PermissionMatrix.tsx` - Permission grid

**Success Criteria:**
- Can manage data sources visually
- Connection test feedback is clear
- Permissions are easy to configure

---

### Phase 10: Polish & Optimization (Week 11-12)

**Goal:** Production-ready polish

**Tasks:**
1. Performance optimization
2. Error boundary improvements
3. Loading states everywhere
4. Keyboard shortcuts
5. Tooltips and help text
6. Onboarding tour
7. Accessibility audit
8. Browser testing
9. Mobile responsive design
10. Analytics integration

**Deliverables:**
- Production-ready UI
- User documentation
- Performance benchmarks
- Accessibility report

---

## Component Structure

### Layout Components

```typescript
// Main layout structure
<DashboardLayout>
  <Sidebar>
    <DatabaseBrowser />
    <SavedQueries />
    <QueryHistory />
  </Sidebar>

  <MainContent>
    <QueryHeader />
    <QueryTabs />
    <TabContent>
      <SQLEditor />
      <QueryResults />
      <QueryStatus />
    </TabContent>
  </MainContent>

  <ContextPanel>
    {/* Context-sensitive */}
    <SchemaInfo />
    <QueryHistory />
    <Tips />
  </ContextPanel>
</DashboardLayout>
```

### Editor Components

```typescript
<SQLEditor>
  <MonacoEditor
    language="sql"
    theme="vs-dark"
    options={{
      minimap: { enabled: true },
      fontSize: 14,
      scrollBeyondLastLine: false,
      automaticLayout: true,
    }}
  />

  <EditorToolbar>
    <RunButton />
    <ExplainButton />
    <DryRunButton />
    <FormatButton />
    <SaveButton />
  </EditorToolbar>

  <StatusBar>
    <StatusBadge />
    <ExecutionTime />
    <RowCount />
  </StatusBar>
</SQLEditor>
```

---

## Design System

### Color Palette (Tailwind)

```css
/* Light Mode */
--primary: #3b82f6;      /* Blue 500 */
--success: #10b981;      /* Emerald 500 */
--warning: #f59e0b;      /* Amber 500 */
--danger: #ef4444;       /* Red 500 */
--info: #6366f1;         /* Indigo 500 */

/* Dark Mode */
--primary: #60a5fa;      /* Blue 400 */
--success: #34d399;      /* Emerald 400 */
--warning: #fbbf24;      /* Amber 400 */
--danger: #f87171;       /* Red 400 */
--info: #818cf8;         /* Indigo 400 */

/* Neutral Grays */
--gray-50: #f9fafb;
--gray-100: #f3f4f6;
--gray-200: #e5e7eb;
--gray-300: #d1d5db;
--gray-400: #9ca3af;
--gray-500: #6b7280;
--gray-600: #4b5563;
--gray-700: #374151;
--gray-800: #1f2937;
--gray-900: #111827;
```

### Typography

```css
/* Font Families */
--font-sans: Inter, system-ui, sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', monospace;

/* Font Sizes */
--text-xs: 0.75rem;   /* 12px */
--text-sm: 0.875rem;  /* 14px */
--text-base: 1rem;   /* 16px */
--text-lg: 1.125rem;  /* 18px */
--text-xl: 1.25rem;   /* 20px */
--text-2xl: 1.5rem;    /* 24px */
--text-3xl: 1.875rem;  /* 30px */
```

### Spacing Scale

```css
/* Spacing (Tailwind) */
--spacing-1: 0.25rem;  /* 4px */
--spacing-2: 0.5rem;   /* 8px */
--spacing-3: 0.75rem;  /* 12px */
--spacing-4: 1rem;     /* 16px */
--spacing-5: 1.25rem;  /* 20px */
--spacing-6: 1.5rem;   /* 24px */
--spacing-8: 2rem;     /* 32px */
--spacing-10: 2.5rem;  /* 40px */
--spacing-12: 3rem;    /* 48px */
```

---

## Development Roadmap

### Week 1-2: Foundation
- [ ] Project setup
- [ ] Authentication
- [ ] Layout components
- [ ] API integration
- [ ] Basic routing

### Week 3-4: SQL Editor
- [ ] Monaco Editor
- [ ] Syntax highlighting
- [ ] Autocomplete (v1)
- [ ] Editor toolbar
- [ ] Tab management

### Week 4-5: Query Results
- [ ] Results table
- [ ] Pagination
- [ ] Export functionality
- [ ] Loading states
- [ ] Error handling

### Week 5-6: Schema Browser
- [ ] Schema tree
- [ ] Table viewer
- [ ] Column details
- [ ] Index viewer
- [ ] Search functionality

### Week 6-7: Query Management
- [ ] Saved queries
- [ ] Folders
- [ ] Query history
- [ ] Search/filter

### Week 8-9: Approvals
- [ ] Approval dashboard
- [ ] Review workflow
- [ ] Dry run preview
- [ ] Transaction status

### Week 9: Advanced Features
- [ ] EXPLAIN visualization
- [ ] Performance metrics
- [ ] Query comparison

### Week 10: Data Sources
- [ ] Data source management
- [ ] Connection test UI
- [ ] Permissions UI

### Week 11-12: Polish
- [ ] Performance optimization
- [ ] Error boundaries
- [ ] Keyboard shortcuts
- [ ] Accessibility
- [ ] Documentation

---

## API Integration

### API Client Setup

```typescript
// lib/api-client.ts
import axios from 'axios';
import { QueryBaseAPI } from '@/lib/api';

const apiClient = new QueryBaseAPI({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 30000,
});

// API Routes (Next.js App Router)
app/api/proxy/[...path]/route.ts
```

### Encrypted Communication

```typescript
// Use encryption when ready (see ENCRYPTED_COMMUNICATION.md)
import { encryptPayload, decryptResponse } from '@/lib/encryption';

const encryptedAPI = {
  async executeQuery(data: any) {
    const encrypted = await encryptPayload(data, encryptionKey);
    return apiClient.post('/queries', encrypted);
  }
};
```

### Real-time Updates

```typescript
// WebSocket connection for real-time updates
import { useWebSocket } from '@/hooks/useWebSocket';

function QueryStatusUpdater({ queryId }: { queryId: string }) {
  const { lastMessage, sendMessage } = useWebSocket(
    `ws://localhost:8080/ws/queries/${queryId}`
  );

  useEffect(() => {
    if (lastMessage?.status === 'completed') {
      // Refresh results
      queryClient.invalidateQueries(['query', queryId]);
    }
  }, [lastMessage]);
}
```

---

## Testing Strategy

### Unit Tests

```typescript
// Component tests
describe('SQLEditor', () => {
  it('should format SQL query', () => {
    const { result } = render(<SQLEditor query="SELECT  *  from  users" />);
    expect(formatButton()).toBeEnabled();
  });
});

// Store tests
describe('queryStore', () => {
  it('should execute query', async () => {
    const { result } = await store.dispatch(executeQuery(query));
    expect(result).toMatchObject({ status: 'completed' });
  });
});
```

### Integration Tests

```typescript
// E2E tests with Playwright
test('execute query flow', async ({ page }) => {
  await page.goto('/editor');
  await page.fill('[data-testid="sql-editor"]', 'SELECT * FROM users');
  await page.click('[data-testid="run-button"]');
  await expect(page.locator('[data-testid="results"]')).toBeVisible();
});
```

### Performance Tests

```typescript
// Lighthouse CI
// - Performance score > 90
// - Accessibility score > 90
// - Best practices > 90

// Bundle size
// - Main bundle: < 200KB gzipped
// - Vendor chunks: code-split
```

---

## Deployment

### Build Configuration

```javascript
// next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',  // For Docker
  reactStrictMode: true,
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  },
  // Code splitting
  splitChunks: (chunks) => [
    /node_modules\/(monaco-editor)/.*/.map((chunk) => ({
      name: 'monaco',
      priority: 40,
      test: /[\\/]monaco-editor[\\/]/,
    })),
  ],
};
```

### Docker Deployment

```dockerfile
# Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
```

### Vercel Deployment

```bash
# Deploy to Vercel
vercel --prod

# Environment variables
NEXT_PUBLIC_API_URL=https://api.querybase.com
ENCRYPTION_KEY=your-encryption-key
```

---

## Success Metrics

### Technical Metrics

| Metric | Target | Measurement |
|--------|--------|------------|
| Page Load | < 2s | Lighthouse |
| Time to Interactive | < 3s | Lighthouse |
| First Contentful Paint | < 1.5s | Lighthouse |
| Bundle Size | < 500KB | Webpack Bundle Analyzer |
| API Response | < 200ms | DevTools Network |

### User Metrics

| Metric | Target | Measurement |
|--------|--------|------------|
| Login to Query | < 30s | Analytics |
| Query Execution | < 500ms perceived | User feedback |
| Error Rate | < 1% | Error tracking |
| User Satisfaction | > 4.5/5 | Survey |
| Weekly Active Users | 80% target | Analytics |

---

## Related Documentation

- **[Bytebase CE](https://github.com/bytebase/bytebase)** - Reference implementation
- **[Getting Started](../getting-started/)** - Setup guide
- **[API Documentation](../../CLAUDE.md)** - Backend API reference
- **[Encrypted Communication](ENCRYPTED_COMMUNICATION.md)** - Security implementation

---

**Status:** Planning Phase - Ready to Start Development
**Estimated Timeline:** 12 weeks
**Team Size:** 1-2 frontend developers
**Dependencies:** Backend API (✅ Complete), Encryption feature (🔄 Planned)

---

**Next Steps:**

1. Review and approve plan
2. Set up Next.js project
3. Begin Phase 1 (Foundation)
4. Create initial designs in Figma
5. Start with SQL editor (MVP)
