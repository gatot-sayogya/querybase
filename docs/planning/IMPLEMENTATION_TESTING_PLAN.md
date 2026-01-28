# Implementation & Testing Plan
**Backend Polish + Frontend Development**

**Date:** January 28, 2025
**Status:** Ready to Start
**Duration:** 11 weeks total (3 weeks backend + 8 weeks frontend)

---

## Overview

This plan provides a structured approach to:
1. **Implement backend improvements** (3 weeks)
2. **Build dashboard UI** (6-8 weeks)
3. **Test thoroughly** (throughout)
4. **Deploy to production**

**Parallel Development:**
- Backend team: Polish current workflow (Week 1-3)
- Frontend team: Build UI (Week 1-8)
- QA team: Continuous testing

---

## Part 1: Backend Implementation Plan (Week 1-3)

### Week 1: Query Enhancements

#### Day 1-2: Query Results Pagination

**Goal:** Add pagination API for large result sets

**Implementation Tasks:**

**1. Update Query Service** (`internal/service/query.go`)

Add pagination method:
```go
func (s *QueryService) GetPaginatedResults(
    queryID uuid.UUID,
    page, limit int,
    sortColumn, sortDirection string,
) (*PaginatedResultDTO, error)
```

**Steps:**
- Fetch QueryResult from DB
- Parse JSONB rows to memory
- Sort if sortColumn provided
- Paginate with offset/limit
- Return subset with metadata

**Files:**
- `internal/service/query.go` - Add GetPaginatedResults method
- `internal/service/query_test.go` - Add tests

**2. Update Query Handler** (`internal/api/handlers/query.go`)

Add pagination endpoint:
```go
func (h *Handler) GetQueryResults(c *gin.Context)
```

**Steps:**
- Parse query params (page, limit, sort_column, sort_direction)
- Validate page ≥ 1, limit ≤ 1000
- Call service.GetPaginatedResults()
- Return paginated response

**Files:**
- `internal/api/handlers/query.go` - Add GetQueryResults handler
- `internal/api/routes/routes.go` - Add route

**3. Update DTOs** (`internal/api/dto/query.go`)

Add pagination DTO:
```go
type PaginatedResultDTO struct {
    QueryID     uuid.UUID        `json:"queryId"`
    Columns     []string         `json:"columns"`
    ColumnTypes []string         `json:"columnTypes"`
    Rows        []map[string]any   `json:"rows"`
    Pagination  PaginationMeta   `json:"pagination"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    TotalRows  int `json:"totalRows"`
    TotalPages int `json:"totalPages"`
}
```

**4. Update Routes** (`internal/api/routes/routes.go`)

Add route:
```go
queryRoutes.GET("/:id/results", handlers.GetQueryResults)
```

**5. Testing**

**Unit Tests:**
```go
func TestGetPaginatedResults(t *testing.T) {
    // Test pagination with 1000 rows, page=1, limit=100
    // Test page 2, page 3
    // Test sorting
    // Test invalid page (returns 404)
    // Test invalid limit (defaults to 100)
}
```

**Integration Tests:**
```go
func TestQueryPaginationIntegration(t *testing.T) {
    // Execute query
    // Fetch results with pagination
    // Verify correct subset returned
    // Verify pagination metadata
}
```

**Acceptance Criteria:**
- ✅ Can paginate 10,000 row result set into 100-row pages
- ✅ Sorting works on all column types
- ✅ Returns 404 for non-existent query
- ✅ Defaults to page=1, limit=100
- ✅ All tests passing

---

#### Day 3-4: Query Export API

**Goal:** Allow users to download query results

**Implementation Tasks:**

**1. Create Export Service** (`internal/service/export.go`)

```go
type ExportService struct {
    db *gorm.DB
}

func (s *ExportService) ExportQueryResults(
    queryID uuid.UUID,
    format string, // csv, json
) ([]byte, string, string, error)
```

**Steps:**
- Fetch QueryResult from DB
- Parse rows from JSONB
- Format to CSV or JSON
- Return bytes, content-type, filename

**Files:**
- `internal/service/export.go` - New file

**2. Create Export Handler** (`internal/api/handlers/export.go`)

```go
func (h *Handler) ExportQueryResults(c *gin.Context)
```

**Steps:**
- Parse query param (format)
- Call export service
- Set headers (Content-Type, Content-Disposition)
- Return file response

**Files:**
- `internal/api/handlers/export.go` - New file

**3. Create Export DTOs** (`internal/api/dto/export.go`)

```go
type ExportRequestDTO struct {
    QueryID uuid.UUID `form:"id" binding:"required,uuid"`
    Format string     `form:"format" binding:"required,oneof=csv json"`
}
```

**Files:**
- `internal/api/dto/export.go` - New file

**4. Update Routes**

Add route:
```go
queryRoutes.GET("/:id/export", handlers.ExportQueryResults)
```

**5. Testing**

**Unit Tests:**
```go
func TestExportToCSV(t *testing.T) {
    // Test CSV export
    // Verify headers
    // Verify data rows
    // Verify special characters escaped
}

func TestExportToJSON(t *testing.T) {
    // Test JSON export
    // Verify valid JSON
    // Verify data types preserved
}
```

**Integration Tests:**
- Execute real query
- Export to CSV
- Export to JSON
- Verify files download correctly

**Acceptance Criteria:**
- ✅ CSV export works with special characters
- ✅ JSON export preserves data types
- ✅ Filename includes timestamp
- ✅ Returns 404 for non-existent query
- ✅ All tests passing

---

### Week 2: Collaboration & Monitoring

#### Day 1-3: Approval Comments System

**Goal:** Allow discussion on approval requests

**Implementation Tasks:**

**1. Database Migration**

`migrations/000006_add_approval_comments.up.sql`:
```sql
CREATE TABLE approval_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_comments_approval ON approval_comments(approval_id);
CREATE INDEX idx_approval_comments_created ON approval_comments(created_at);
```

`migrations/000006_add_approval_comments.down.sql`:
```sql
DROP TABLE IF EXISTS approval_comments;
```

**2. Update Approval Model** (`internal/models/approval.go`)

Add:
```go
type ApprovalComment struct {
    ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
    ApprovalID uuid.UUID  `gorm:"type:uuid;not null" json:"approvalId"`
    UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
    User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Comment    string     `gorm:"type:text;not null" json:"comment"`
    CreatedAt  time.Time  `json:"createdAt"`
    UpdatedAt  time.Time  `json:"updatedAt"`
}

// BeforeCreate hook
func (ac *ApprovalComment) BeforeCreate(tx *gorm.DB) error {
    if ac.ID == uuid.Nil {
        ac.ID = uuid.New()
    }
    return nil
}
```

**3. Create Comment Service** (`internal/service/comment.go`)

```go
type CommentService struct {
    db *gorm.DB
}

func (s *CommentService) AddComment(approvalID, userID uuid.UUID, comment string) (*ApprovalComment, error)
func (s *CommentService) GetComments(approvalID uuid.UUID) ([]ApprovalComment, error)
func (s *CommentService) UpdateComment(commentID, userID uuid.UUID, comment string) (*ApprovalComment, error)
func (s *CommentService) DeleteComment(commentID uuid.UUID) error
```

**Files:**
- `internal/service/comment.go` - New file

**4. Update Approval Handler** (`internal/api/handlers/approval.go`)

Add endpoints:
```go
func (h *Handler) AddComment(c *gin.Context)
func (h *Handler) GetComments(c *gin.Context)
func (h *Handler) UpdateComment(c *gin.Context)
func (h *Handler) DeleteComment(c *gin.Context)
```

**5. Update Approval DTOs** (`internal/api/dto/approval.go`)

Add:
```go
type AddCommentRequestDTO struct {
    Comment string `json:"comment" binding:"required,min=1,max=1000"`
}

type UpdateCommentRequestDTO struct {
    Comment string `json:"comment" binding:"required,min=1,max=1000"`
}

type CommentResponseDTO struct {
    ID         uuid.UUID `json:"id"`
    ApprovalID uuid.UUID `json:"approvalId"`
    Comment    string    `json:"comment"`
    User       UserDTO   `json:"user"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
}
```

**6. Update Routes**

Add routes:
```go
approvalRoutes.POST("/:id/comments", handlers.AddComment)
approvalRoutes.GET("/:id/comments", handlers.GetComments)
approvalRoutes.PUT("/:id/comments/:commentId", handlers.UpdateComment)
approvalRoutes.DELETE("/:id/comments/:commentId", handlers.DeleteComment)
```

**7. Testing**

**Unit Tests:**
```go
func TestAddComment(t *testing.T) {
    // Test adding comment
    // Test comment creation
    // Test notification sent
}

func TestGetComments(t *testing.T) {
    // Test retrieving comments
    // Test pagination
    // Test ordering
}

func TestUpdateComment(t *testing.T) {
    // Test updating own comment
    // Test updating other's comment (admin)
    // Test permission denied
}

func TestDeleteComment(t *testing.T) {
    // Test deleting own comment
    // Test deleting other's comment (admin)
    // Test permission denied
    // Test cascade delete on approval deletion
}
```

**Integration Tests:**
```go
func TestApprovalCommentsWorkflow(t *testing.T) {
    // Create approval request
    // Add comment
    // Retrieve comments
    // Approve request
    // Verify comments still accessible
}
```

**Acceptance Criteria:**
- ✅ Comments can be added to approvals
- ✅ Comments can be retrieved
- ✅ Comments can be updated by owner
- ✅ Comments can be deleted by owner/admin
- ✅ Comments cascade delete when approval deleted
- ✅ All tests passing

---

#### Day 4-5: Data Source Health Check

**Goal:** Monitor data source health

**Implementation Tasks:**

**1. Update Data Source Service** (`internal/service/datasource.go`)

Add:
```go
func (s *DataSourceService) GetHealthStatus(dataSourceID string) (*HealthStatusDTO, error)
```

**Steps:**
- Get data source
- Test connection
- Measure latency
- Get database version (SELECT version())
- Get connection count

**Files:**
- `internal/service/datasource.go` - Update

**2. Update Data Source Handler** (`internal/api/handlers/datasource.go`)

Add:
```go
func (h *Handler) GetHealthStatus(c *gin.Context)
```

**3. Update Data Source DTOs** (`internal/api/dto/datasource.go`)

Add:
```go
type HealthStatusDTO struct {
    Status      string    `json:"status"` // healthy, unhealthy, unknown
    Latency     string    `json:"latency"`
    LastCheck   time.Time `json:"lastCheck"`
    Version     string    `json:"version,omitempty"`
    Connections int       `json:"connections"`
    Error       string    `json:"error,omitempty"`
}
```

**4. Update Routes**

Add route:
```go
dataSourceRoutes.GET("/:id/health", handlers.GetHealthStatus)
```

**5. Testing**

**Unit Tests:**
```go
func TestHealthCheck(t *testing.T) {
    // Test healthy data source
    // Test unreachable data source
    // Test timeout handling
}
```

**Integration Tests:**
- Test with real PostgreSQL
- Test with real MySQL
- Test connection pooling stats

**Acceptance Criteria:**
- ✅ Returns "healthy" for working connections
- ✅ Returns "unhealthy" for failed connections
- ✅ Measures latency accurately
- ✅ Returns database version
- ✅ All tests passing

---

### Week 3: Polish & Production Ready

#### Day 1-2: Error Handling Improvements

**Goal:** Consistent error responses

**Implementation Tasks:**

**1. Create Error Types** (`internal/api/errors/errors.go`)

```go
package errors

type ErrorCode string

const (
    ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden         ErrorCode = "FORBIDDEN"
    ErrCodeNotFound          ErrorCode = "NOT_FOUND"
    ErrCodeValidationError ErrorCode = "VALIDATION_ERROR"
    ErrCodeInternalError     ErrorCode = "INTERNAL_ERROR"
    ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
)

type AppError struct {
    Code       ErrorCode   `json:"code"`
    Message    string     `json:"message"`
    Details    string     `json:"details,omitempty"`
    RequestID  string     `json:"requestId,omitempty"`
    StatusCode int        `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}
```

**2. Create Error Middleware** (`internal/api/middleware/errors.go`)

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last()

            // Log error
            logger.Error("Request error",
                "path", c.Request.URL.Path,
                "method", c.Request.Method,
                "error", err.Error(),
                "requestId", c.GetString("requestId"),
            )

            // Return consistent error response
            var statusCode int
            var appErr *errors.AppError

            switch e := err.Err.(type) {
            case *errors.AppError:
                appErr = e
                statusCode = appErr.StatusCode
            case validator.ValidationErrors:
                statusCode = 400
                appErr = &errors.AppError{
                    Code: errors.ErrCodeValidationError,
                    Message: "Validation failed",
                    Details: e.Error(),
                    StatusCode: 400,
                }
            default:
                statusCode = 500
                appErr = &errors.AppError{
                    Code: errors.ErrCodeInternalError,
                    Message: "Internal server error",
                    StatusCode: 500,
                }
            }

            c.JSON(statusCode, appErr)
            c.Abort()
        }
    }
}
```

**3. Update Main.go** (`cmd/api/main.go`)

Add error middleware:
```go
r.Use(middleware.ErrorHandler())
```

**4. Testing:**
```go
func TestErrorHandler(t *testing.T) {
    // Test 401 unauthorized
    // Test 403 forbidden
    // Test 404 not found
    // Test 500 internal error
    // Test request ID included
}
```

**Acceptance Criteria:**
- ✅ All errors return consistent JSON format
- ✅ Request IDs included for tracing
- ✅ Errors logged properly
- ✅ Appropriate HTTP status codes

---

#### Day 3-4: Request Validation Improvements

**Goal:** Better input validation

**Implementation Tasks:**

**1. Add Validation**

Update DTOs with validation tags:
```go
type ExecuteQueryRequestDTO struct {
    DataSourceID uuid.UUID `json:"dataSourceId" binding:"required,uuid"`
    QueryText    string    `json:"queryText" binding:"required,min=1,max=10000"`
    Save         bool      `json:"save"`
}
```

**2. Add Custom Validators**

Create `internal/api/validators/validators.go`:
```go
func validateSQL(fl validator.FieldLevel) validator.FieldLevel {
    field := fl.Field()
    param := fl.Param()

    // Custom validation logic
    // Can add SQL syntax validation here
    return nil
}
```

**3. Testing:**
- Test all request DTOs
- Test validation error messages
- Test sanitization

**Acceptance Criteria:**
- ✅ All input validated
- ✅ Clear error messages
- ✅ SQL injection prevention

---

#### Day 5: Performance Benchmarks

**Goal:** Measure performance

**Implementation Tasks:**

**1. Create Benchmark Tests** (`internal/service/query_benchmark_test.go`)

```go
func BenchmarkQueryExecution(b *testing.B) {
    service := setupService()
    query := "SELECT * FROM users LIMIT 1000"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.ExecuteQuery(dataSourceID, query, userID)
    }
}

func BenchmarkGetPaginatedResults(b *testing.B) {
    service := setupService()
    queryID := createTestQuery(10000) // 10,000 rows

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.GetPaginatedResults(queryID, 1, 100, "id", "asc")
    }
}
```

**2. Run Benchmarks**

```bash
go test -bench=. -benchmem
```

**3. Performance Targets:**

| Metric | Target | Measurement |
|--------|--------|------------|
| Query execution | < 500ms | SELECT 1000 rows |
| Pagination | < 100ms | 10,000 row result |
| Export CSV | < 1s | 10,000 rows |
| Export JSON | < 500ms | 10,000 rows |

**4. Create Performance Baseline Document** (`docs/performance/PERFORMANCE_BASELINE.md`)

Document:
- Current benchmark results
- Target metrics
- Optimization opportunities

**Acceptance Criteria:**
- ✅ All benchmarks run successfully
- ✅ Performance within targets
- ✅ Documentation created

---

## Part 2: Frontend Implementation Plan (Week 1-8)

### Week 1-2: Foundation

#### Day 1-2: Project Setup

**Tasks:**

**1. Initialize Next.js Project**

```bash
cd web
npx create-next-app@latest querybase --typescript --tailwind --app
```

**2. Install Dependencies**

```bash
cd querybase
npm install @monaco-editor/react @tanstack/react-query zustand axios date-fns
npm install -D @types/node
```

**3. Install shadcn/ui**

```bash
npx shadcn-ui@latest init
npx shadcn-ui@latest add button
npx shadcn-ui@latest add input
npx shadcn-ui@latest add table
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add dropdown-menu
npx shadcn-ui@latest add select
npx shadcn-ui@latest add tabs
npx shadcn-ui@latest add badge
npx shadcn-ui@latest add toast
```

**4. Setup Project Structure**

```
web/
├── app/
├── components/
├── lib/
└── public/
```

**Acceptance Criteria:**
- ✅ Next.js project created
- ✅ Dependencies installed
- ✅ shadcn/ui configured
- ✅ TypeScript configured

---

#### Day 3-4: Authentication

**Tasks:**

**1. Create API Client** (`lib/api-client.ts`)

```typescript
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 30000,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

**2. Create Login Page** (`app/(auth)/login/page.tsx`)

```typescript
'use client';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const router = useRouter();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      const response = await apiClient.post('/api/v1/auth/login', {
        email,
        password,
      });

      const { token } = response.data;
      localStorage.setItem('token', token);
      router.push('/editor');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed');
    }
  };

  return (
    // Login form UI
  );
}
```

**3. Create Auth Hook** (`lib/hooks/useAuth.ts`)

```typescript
export function useAuth() {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const router = useRouter();

  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    setToken(storedToken);
    if (storedToken) {
      fetchUser(storedToken);
    }
  }, []);

  const fetchUser = async (authToken: string) => {
    try {
      const response = await apiClient.get('/api/v1/auth/me');
      setUser(response.data);
    } catch (err) {
      // Token might be invalid
      localStorage.removeItem('token');
      setToken(null);
    }
  };

  const login = async (email: string, password: string) => {
    const response = await apiClient.post('/api/v1/auth/login', {
      email,
      password,
    });
    const { token } = response.data;
    localStorage.setItem('token', token);
    setToken(token);
    await fetchUser(token);
    router.push('/editor');
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setUser(null);
    router.push('/login');
  };

  return { token, user, login, logout, isAuthenticated: !!token };
}
```

**Acceptance Criteria:**
- ✅ Login page works
- ✅ JWT token stored and used
- ✅ Protected routes redirect to login
- ✅ Logout clears token
- ✅ 401 redirects to login

---

#### Day 5: Layout Components

**Tasks:**

**1. Create Layout** (`app/(dashboard)/layout.tsx`)

```typescript
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { token } = useAuth();

  if (!token) {
    redirect('/login');
  }

  return (
    <div className="flex h-screen">
      <Sidebar />
      <div className="flex-1 flex flex-col">
        <Header />
        <main className="flex-1 overflow-auto p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
```

**2. Create Sidebar** (`components/layout/Sidebar.tsx`)

```typescript
export function Sidebar() {
  const pathname = usePathname();

  const menuItems = [
    { href: '/editor', label: 'SQL Editor', icon: DatabaseIcon },
    { href: '/queries', label: 'Saved Queries', icon: SaveIcon },
    { href: '/history', label: 'Query History', icon: ClockIcon },
    { href: '/approvals', label: 'Approvals', icon: CheckCircleIcon },
    { href: '/datasources', label: 'Data Sources', icon: ServerIcon, admin: true },
  ];

  return (
    <aside className="w-64 border-r bg-gray-50 p-4">
      <nav className="space-y-2">
        {menuItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-100",
              pathname === item.href && "bg-gray-200"
            )}
          >
            {item.icon}
            <span>{item.label}</span>
          </Link>
        ))}
      </nav>
    </aside>
  );
}
```

**Acceptance Criteria:**
- ✅ Layout responsive
- ✅ Sidebar navigation works
- ✅ Active page highlighted
- ✅ Admin-only routes hidden for non-admins

---

### Week 3-4: SQL Editor & Query Results

#### Day 1-3: SQL Editor

**Tasks:**

**1. Create SQL Editor Component** (`components/editor/SQLEditor.tsx`)

```typescript
'use client';

import Editor from '@monaco-editor/react';

interface SQLEditorProps {
  value: string;
  onChange: (value: string) => void;
  onExecute: () => void;
  onSave: () => void;
  readOnly?: boolean;
}

export function SQLEditor({
  value,
  onChange,
  onExecute,
  onSave,
  readOnly = false,
}: SQLEditorProps) {
  return (
    <div className="border rounded-lg overflow-hidden">
      <Editor
        height="500px"
        defaultLanguage="sql"
        value={value}
        onChange={(value) => onChange(value || '')}
        theme="vs-dark"
        options={{
          minimap: { enabled: false },
          fontSize: 14,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          readOnly,
        }}
      />
      <EditorToolbar
        onExecute={onExecute}
        onSave={onSave}
        canSave={!readOnly && value.trim().length > 0}
      />
    </div>
  );
}
```

**2. Create Editor Toolbar** (`components/editor/EditorToolbar.tsx`)

```typescript
interface EditorToolbarProps {
  onExecute: () => void;
  onSave: () => void;
  canSave: boolean;
}

export function EditorToolbar({ onExecute, onSave, canSave }: EditorToolbarProps) {
  return (
    <div className="flex gap-2 p-2 bg-gray-100 border-t">
      <Button onClick={onExecute} disabled={!canSave}>
        <PlayIcon className="w-4 h-4" />
        Run (F5)
      </Button>
      <Button onClick={onSave} disabled={!canSave} variant="secondary">
        <SaveIcon className="w-4 h-4" />
        Save
      </Button>
    </div>
  );
}
```

**Acceptance Criteria:**
- ✅ Monaco editor renders correctly
- ✅ SQL syntax highlighting works
- ✅ Run button executes query
- ✅ Save button works
- ✅ Keyboard shortcuts (F5 to run)

---

#### Day 4-5: Query Results Display

**Tasks:**

**1. Create Results Table** (`components/results/ResultsTable.tsx`)

```typescript
interface ResultsTableProps {
  result: QueryResult;
  pageSize?: number;
}

export function ResultsTable({ result, pageSize = 100 }: ResultsTableProps) {
  const [page, setPage] = useState(1);

  // Client-side pagination
  const totalPages = Math.ceil(result.rows.length / pageSize);
  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedRows = result.rows.slice(startIndex, endIndex);

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

      <div className="text-sm text-gray-500 mt-2">
        Showing {startIndex + 1}-{Math.min(endIndex, result.rows.length)} of {result.rows.length} rows
      </div>
    </div>
  );
}
```

**2. Create Pagination Component** (`components/results/Pagination.tsx`)

**Acceptance Criteria:**
- ✅ Results table renders correctly
- ✅ Client-side pagination works
- ✅ Different data types formatted correctly
- ✅ Large result sets (10,000+ rows) handled

---

### Week 5: Approval Dashboard

#### Day 1-3: Approval UI

**Tasks:**

**1. Create Approval List** (`app/(dashboard)/approvals/page.tsx`)

```typescript
'use client';

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
          <TabsTrigger value="pending">Pending ({approvals.filter(a => a.status === 'pending').length})</TabsTrigger>
          <TabsTrigger value="all">All</TabsTrigger>
        </TabsList>
      </Tabs>

      <div className="grid gap-4 mt-6">
        {approvals.map((approval) => (
          <ApprovalCard key={approval.id} approval={approval} />
        ))}
      </div>
    </div>
  );
}
```

**2. Create Approval Card** (`components/approvals/ApprovalCard.tsx`)

**3. Create Review Dialog** (`components/approvals/ReviewDialog.tsx`)

**4. Implement Polling Hook** (`lib/hooks/usePolling.ts`)

```typescript
export function usePolling(url: string, interval = 5000) {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(url);
        const result = await response.json();
        setData(result);
        setError(null);
      } catch (err) {
        setError(err);
      }
    };

    fetchData();
    const intervalId = setInterval(fetchData, interval);
    return () => clearInterval(intervalId);
  }, [url, interval]);

  return { data, error, refetch: fetchData };
}
```

**Acceptance Criteria:**
- ✅ Approval list displays correctly
- ✅ Polling updates every 5 seconds
- ✅ Can approve/reject approvals
- ✅ Transaction preview works
- ✅ Real-time status updates (via polling)

---

### Week 6: Admin Features

#### Day 1-2: Data Source Management

**Tasks:**

**1. Data Source List Page** (`app/(dashboard)/datasources/page.tsx`)

**2. Data Source Card** (`components/datasources/DataSourceCard.tsx`)

**3. Data Source Form** (`components/datasources/DataSourceForm.tsx`)

**Acceptance Criteria:**
- ✅ Admins can list data sources
- ✅ Admins can add/edit/delete data sources
- ✅ Connection test works
- ✅ Permission management UI

---

#### Day 3-5: User & Group Management

**Tasks:**

**1. User Management** (`app/(dashboard)/users/page.tsx`)

**2. Group Management** (`app/(dashboard)/groups/page.tsx`)

**Acceptance Criteria:**
- ✅ Admins can manage users
- ✅ Admins can manage groups
- ✅ User-group assignment works

---

### Week 7-8: Polish & Optimization

#### Day 1-2: Error Handling & Loading States

**Tasks:**

**1. Global Error Boundary** (`app/error.tsx`)

**2. Loading Spinners** (throughout app)

**3. Empty States** (throughout app)

**Acceptance Criteria:**
- ✅ Errors displayed gracefully
- ✅ Loading states on all async operations
- ✅ Empty states for empty data
- ✅ No console errors

---

#### Day 3-4: Performance Optimization

**Tasks:**

**1. Code Splitting** (Next.js automatic)

**2. Lazy Loading** (`next/dynamic`)

**3. Image Optimization** (Next.js Image component)

**4. Bundle Size Analysis**

```bash
npm run build
npm run analyze
```

**Acceptance Criteria:**
- ✅ Page load time < 2s
- ✅ Time to interactive < 3s
- ✅ Bundle size < 500KB (gzipped)
- ✅ Lighthouse score > 90

---

#### Day 5: Testing

**Tasks:**

**1. Component Tests** (React Testing Library)

```bash
npm run test
```

**2. E2E Tests** (Playwright)

```bash
npm run test:e2e
```

**3. Accessibility Audit** (Lighthouse)

```bash
npm run lighthouse
```

**Acceptance Criteria:**
- ✅ Component tests passing
- ✅ E2E tests covering critical paths
- ✅ Accessibility score > 90
- ✅ Performance score > 90

---

## Part 3: Testing Strategy

### Backend Testing

#### Unit Tests

**Target:** 80% code coverage

**Tools:**
- Go built-in testing
- `testify` (test runner)

**Commands:**
```bash
make test                # Run all tests
make test-short          # Skip integration tests
make test-coverage       # With coverage report
make test-bench          # Performance benchmarks
```

**Test Structure:**
```
internal/
├── service/
│   ├── query_test.go           # Query service tests
│   ├── parser_test.go          # Parser tests
│   ├── approval_test.go       # Approval tests
│   └── export_test.go          # Export tests (NEW)
├── models/
│   ├── user_test.go
│   ├── folder_test.go          # Folder tests (NEW)
│   └── approval_test.go       # Approval model tests (UPDATED)
├── api/
│   └── handlers/
│       └── export_test.go      # Handler tests (NEW)
└── validators/
    └── validators_test.go     # Validation tests (NEW)
```

#### Integration Tests

**Target:** Cover all user workflows

**Tools:**
- Docker Compose for test environment
- Test database cleanup between tests

**Test Cases:**

**Workflow 1: Authentication**
```go
func TestAuthenticationWorkflow(t *testing.T) {
    // 1. Login
    // 2. Get current user
    // 3. Change password
    // 4. Logout
    // 5. Verify token invalid
}
```

**Workflow 2: Query Execution**
```go
func TestQueryExecutionWorkflow(t *testing.T) {
    // 1. Execute SELECT query
    // 2. Verify results stored
    // 3. Check query history
    // 4. Verify pagination works
    // 5. Export to CSV/JSON
}
```

**Workflow 3: Approval Workflow**
```go
func TestApprovalWorkflow(t *testing.T) {
    // 1. Submit write query
    // 2. Create approval request
    // 3. Start transaction
    // 4. View dry run results
    // 5. Commit transaction
    // 6. Verify comments
    // 7. Verify notifications sent
}
```

---

### Frontend Testing

#### Unit Tests

**Target:** 70% component coverage

**Tools:**
- React Testing Library
- Jest
- Test Component

**Test Structure:**
```
components/
├── editor/
│   ├── SQLEditor.test.tsx
│   └── EditorToolbar.test.tsx
├── results/
│   ├── ResultsTable.test.tsx
│   └── Pagination.test.tsx
├── approvals/
│   ├── ApprovalCard.test.tsx
│   ├── ReviewDialog.test.tsx
│   └── TransactionStatus.test.tsx
└── layout/
    ├── Sidebar.test.tsx
    └── Header.test.tsx
```

#### E2E Tests

**Target:** Cover all critical user paths

**Tools:**
- Playwright

**Test Cases:**

**Path 1: Login to Query**
```typescript
test('login and execute query', async ({ page }) => {
  // Login
  await page.goto('/login');
  await page.fill('[name="email"]', 'admin@querybase.local');
  await page.fill('[name="password"]', 'admin123');
  await page.click('button[type="submit"]');

  // Navigate to editor
  await page.waitForURL('/editor');

  // Execute query
  await page.selectOption('[name="dataSource"]', '1');
  await page.fill('[name="query"]', 'SELECT * FROM users LIMIT 10');
  await page.click('button:has-text("Run")');

  // Wait for results
  await page.waitForSelector('[data-testid="results-table"]');

  // Verify results
  const rows = await page.locator('[data-testid="results-table"] tbody tr').count();
  expect(rows).toBeGreaterThan(0);
});
```

**Path 2: Approval Workflow**
```typescript
test('approve write query', async ({ page }) => {
  // Login as approver
  // Navigate to approvals
  // Click approval
  // Click approve
  // Verify status changed
});
```

**Path 3: Admin Operations**
```typescript
test('manage data sources', async ({ page }) => {
  // Login as admin
  // Navigate to data sources
  // Add new data source
  // Test connection
  // Verify added
});
```

---

### Performance Testing

#### Backend Benchmarks

**Run:**
```bash
cd internal/service
go test -bench=. -benchmem > benchmark.txt
```

**Metrics to Track:**

| Metric | Target | Current |
|--------|--------|---------|
| Query execution (1000 rows) | < 500ms | TBD |
| Pagination (10k rows) | < 100ms | TBD |
| Export CSV (10k rows) | < 1s | TBD |
| Export JSON (10k rows) | < 500ms | TBD |
| Approval creation | < 200ms | TBD |
| Health check | < 50ms | TBD |

---

#### Frontend Performance

**Run:**
```bash
npm run build
npm run analyze
npm run lighthouse
```

**Metrics to Track:**

| Metric | Target | Measurement |
|--------|--------|------------|
| First Contentful Paint | < 1.5s | Lighthouse |
| Time to Interactive | < 3s | Lighthouse |
| Speed Index | < 3s | Lighthouse |
| Bundle Size (gzipped) | < 500KB | Webpack |
| API Response Time | < 200ms | DevTools |
| Page Load Time | < 2s | DevTools |

---

## Part 4: Deployment Plan

### Week 11: Deployment

### Pre-Deployment Checklist

**Backend:**
- ✅ All tests passing (unit + integration)
- ✅ Performance benchmarks within targets
- ✅ Security audit passed
- ✅ Error handling consistent
- ✅ Logging configured
- ✅ Environment variables documented

**Frontend:**
- ✅ All component tests passing
- ✅ E2E tests passing
- ✅ Lighthouse scores > 90
- ✅ Bundle size optimized
- ✅ Environment variables configured
- ✅ API endpoints configured

### Deployment Steps

**1. Backend Deployment**

```bash
# Build for all platforms
make build-all

# Run tests
make test
make test-coverage

# Build Docker images
docker-compose -f docker/docker-compose.yml build

# Push to registry
docker push querybase/api:latest
docker push querybase/worker:latest
```

**2. Frontend Deployment**

```bash
# Build production bundle
npm run build

# Deploy to Vercel
vercel --prod

# OR deploy to Docker
docker build -t querybase-web .
docker push querybase-web:latest
```

**3. Database Migrations**

```bash
# Run migrations
make migrate-up
```

---

### Monitoring Setup

**Backend Monitoring:**
- Application logs (JSON format)
- Query execution metrics
- Error tracking (Sentry?)
- Performance monitoring

**Frontend Monitoring:**
- Page views
- API errors
- Performance metrics
- User analytics

---

## Summary Timeline

### Backend Implementation (3 weeks)

**Week 1:**
- Query pagination (Day 1-2)
- Query export (Day 3-4)

**Week 2:**
- Approval comments (Day 1-3)
- Health check (Day 4-5)

**Week 3:**
- Error handling (Day 1-2)
- Request validation (Day 3-4)
- Performance benchmarks (Day 5)

**Deliverables:**
- ✅ 4 new API endpoints
- ✅ 1 new database table
- ✅ Production-ready error handling
- ✅ Performance baseline

---

### Frontend Implementation (8 weeks)

**Week 1-2:**
- Project setup
- Authentication
- Layout

**Week 3-4:**
- SQL Editor
- Query Results

**Week 5:**
- Approval Dashboard

**Week 6:**
- Admin Features

**Week 7-8:**
- Polish
- Optimization
- Testing

**Deliverables:**
- ✅ Full-featured dashboard UI
- ✅ All core workflows working
- ✅ Production-ready quality

---

### Testing (Ongoing)

**Backend:**
- Unit tests (80% coverage)
- Integration tests (all workflows)
- Performance benchmarks

**Frontend:**
- Component tests (70% coverage)
- E2E tests (critical paths)
- Lighthouse audit

---

## Success Criteria

### Backend

**Functional:**
- ✅ All new features working correctly
- ✅ All tests passing (90/90 = 100%)
- ✅ No regression in existing features

**Performance:**
- ✅ Query execution < 500ms
- ✅ Pagination < 100ms
- ✅ Export < 1s

**Quality:**
- ✅ Error handling consistent
- ✅ Input validation thorough
- ✅ Code documented

---

### Frontend

**Functional:**
- ✅ All workflows working
- ✅ No blocking bugs
- ✅ All E2E tests passing

**Performance:**
- ✅ Page load < 2s
- ✅ Time to interactive < 3s
- ✅ Bundle size < 500KB

**Quality:**
- ✅ Lighthouse Performance > 90
- ✅ Lighthouse Accessibility > 90
- ✅ Good UX (loading, errors, empty states)

---

## Next Steps

**Immediate (This Week):**

1. **Start Backend Week 1**
   - Begin query pagination implementation
   - Write tests
   - Update documentation

2. **Setup Frontend Environment**
   - Initialize Next.js project
   - Install dependencies
   - Setup shadcn/ui

**Week 2:**
- Complete query export API
- Start approval comments
- Begin frontend authentication

---

**Documentation Updated:**

- ✅ [docs/IMPLEMENTATION_TESTING_PLAN.md](docs/IMPLEMENTATION_TESTING_PLAN.md) - This file
- ✅ [CLAUDE.md](CLAUDE.md) - Updated TODO with implementation tasks
- ✅ [README.md](README.md) - Updated with current focus

---

**Last Updated:** January 28, 2025
**Status:** Ready to Start
**Next Action:** Begin query pagination implementation
