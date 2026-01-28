# Frontend vs Backend Gap Analysis
**Dashboard UI Requirements → Architecture & Implementation Plan**

**Date:** January 28, 2025
**Status:** Ready for Implementation
**Total Features Required:** 15 new backend APIs

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Feature Categorization](#feature-categorization)
3. [Critical System Features](#critical-system-features)
4. [Main Features](#main-features)
5. [Additional Features](#additional-features)
6. [Architecture Plan](#architecture-plan)
7. [Implementation Roadmap](#implementation-roadmap)

---

## Executive Summary

The Dashboard UI requires **15 additional backend features**. While the core backend is 98% complete, the frontend needs supporting APIs organized into three categories:

| Category | Features | Effort | Timeline | Blocks Frontend? |
|----------|----------|--------|----------|------------------|
| **Critical System** | 3 | 8-11 days | 2 weeks | ✅ YES |
| **Main Features** | 6 | 14-20 days | 3-4 weeks | Partially |
| **Additional** | 6 | 7-13 days | 2-3 weeks | ❌ NO |

**Total:** 29-44 days (6-9 weeks)

---

## Feature Categorization

### 🎯 Critical System Features (3)
**Definition:** Features without which the frontend cannot function at all.

| # | Feature | Purpose | Priority |
|---|---------|---------|----------|
| 1 | **Schema Introspection API** | Autocomplete, schema browser | 🔴 CRITICAL |
| 2 | **Query Results Pagination** | Display large result sets | 🔴 CRITICAL |
| 3 | **Folder System** | Organize saved queries | 🔴 CRITICAL |

**Impact:** Frontend development is **BLOCKED** until these are complete.

---

### ⚡ Main Features (6)
**Definition:** Core functionality that provides essential UX, but workarounds exist.

| # | Feature | Purpose | Priority |
|---|---------|---------|----------|
| 4 | **Query Export API** | Download results (CSV/JSON) | 🟡 HIGH |
| 5 | **Tag System** | Query organization | 🟡 HIGH |
| 6 | **Comment System** | Approval collaboration | 🟡 HIGH |
| 7 | **Table Statistics** | Schema browser enhancement | 🟡 MEDIUM |
| 8 | **WebSocket Support** | Real-time updates | 🟡 MEDIUM |
| 9 | **Performance Metrics API** | Query monitoring | 🟡 MEDIUM |

**Impact:** Frontend can launch with workarounds, but UX will be degraded.

---

### ✨ Additional Features (6)
**Definition:** Nice-to-have enhancements that improve polish and convenience.

| # | Feature | Purpose | Priority |
|---|---------|---------|----------|
| 10 | **SQL Formatting Endpoint** | Beautify queries | 🟢 LOW |
| 11 | **Favorites System** | Quick access to queries | 🟢 LOW |
| 12 | **Health Check API** | Data source status | 🟢 LOW |
| 13 | **Usage Statistics API** | Analytics dashboard | 🟢 LOW |
| 14 | **Bulk Operations** | Batch approvals | 🟢 LOW |
| 15 | **Query Comparison API** | Diff query results | 🟢 LOW |

**Impact:** Frontend works perfectly without these; these are convenience features.

---

## Critical System Features

### 1. Schema Introspection API 🔴

**Problem:** Frontend cannot provide SQL autocomplete or show schema browser without database structure information.

**Current State:**
```yaml
✅ Can connect to data sources
✅ Can execute queries
❌ Cannot retrieve schema information
```

**Required API:**
```yaml
GET    /api/v1/datasources/:id/schema           # Full schema
GET    /api/v1/datasources/:id/tables           # Table list
GET    /api/v1/datasources/:id/tables/:name     # Table details
GET    /api/v1/datasources/:id/tables/:name/columns  # Columns
GET    /api/v1/datasources/:id/tables/:name/indexes  # Indexes
```

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Request                         │
│  GET /api/v1/datasources/{id}/schema                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Schema Handler (schema.go)                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 1. Auth middleware (JWT validation)                  │  │
│  │ 2. RBAC middleware (check can_read permission)       │  │
│  │ 3. Validate data source exists                       │  │
│  │ 4. Call SchemaService.GetSchema()                    │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│           Schema Service (service/schema.go)                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ GetSchema(dataSourceID)                               │  │
│  │   │                                                    │  │
│  │   ├─> Get data source connection                      │  │
│  │   │                                                    │  │
│  │   ├─> Query information_schema (PostgreSQL/MySQL)     │  │
│  │   │   - SELECT * FROM information_schema.tables       │  │
│  │   │   - SELECT * FROM information_schema.columns      │  │
│  │   │   - SELECT * FROM information_schema.statistics   │  │
│  │   │                                                    │  │
│  │   ├─> Build schema tree structure                     │  │
│  │   │   {                                                │  │
│  │   │     tables: [                                      │  │
│  │   │       {                                            │  │
│  │   │         name, columns: [], indexes: []            │  │
│  │   │       }                                            │  │
│  │   │     ]                                              │  │
│  │   │   }                                                │  │
│  │   │                                                    │  │
│  │   └─> Cache in Redis (5 minutes)                      │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              User Data Source Database                      │
│  - information_schema.tables                                │
│  - information_schema.columns                               │
│  - information_schema.statistics                             │
│  - pg_indexes (PostgreSQL)                                  │
│  - information_schema.statistics (MySQL)                    │
└─────────────────────────────────────────────────────────────┘
```

**Database Queries:**

**PostgreSQL:**
```sql
-- Get all tables
SELECT
    t.table_schema,
    t.table_name,
    obj_description((t.table_schema||'.'||t.table_name)::regclass, 'pg_class') as table_comment
FROM information_schema.tables t
WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY t.table_schema, t.table_name;

-- Get columns for a table
SELECT
    c.column_name,
    c.data_type,
    c.character_maximum_length,
    c.is_nullable,
    c.column_default,
    c.ordinal_position,
    CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
    fk.foreign_table_name,
    fk.foreign_column_name
FROM information_schema.columns c
LEFT JOIN (
    SELECT ku.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage ku
        ON tc.constraint_name = ku.constraint_name
    WHERE tc.constraint_type = 'PRIMARY KEY'
        AND tc.table_name = $1
) pk ON c.column_name = pk.column_name
LEFT JOIN (
    SELECT
        kcu.column_name,
        ccu.table_name AS foreign_table_name,
        ccu.column_name AS foreign_column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage ccu
        ON ccu.constraint_name = tc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_name = $1
) fk ON c.column_name = fk.column_name
WHERE c.table_name = $1
ORDER BY c.ordinal_position;

-- Get indexes for a table
SELECT
    i.relname as index_name,
    a.attname as column_name,
    ix.indisunique as is_unique,
    ix.indisprimary as is_primary
FROM pg_index ix
JOIN pg_class t ON t.oid = ix.indrelid
JOIN pg_class i ON i.oid = ix.indexrelid
JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
WHERE t.relname = $1
ORDER BY i.relname, a.attnum;
```

**MySQL:**
```sql
-- Get all tables
SELECT
    TABLE_SCHEMA as table_schema,
    TABLE_NAME as table_name,
    TABLE_COMMENT as table_comment
FROM information_schema.TABLES
WHERE TABLE_SCHEMA NOT IN ('information_schema', 'mysql', 'performance_schema')
ORDER BY TABLE_SCHEMA, TABLE_NAME;

-- Get columns for a table
SELECT
    COLUMN_NAME as column_name,
    DATA_TYPE as data_type,
    CHARACTER_MAXIMUM_LENGTH as character_maximum_length,
    IS_NULLABLE as is_nullable,
    COLUMN_DEFAULT as column_default,
    ORDINAL_POSITION as ordinal_position,
    CASE WHEN COLUMN_KEY = 'PRI' THEN true ELSE false END as is_primary_key,
    REFERENCED_TABLE_NAME as foreign_table_name,
    REFERENCED_COLUMN_NAME as foreign_column_name
FROM information_schema.COLUMNS c
LEFT JOIN information_schema.KEY_COLUMN_USAGE kcu
    ON c.TABLE_NAME = kcu.TABLE_NAME
    AND c.COLUMN_NAME = kcu.COLUMN_NAME
    AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
WHERE c.TABLE_NAME = $1
ORDER BY c.ORDINAL_POSITION;

-- Get indexes for a table
SELECT
    INDEX_NAME as index_name,
    COLUMN_NAME as column_name,
    NOT NON_UNIQUE as is_unique,
    INDEX_NAME = 'PRIMARY' as is_primary
FROM information_schema.STATISTICS
WHERE TABLE_NAME = $1
ORDER BY INDEX_NAME, SEQ_IN_INDEX;
```

**Implementation Files:**

```
internal/
├── models/
│   └── schema.go                    # NEW: Schema models
│       type TableInfo struct
│       type ColumnInfo struct
│       type IndexInfo struct
│       type SchemaResponse struct
│
├── service/
│   └── schema.go                    # NEW: Schema service
│       type SchemaService struct
│       func (s *SchemaService) GetSchema(dataSourceID string) (*SchemaResponse, error)
│       func (s *SchemaService) GetTables(dataSourceID string) ([]TableInfo, error)
│       func (s *SchemaService) GetTableColumns(dataSourceID, tableName string) ([]ColumnInfo, error)
│       func (s *SchemaService) GetIndexes(dataSourceID, tableName string) ([]IndexInfo, error)
│       func (s *SchemaService) cacheSchema(key string, data interface{}, duration time.Duration)
│       func (s *SchemaService) getCachedSchema(key string) (interface{}, bool)
│
├── api/handlers/
│   └── schema.go                    # NEW: Schema handlers
│       func (h *Handler) GetSchema(c *gin.Context)
│       func (h *Handler) GetTables(c *gin.Context)
│       func (h *Handler) GetTableDetails(c *gin.Context)
│
├── api/dto/
│   └── schema.go                    # NEW: Schema DTOs
│       type SchemaResponseDTO
│       type TableInfoDTO
│       type ColumnInfoDTO
│       type IndexInfoDTO
│
└── api/routes/
    └── routes.go                    # UPDATE: Add schema routes
```

**DTOs:**

```go
// internal/api/dto/schema.go
package dto

type SchemaResponseDTO struct {
    DataSourceID string      `json:"dataSourceId"`
    Schema       SchemaDTO   `json:"schema"`
    CachedAt     *time.Time  `json:"cachedAt,omitempty"`
}

type SchemaDTO struct {
    Tables []TableInfoDTO `json:"tables"`
}

type TableInfoDTO struct {
    Name        string         `json:"name"`
    Schema      string         `json:"schema"`
    Comment     string         `json:"comment,omitempty"`
    Columns     []ColumnInfoDTO `json:"columns"`
    Indexes     []IndexInfoDTO  `json:"indexes"`
}

type ColumnInfoDTO struct {
    Name             string  `json:"name"`
    Type             string  `json:"type"`
    MaxLength        *int    `json:"maxLength,omitempty"`
    Nullable         bool    `json:"nullable"`
    DefaultValue     *string `json:"defaultValue,omitempty"`
    Position         int     `json:"position"`
    IsPrimaryKey     bool    `json:"isPrimaryKey"`
    IsForeignKey     bool    `json:"isForeignKey"`
    ReferencedTable  *string `json:"referencedTable,omitempty"`
    ReferencedColumn *string `json:"referencedColumn,omitempty"`
}

type IndexInfoDTO struct {
    Name      string   `json:"name"`
    Columns   []string `json:"columns"`
    IsUnique  bool     `json:"isUnique"`
    IsPrimary bool     `json:"isPrimary"`
}
```

**Caching Strategy:**

```go
// Cache schema in Redis to reduce database load
// Key: "schema:{dataSourceID}"
// TTL: 5 minutes
// Invalidate on: Data source update, schema change detection

func (s *SchemaService) GetSchema(dataSourceID string) (*SchemaResponse, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("schema:%s", dataSourceID)
    if cached, found := s.getCachedSchema(cacheKey); found {
        return cached.(*SchemaResponse), nil
    }

    // Query database
    schema, err := s.querySchema(dataSourceID)
    if err != nil {
        return nil, err
    }

    // Cache for 5 minutes
    s.cacheSchema(cacheKey, schema, 5*time.Minute)

    return schema, nil
}
```

**Testing:**

```go
// internal/service/schema_test.go
func TestSchemaService_GetSchema(t *testing.T) {
    tests := []struct {
        name        string
        dataSourceID string
        wantTables  int
        wantErr     bool
    }{
        {
            name:        "valid data source",
            dataSourceID: testDataSourceID,
            wantTables:  5, // Expected table count
            wantErr:     false,
        },
        {
            name:        "invalid data source",
            dataSourceID: "invalid-uuid",
            wantTables:  0,
            wantErr:     true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            schema, err := service.GetSchema(tt.dataSourceID)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetSchema() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && len(schema.Schema.Tables) != tt.wantTables {
                t.Errorf("GetSchema() got %d tables, want %d", len(schema.Schema.Tables), tt.wantTables)
            }
        })
    }
}
```

**Estimated Effort:** 3-4 days

---

### 2. Query Results Pagination API 🔴

**Problem:** Frontend cannot display large result sets (10,000+ rows) without pagination.

**Current State:**
```yaml
✅ Query execution works
✅ Results stored in JSONB
❌ No pagination endpoint
❌ Entire result set returned at once
```

**Required API:**
```yaml
GET /api/v1/queries/:id/results?page=1&limit=100
GET /api/v1/queries/:id/results?sort_column=id&sort_direction=asc
```

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Request                         │
│  GET /api/v1/queries/{id}/results?page=1&limit=100          │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Query Handler (query.go)                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 1. Auth middleware                                    │  │
│  │ 2. Validate query exists                             │  │
│  │ 3. Check user has permission to view                 │  │
│  │ 4. Parse pagination params (page, limit, sort)       │  │
│  │ 5. Call QueryService.GetPaginatedResults()           │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│           Query Service (service/query.go)                   │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ GetPaginatedResults(queryID, page, limit, sort)      │  │
│  │   │                                                    │  │
│  │   ├─> Fetch QueryResult from DB                      │  │
│  │   │   - Get stored rows (JSONB)                      │  │
│  │   │   - Get column metadata                          │  │
│  │   │                                                    │  │
│  │   ├─> Calculate pagination                           │  │
│  │   │   - offset = (page - 1) * limit                  │  │
│  │   │   - Extract rows[offset : offset+limit]          │  │
│  │   │   - Use JSONB array slicing                      │  │
│  │   │                                                    │  │
│  │   ├─> Sort if requested                              │  │
│  │   │   - Sort rows in memory by column                │  │
│  │   │   - Handle ASC/DESC                              │  │
│  │   │                                                    │  │
│  │   └─> Return paginated result                        │  │
│  │       - rows (subset)                                 │  │
│  │       - pagination metadata                           │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              PostgreSQL (QueryBase DB)                       │
│  SELECT id, rows, column_names, column_types                │
│  FROM query_results                                         │
│  WHERE query_id = $1                                        │
│  LIMIT 1                                                    │
└─────────────────────────────────────────────────────────────┘
```

**JSONB Pagination:**

PostgreSQL JSONB supports array slicing:

```sql
-- Get paginated subset from JSONB array
SELECT
    id,
    column_names,
    column_types,
    jsonb_path_query_array(rows, '$[*]') as rows,
    jsonb_array_length(rows) as total_rows
FROM query_results
WHERE query_id = $1;
```

**Go Implementation:**

```go
// internal/service/query.go
func (s *QueryService) GetPaginatedResults(
    queryID uuid.UUID,
    page, limit int,
    sortColumn, sortDirection string,
) (*PaginatedResultDTO, error) {
    // Fetch stored result
    var result QueryResult
    if err := s.db.Where("query_id = ?", queryID).First(&result).Error; err != nil {
        return nil, err
    }

    // Parse stored rows
    var rows []map[string]interface{}
    if err := json.Unmarshal([]byte(result.Rows), &rows); err != nil {
        return nil, err
    }

    // Sort if requested
    if sortColumn != "" {
        sortRows(rows, sortColumn, sortDirection)
    }

    // Paginate
    totalRows := len(rows)
    totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))
    offset := (page - 1) * limit

    var paginatedRows []map[string]interface{}
    if offset < totalRows {
        end := offset + limit
        if end > totalRows {
            end = totalRows
        }
        paginatedRows = rows[offset:end]
    }

    return &PaginatedResultDTO{
        QueryID:    queryID,
        Columns:    result.ColumnNames, // Already strings
        ColumnTypes: result.ColumnTypes,
        Rows:       paginatedRows,
        Pagination: PaginationMeta{
            Page:       page,
            Limit:      limit,
            TotalRows:  totalRows,
            TotalPages: totalPages,
        },
    }, nil
}

func sortRows(rows []map[string]interface{}, column, direction string) {
    sort.Slice(rows, func(i, j int) bool {
        valI := rows[i][column]
        valJ := rows[j][column]

        // Type-aware comparison
        switch vi := valI.(type) {
        case float64:
            vj := valJ.(float64)
            if direction == "asc" {
                return vi < vj
            }
            return vi > vj
        case string:
            vj := valJ.(string)
            if direction == "asc" {
                return vi < vj
            }
            return vi > vj
        default:
            // Fallback to string comparison
            viStr := fmt.Sprintf("%v", vi)
            vjStr := fmt.Sprintf("%v", valJ)
            if direction == "asc" {
                return viStr < vjStr
            }
            return viStr > vjStr
        }
    })
}
```

**DTOs:**

```go
// internal/api/dto/query.go
type PaginatedResultDTO struct {
    QueryID     uuid.UUID        `json:"queryId"`
    Columns     []string         `json:"columns"`
    ColumnTypes []string         `json:"columnTypes"`
    Rows        []map[string]any `json:"rows"`
    Pagination  PaginationMeta   `json:"pagination"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    TotalRows  int `json:"totalRows"`
    TotalPages int `json:"totalPages"`
}
```

**Estimated Effort:** 2-3 days

---

### 3. Folder System 🔴

**Problem:** Frontend cannot organize saved queries without folder structure.

**Current State:**
```yaml
✅ Can save queries
✅ Can list queries (flat list)
❌ No folder organization
```

**Required API:**
```yaml
POST   /api/v1/folders                    # Create folder
GET    /api/v1/folders                    # List folders (tree)
GET    /api/v1/folders/:id                # Get folder details
PUT    /api/v1/folders/:id                # Update folder
DELETE /api/v1/folders/:id                # Delete folder
POST   /api/v1/folders/:id/move           # Move folder
POST   /api/v1/queries/:id/move           # Move query to folder
```

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                   Folder Data Model                         │
│                                                              │
│  ┌──────────────┐       ┌──────────────┐                   │
│  │   Folder     │──────>│  Folder      │ (parent-child)    │
│  │  (root)      │       │  (subfolder) │                   │
│  └──────┬───────┘       └──────────────┘                   │
│         │                                                     │
│         │                                                     │
│         ▼                                                     │
│  ┌──────────────┐                                           │
│  │    Query     │ (belongs to folder)                       │
│  └──────────────┘                                           │
└─────────────────────────────────────────────────────────────┘

                    Folder Tree Structure

Root Folder (id: null, name: "My Queries")
├── Team Queries (folder)
│   ├── Daily Reports (folder)
│   │   ├── Daily Sales Report (query)
│   │   └── Daily User Stats (query)
│   └── Monthly Reports (folder)
│       └── Monthly Revenue (query)
└── Personal (folder)
    ├── My Drafts (folder)
    └── Favorites (not a folder, just filtered)
```

**Database Migration:**

```sql
-- Migration 000005
CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_folders_parent ON folders(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_folders_created_by ON folders(created_by) WHERE deleted_at IS NULL;

-- Add folder_id to queries
ALTER TABLE queries ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;
CREATE INDEX idx_queries_folder ON queries(folder_id) WHERE deleted_at IS NULL;
```

**Model:**

```go
// internal/models/folder.go
package models

type Folder struct {
    ID        uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
    Name      string     `gorm:"size:255;not null" json:"name"`
    ParentID  *uuid.UUID `gorm:"type:uuid" json:"parentId,omitempty"`
    Parent    *Folder    `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children  []*Folder  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
    Queries   []*Query   `gorm:"foreignKey:FolderID" json:"queries,omitempty"`
    CreatedBy uuid.UUID  `gorm:"type:uuid;not null" json:"createdBy"`
    CreatedByUser *User   `gorm:"foreignKey:CreatedBy" json:"createdByUser,omitempty"`
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt time.Time  `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate - set UUID
func (f *Folder) BeforeCreate(tx *gorm.DB) error {
    if f.ID == uuid.Nil {
        f.ID = uuid.New()
    }
    return nil
}
```

**Service:**

```go
// internal/service/folder.go
package service

type FolderService struct {
    db *gorm.DB
}

func (s *FolderService) CreateFolder(userID uuid.UUID, name string, parentID *uuid.UUID) (*Folder, error) {
    folder := &Folder{
        ID:        uuid.New(),
        Name:      name,
        ParentID:  parentID,
        CreatedBy: userID,
    }

    if err := s.db.Create(folder).Error; err != nil {
        return nil, err
    }

    return folder, nil
}

func (s *FolderService) GetFolderTree(userID uuid.UUID) ([]Folder, error) {
    var folders []Folder

    // Get all folders for user
    if err := s.db.Where("created_by = ? AND deleted_at IS NULL", userID).
        Order("name ASC").
        Find(&folders).Error; err != nil {
        return nil, err
    }

    // Build tree structure
    return s.buildTree(folders, nil), nil
}

func (s *FolderService) buildTree(folders []Folder, parentID *uuid.UUID) []Folder {
    var tree []Folder

    for _, folder := range folders {
        if (parentID == nil && folder.ParentID == nil) ||
           (parentID != nil && folder.ParentID != nil && *folder.ParentID == *parentID) {
            folder.Children = s.buildTree(folders, &folder.ID)
            tree = append(tree, folder)
        }
    }

    return tree
}

func (s *FolderService) MoveFolder(folderID, newParentID uuid.UUID) error {
    return s.db.Model(&Folder{}).
        Where("id = ?", folderID).
        Update("parent_id", newParentID).Error
}

func (s *FolderService) DeleteFolder(folderID uuid.UUID) error {
    // Soft delete (cascade deletes queries in folder too)
    return s.db.Delete(&Folder{}, "id = ?", folderID).Error
}

func (s *FolderService) MoveQuery(queryID, folderID uuid.UUID) error {
    return s.db.Model(&Query{}).
        Where("id = ?", queryID).
        Update("folder_id", folderID).Error
}
```

**Estimated Effort:** 3-4 days

---

## Main Features

### 4. Query Export API 🟡

**Purpose:** Allow users to download query results (CSV, JSON, Excel).

**Required API:**
```yaml
GET /api/v1/queries/:id/export?format=csv
GET /api/v1/queries/:id/export?format=json
```

**Architecture:**

```
Frontend Request → Handler → Service → Export Format → Download
                    │
                    ▼
            Generate CSV/JSON
                    │
                    ▼
            Set Content-Disposition header
                    │
                    ▼
            Return file response
```

**Implementation:**

```go
// internal/service/export.go
func (s *ExportService) ExportQueryResults(queryID uuid.UUID, format string) ([]byte, string, error) {
    var result QueryResult
    if err := s.db.Where("query_id = ?", queryID).First(&result).Error; err != nil {
        return nil, "", err
    }

    var rows []map[string]interface{}
    json.Unmarshal([]byte(result.Rows), &rows)

    switch format {
    case "csv":
        return s.exportToCSV(result.ColumnNames, rows)
    case "json":
        return s.exportToJSON(result.ColumnNames, rows)
    default:
        return nil, "", fmt.Errorf("unsupported format: %s", format)
    }
}

func (s *ExportService) exportToCSV(columns []string, rows []map[string]interface{}) ([]byte, string, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Write header
    writer.Write(columns)

    // Write rows
    for _, row := range rows {
        var record []string
        for _, col := range columns {
            val := row[col]
            record = append(record, fmt.Sprintf("%v", val))
        }
        writer.Write(record)
    }

    writer.Flush()
    return buf.Bytes(), "text/csv", nil
}
```

**Estimated Effort:** 1-2 days

---

### 5. Tag System 🟡

**Purpose:** Add tags/labels to queries for better organization.

**Required API:**
```yaml
GET    /api/v1/tags                      # List all tags
POST   /api/v1/tags                      # Create tag
DELETE /api/v1/tags/:id                  # Delete tag
POST   /api/v1/queries/:id/tags          # Add tag to query
DELETE /api/v1/queries/:id/tags/:tagId   # Remove tag
GET    /api/v1/queries?tag=important     # Filter by tag
```

**Database Migration:**

```sql
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
```

**Estimated Effort:** 2-3 days

---

### 6. Comment System 🟡

**Purpose:** Allow approvers to comment on approval requests.

**Required API:**
```yaml
POST   /api/v1/approvals/:id/comments     # Add comment
GET    /api/v1/approvals/:id/comments     # List comments
PUT    /api/v1/approvals/:id/comments/:id # Update comment
DELETE /api/v1/approvals/:id/comments/:id # Delete comment
```

**Database Migration:**

```sql
CREATE TABLE approval_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Estimated Effort:** 2-3 days

---

### 7. Table Statistics API 🟡

**Purpose:** Show row counts, table sizes in schema browser.

**Required API:**
```yaml
GET /api/v1/datasources/:id/tables/:name/stats
```

**Response:**
```json
{
  "tableName": "users",
  "rowCount": 15432,
  "sizeBytes": 5242880,
  "indexSizeBytes": 2097152,
  "columns": 12,
  "indexes": 5
}
```

**Estimated Effort:** 1-2 days

---

### 8. WebSocket Support 🟡

**Purpose:** Real-time updates for query status, transaction changes.

**Required:**
```yaml
WS /ws/notifications           # General notifications
WS /ws/queries/:id             # Query status updates
WS /ws/approvals               # Approval updates
WS /ws/transactions/:id        # Transaction status
```

**Architecture:**

```
Frontend ←WebSocket→ Hub ←Broadcast→ All Connected Clients
                ↓
         Event Handler
                ↓
         Backend Events
    (query completed,
     transaction committed,
     approval received)
```

**Estimated Effort:** 4-5 days

---

### 9. Performance Metrics API 🟡

**Purpose:** Show query execution statistics.

**Required API:**
```yaml
GET /api/v1/metrics/performance?start_date=2024-01-01&end_date=2024-01-31
```

**Response:**
```json
{
  "summary": {
    "totalQueries": 1234,
    "avgExecutionTime": "234ms",
    "p95ExecutionTime": "500ms",
    "slowQueries": 23
  },
  "trend": [
    {"date": "2024-01-01", "avgTime": 200, "queryCount": 100}
  ]
}
```

**Estimated Effort:** 2-3 days

---

## Additional Features

### 10. SQL Formatting Endpoint 🟢

**Purpose:** Beautify SQL queries.

**Required API:**
```yaml
POST /api/v1/queries/format
```

**Request:**
```json
{
  "query": "select*from users where id=1"
}
```

**Response:**
```json
{
  "formatted": "SELECT *\nFROM users\nWHERE id = 1;"
}
```

**Estimated Effort:** 1-2 days

---

### 11. Favorites System 🟢

**Purpose:** Quick access to important queries.

**Database Migration:**
```sql
ALTER TABLE queries ADD COLUMN is_favorite BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_queries_favorite ON queries(created_by, is_favorite) WHERE is_favorite = TRUE;
```

**Required API:**
```yaml
POST   /api/v1/queries/:id/favorite      # Mark favorite
DELETE /api/v1/queries/:id/favorite      # Unmark favorite
GET    /api/v1/queries/favorites         # List favorites
```

**Estimated Effort:** 1 day

---

### 12. Health Check API 🟢

**Purpose:** Show data source connection status.

**Required API:**
```yaml
GET /api/v1/datasources/:id/health
```

**Response:**
```json
{
  "status": "healthy",
  "latency": "12ms",
  "version": "PostgreSQL 15.2",
  "connections": 15
}
```

**Estimated Effort:** 1 day

---

### 13. Usage Statistics API 🟢

**Purpose:** Analytics dashboard.

**Required API:**
```yaml
GET /api/v1/datasources/:id/stats
```

**Response:**
```json
{
  "totalQueries": 1234,
  "last30Days": 456,
  "lastQueryTime": "2024-01-15T10:28:00Z"
}
```

**Estimated Effort:** 1-2 days

---

### 14. Bulk Operations 🟢

**Purpose:** Batch approve/reject multiple approvals.

**Required API:**
```yaml
POST /api/v1/approvals/bulk-review
```

**Request:**
```json
{
  "action": "approve",
  "approvalIds": ["uuid1", "uuid2", "uuid3"],
  "comment": "Batch approved"
}
```

**Estimated Effort:** 1-2 days

---

### 15. Query Comparison API 🟢

**Purpose:** Compare two query results side-by-side.

**Required API:**
```yaml
POST /api/v1/queries/compare
```

**Request:**
```json
{
  "queryId1": "uuid",
  "queryId2": "uuid"
}
```

**Estimated Effort:** 2-3 days

---

## Architecture Plan

### System Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                         Frontend Layer                         │
│  (Next.js + TypeScript + Tailwind + Monaco + shadcn/ui)        │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ SQL Editor   │  │ Schema       │  │ Approval     │         │
│  │ (Monaco)     │  │ Browser      │  │ Dashboard    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            │ HTTP/WebSocket
                            ▼
┌────────────────────────────────────────────────────────────────┐
│                       API Gateway (Go)                         │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ Middleware Chain                                         │ │
│  │ - Auth (JWT)                                             │ │
│  │ - RBAC (permissions)                                    │ │
│  │ - CORS (TODO)                                           │ │
│  │ - Logging (TODO)                                        │ │
│  │ - Rate Limiting (TODO)                                  │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                 │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────────┐  │
│  │ Query      │ │ Schema     │ │ Folder     │ │ WebSocket │  │
│  │ Handlers   │ │ Handlers   │ │ Handlers   │ │ Handler   │  │
│  └────────────┘ └────────────┘ └────────────┘ └───────────┘  │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│                        Service Layer                           │
│                                                                 │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────────┐  │
│  │ Query      │ │ Schema     │ │ Folder     │ │ Export    │  │
│  │ Service    │ │ Service    │ │ Service    │ │ Service   │  │
│  └────────────┘ └────────────┘ └────────────┘ └───────────┘  │
│                                                                 │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────────┐  │
│  │ Tag        │ │ Comment    │ │ Analytics  │ │ WebSocket │  │
│  │ Service    │ │ Service    │ │ Service    │ │ Hub       │  │
│  └────────────┘ └────────────┘ └────────────┘ └───────────┘  │
└───────────────────────────┬────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            │                               │
            ▼                               ▼
┌──────────────────────┐         ┌──────────────────────┐
│   PostgreSQL DB      │         │      Redis Cache     │
│   (QueryBase)        │         │  (Schema cache,      │
│                      │         │   sessions)          │
│ - users              │         └──────────────────────┘
│ - groups             │
│ - data_sources       │
│ - queries            │
│ - folders ✨ NEW     │
│ - tags ✨ NEW        │
│ - approval_comments ✨ │
│ - query_results      │
└──────────────────────┘
            │
            │ (encrypted connections)
            ▼
┌────────────────────────────────────────────────────────────────┐
│                   User Data Sources                            │
│  ┌──────────────────┐         ┌──────────────────┐            │
│  │  PostgreSQL      │         │     MySQL         │            │
│  │  (User DB 1)     │         │   (User DB 2)     │            │
│  └──────────────────┘         └──────────────────┘            │
└────────────────────────────────────────────────────────────────┘
```

### New Components Required

**Critical System:**
- ✅ Schema Service (introspection)
- ✅ Pagination in Query Service
- ✅ Folder Service + Model

**Main Features:**
- ✅ Export Service
- ✅ Tag Service + Model
- ✅ Comment Service + Model
- ✅ Analytics Service
- ✅ WebSocket Hub

**Additional:**
- ✅ Formatter Service
- ✅ Comparison Service

### Database Schema Changes

**New Tables:**
```sql
-- Folders (Critical)
CREATE TABLE folders (...);

-- Tags (Main)
CREATE TABLE tags (...);
CREATE TABLE query_tags (...);

-- Comments (Main)
CREATE TABLE approval_comments (...);

-- Query additions
ALTER TABLE queries ADD COLUMN folder_id UUID;
ALTER TABLE queries ADD COLUMN is_favorite BOOLEAN;
```

---

## Implementation Roadmap

### Phase 1: Critical System Features (Week 1-2) 🔴

**Week 1: Schema Introspection API**
- Day 1-2: Schema service and models
- Day 3-4: Schema handlers and routes
- Day 5: Testing and documentation

**Week 2: Pagination + Folders**
- Day 1-2: Query pagination
- Day 3-4: Folder system (migration, model, service, handlers)
- Day 5: Testing and documentation

**Deliverables:**
- ✅ Schema API (5 endpoints)
- ✅ Query pagination API
- ✅ Folder CRUD API (6 endpoints)

**Frontend Unblocked:** SQL Editor, Schema Browser, Saved Queries

---

### Phase 2: Main Features (Week 3-4) 🟡

**Week 3: Export + Tags + Comments**
- Day 1: Export service
- Day 2-3: Tag system
- Day 4-5: Comment system

**Week 4: Analytics + WebSocket**
- Day 1-2: Table statistics
- Day 3-4: Performance metrics
- Day 5: WebSocket hub (basic)

**Deliverables:**
- ✅ Export API (2 formats)
- ✅ Tag CRUD API (5 endpoints)
- ✅ Comment CRUD API (4 endpoints)
- ✅ Statistics API (2 endpoints)
- ✅ WebSocket endpoints (4 topics)

**Frontend Enhanced:** Real-time updates, better organization, analytics

---

### Phase 3: Additional Features (Week 5-6) 🟢

**Week 5: Polish Features**
- Day 1: SQL formatter
- Day 2: Favorites system
- Day 3: Health check API
- Day 4-5: Usage statistics

**Week 6: Advanced Features**
- Day 1-2: Bulk operations
- Day 3-4: Query comparison
- Day 5: Testing and documentation

**Deliverables:**
- ✅ SQL formatting API
- ✅ Favorites API
- ✅ Health check API
- ✅ Usage statistics API
- ✅ Bulk operations API
- ✅ Query comparison API

**Frontend Complete:** Full feature parity with Bytebase CE

---

## Summary

### Files to Create

**Models (5):**
- `internal/models/folder.go`
- `internal/models/tag.go`
- `internal/models/schema.go`

**Services (8):**
- `internal/service/schema.go`
- `internal/service/folder.go`
- `internal/service/tag.go`
- `internal/service/export.go`
- `internal/service/comment.go`
- `internal/service/analytics.go`
- `internal/service/formatter.go`
- `internal/service/comparison.go`

**Handlers (6):**
- `internal/api/handlers/schema.go`
- `internal/api/handlers/folder.go`
- `internal/api/handlers/tag.go`
- `internal/api/handlers/export.go`
- `internal/api/handlers/analytics.go`
- `internal/api/handlers/websocket.go`

**DTOs (5):**
- `internal/api/dto/schema.go`
- `internal/api/dto/folder.go`
- `internal/api/dto/tag.go`
- `internal/api/dto/export.go`
- `internal/api/dto/analytics.go`

**Migrations (2):**
- `migrations/000005_add_folders_and_tags.up.sql`
- `migrations/000006_add_approval_comments.up.sql`

### Total Effort

- **Critical:** 8-11 days (2 weeks)
- **Main:** 14-20 days (3-4 weeks)
- **Additional:** 7-13 days (2-3 weeks)

**Grand Total:** 29-44 days (6-9 weeks)

### Recommendation

**Start Phase 1 immediately** to unblock frontend development. Implement Schema API, Query Pagination, and Folder System before frontend begins Phase 2 (SQL Editor).

---

**Last Updated:** January 28, 2025
**Status:** Ready for Implementation
**Next Step:** Begin Schema Introspection API implementation
