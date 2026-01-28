# QueryBase Schema Features

This document describes the database schema inspection and autocomplete features in QueryBase.

## Features

### 1. Schema Browser

The Schema Browser is a sidebar component that displays the database schema of the selected data source. It provides:

- **Table List**: Shows all tables in the selected database
- **Column Details**: Displays column names, data types, and constraints
- **Expandable/Collapsible**: Click on a table to view its columns
- **Search**: Filter tables by name
- **Visual Indicators**:
  - 🔑 PK - Primary Key
  - 🗝️ FK - Foreign Key
  - NULL - Nullable columns
  - Column counts and data types

#### How to Use

1. Navigate to the Query Editor page
2. The Schema Browser appears on the left side by default
3. Select a data source from the dropdown at the top of the Schema Browser
4. The schema will automatically load and display all tables
5. Click on a table name to expand and view its columns
6. Use the search box to filter tables
7. Click "Expand All" or "Collapse" to toggle all tables at once

### 2. SQL Autocomplete

The SQL Editor now includes intelligent autocomplete powered by Monaco Editor. It provides:

- **SQL Keywords**: All standard SQL keywords (SELECT, FROM, WHERE, etc.)
- **Table Names**: Suggests table names from the selected data source
- **Column Names**: Suggests columns in the format `table_name.column_name`
- **Context-Aware**: Suggests tables after FROM/JOIN/INTO clauses
- **Type Information**: Shows data types and constraints in suggestions
- **Function Signatures**: Provides signatures for common SQL functions (COUNT, SUM, AVG, etc.)

#### How to Use

1. Select a data source in the Query Editor
2. The schema will automatically load in the background
3. Start typing a query in the SQL Editor
4. Press `Ctrl+Space` (or `Cmd+Space` on Mac) to trigger autocomplete
5. Or just start typing and autocomplete will appear automatically

#### Example Usage

```sql
-- Type "SEL" and autocomplete will suggest "SELECT"
-- After "FROM ", type table name or press Ctrl+Space
SELECT * FROM

-- After table name, type "." to see columns
SELECT users.id, users.username, users.email
FROM users

-- Join suggestions
SELECT u.username, q.query_text
FROM users u
JOIN queries q ON u.id = q.user_id
```

### 3. Real-Time Schema Updates

The system uses WebSocket connections to receive real-time schema updates:

- When you select a data source, the WebSocket subscribes to schema updates
- If the database schema changes, the Schema Browser will update automatically
- The autocomplete suggestions will also update in real-time

### 4. API Endpoints

The following API endpoints power the schema features:

#### Schema Inspection

- `GET /api/v1/datasources/:id/schema` - Get complete database schema
- `GET /api/v1/datasources/:id/tables` - List all tables
- `GET /api/v1/datasources/:id/table?table=name` - Get detailed table information
- `GET /api/v1/datasources/:id/search?q=query` - Search for tables

#### WebSocket

- `GET /ws` - WebSocket endpoint for real-time schema updates

**WebSocket Message Types:**

- `get_schema` - Request schema for a data source
- `subscribe_schema` - Subscribe to schema updates
- `schema` - Receive schema data
- `schema_update` - Real-time schema update notification
- `connected` - Connection established confirmation
- `error` - Error messages

## Architecture

### Frontend Components

```
web/src/
├── components/
│   └── query/
│       ├── SchemaBrowser.tsx      # Schema browser sidebar
│       ├── SQLEditor.tsx            # Enhanced with autocomplete
│       └── QueryExecutor.tsx        # Updated with schema integration
├── stores/
│   └── schema-store.ts             # Schema state management
├── lib/
│   ├── api-client.ts               # Updated with schema API methods
│   └── websocket.ts                # WebSocket service
└── types/
    └── index.ts                    # Updated with schema types
```

### Data Flow

1. **User selects data source** → Triggers schema load
2. **API request** → `GET /api/v1/datasources/:id/schema`
3. **Schema stored** → In Zustand store (`schema-store.ts`)
4. **Components subscribe** → React components use the store
5. **Autocomplete** → Monaco editor reads from store
6. **WebSocket updates** → Store updates automatically

## Schema Data Structure

```typescript
interface DatabaseSchema {
  data_source_id: string;
  data_source_name: string;
  database_type: string;  // "postgresql" | "mysql"
  database_name: string;
  tables: TableInfo[];
  schemas?: string[];      // For databases with multiple schemas
}

interface TableInfo {
  table_name: string;
  schema: string;          // Schema name (e.g., "public")
  columns: SchemaColumnInfo[];
  indexes?: IndexInfo[];
}

interface SchemaColumnInfo {
  column_name: string;
  data_type: string;       // e.g., "uuid", "text", "integer"
  is_nullable: boolean;
  column_default?: string;
  is_primary_key: boolean;
  is_foreign_key: boolean;
}
```

## Supported Databases

### PostgreSQL
- ✅ Full schema inspection
- ✅ Multiple schema support (public, custom schemas)
- ✅ Primary key detection
- ✅ Foreign key detection
- ✅ Default values
- ✅ Nullable columns

### MySQL
- ✅ Full schema inspection
- ✅ Primary key detection
- ✅ Default values
- ✅ Nullable columns
- ⚠️ Foreign key detection (coming soon)

## Usage Examples

### Example 1: Exploring a Database

```
1. Open Query Editor
2. Select "PostgreSQL Test" from Schema Browser dropdown
3. See all tables listed (users, groups, queries, etc.)
4. Click on "users" table to see columns
5. Note the PK indicator on the "id" column
```

### Example 2: Writing a Query with Autocomplete

```
1. Select a data source
2. Type: SEL
3. Autocomplete suggests: SELECT
4. Continue: SELECT * FROM
5. Autocomplete shows all tables
6. Type: use
7. Select "users" from suggestions
8. Type: . (dot)
9. Autocomplete shows all columns from users table
10. Complete query: SELECT users.id, users.username FROM users
```

### Example 3: Complex Join with Autocomplete

```
1. SELECT u.username, q.query_text, q.status
2. FROM use
3. Autocomplete suggests: users
4. JOIN qu
5. Autocomplete suggests: queries
6. ON u.id = q.user_id
7. Full query with table and column suggestions!
```

## Performance Considerations

- Schema data is cached in the Zustand store
- Only one API call per data source
- WebSocket updates ensure cache stays fresh
- Autocomplete operates on cached data (instant response)
- Large schemas (>100 tables) are handled efficiently

## Troubleshooting

### Schema Not Loading

1. Check if data source is selected
2. Verify data source connection (test connection)
3. Check browser console for errors
4. Verify API endpoint is accessible

### Autocomplete Not Working

1. Ensure a data source is selected
2. Wait for schema to finish loading
3. Try pressing `Ctrl+Space` manually
4. Check if Monaco Editor initialized correctly

### WebSocket Not Connected

1. Check browser console for WebSocket errors
2. Verify backend is running on correct port
3. Check firewall/proxy settings

## Future Enhancements

- [ ] Foreign key relationships visualization
- [ ] Table relationships diagram
- [ ] Query templates based on schema
- [ ] One-click query generation
- [ ] Schema diff between data sources
- [ ] Export schema documentation
- [ ] Search across all columns
- [ ] Filter columns by type
