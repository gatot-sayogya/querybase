# Schema Browser Guide

Learn how to explore and understand database schemas in QueryBase.

## Overview

The Schema Browser allows you to explore database structures, view tables and columns, and understand data types without writing queries.

## Features

### Schema Exploration

- **Browse Tables**: View all tables in connected data sources
- **View Columns**: See column names, data types, and constraints
- **Search Tables**: Quick filter to find specific tables
- **Sort Tables**: A-Z or Z-A ordering
- **Column Details**: Type, nullable, default values

### Schema Synchronization

The schema browser keeps data source schemas up-to-date through:

1. **Automatic Polling**: Every 60 seconds
2. **Manual Sync**: "Sync Now" button for immediate refresh
3. **Background Worker**: Syncs all schemas every 5 minutes

## How to Use

### Access Schema Browser

1. Navigate to **Dashboard**
2. Select a **Data Source** from the dropdown
3. Click **Schema Browser** tab

### Browse Tables

1. **View All Tables**
   - Scroll through table list
   - Tables sorted A-Z by default
   - Search bar for quick filtering

2. **View Table Details**
   - Click on a table name
   - See all columns in that table
   - Column information:
     - Name
     - Data type (integer, varchar, timestamp, etc.)
     - Nullable (yes/no)
     - Default value (if any)
     - Primary key indicator

### Search Tables

1. Use the **Search** bar at top
2. Type table name (case-insensitive)
3. Results filter as you type
4. Click **Clear** to reset

### Sync Schema Manually

1. Look for **"Last sync: X minutes ago"** text
2. Click **"Sync Now"** button
3. Schema refreshes immediately
4. Updated timestamp displays

## Understanding Schema Information

### Column Types

| Type | Description | Example |
|------|-------------|---------|
| `integer` | Whole numbers | `id: integer` |
| `varchar(n)` | Text with max length | `name: varchar(255)` |
| `text` | Unlimited text | `description: text` |
| `timestamp` | Date and time | `created_at: timestamp` |
| `boolean` | True/false | `is_active: boolean` |
| `decimal(p,s)` | Precise decimal | `price: decimal(10,2)` |
| `json` | JSON data | `metadata: json` |
| `uuid` | Unique identifier | `id: uuid` |

### Constraints

- **PK** (Primary Key): Unique identifier for rows
- **FK** (Foreign Key): References another table
- **NOT NULL**: Column must have a value
- **UNIQUE**: All values must be different
- **DEFAULT**: Default value if not specified

## Schema Browser vs SQL Editor

### When to Use Schema Browser

- **Exploring new databases**
- **Understanding table structure**
- **Finding column names**
- **Checking data types**
- **Planning queries**

### When to Use SQL Editor

- **Executing queries**
- **Writing complex JOINs**
- **Data analysis**
- **Write operations**

## Integration with SQL Editor

The Schema Browser integrates with the SQL Editor for autocomplete:

1. **View Schema First**
   - Browse tables and columns
   - Understand structure

2. **Write Query with Autocomplete**
   - Start typing query
   - Autocomplete suggests tables
   - Autocomplete suggests columns

3. **Example Workflow**
   ```
   -- After viewing "users" table in Schema Browser:
   SELECT id, email, created_at
   FROM users
   WHERE is_active = true;
   ```

## Schema Synchronization Details

### Polling Behavior

- **Interval**: Every 60 seconds
- **Scope**: Only current data source
- **Performance**: Minimal impact (cached for 5 minutes)

### Background Worker

- **Interval**: Every 5 minutes
- **Scope**: All data sources
- **Priority**: Low (doesn't block queries)

### Manual Sync

- **Trigger**: User clicks "Sync Now"
- **Scope**: Current data source only
- **Priority**: High (immediate execution)

## Troubleshooting

### Schema Not Loading

**Possible causes:**
- Data source connection failed
- Insufficient permissions
- Network issues

**Solutions:**
- Check data source status
- Verify permissions
- Test connection in Admin Panel

### Stale Schema Information

**Possible causes:**
- Background worker not running
- Sync interval too long
- Schema changed externally

**Solutions:**
- Click "Sync Now" for immediate refresh
- Check worker is running
- Reduce sync interval in config

### Tables Not Showing

**Possible causes:**
- No permissions on tables
- Schema not synced yet
- Search filter active

**Solutions:**
- Check data source permissions
- Click "Sync Now"
- Clear search filter

## Best Practices

### Before Writing Queries

1. **Explore Schema First**
   - Browse tables you'll query
   - Check column names and types
   - Identify relationships

2. **Understand Constraints**
   - Note primary keys
   - Check foreign key relationships
   - Identify required columns

3. **Plan Your Query**
   - Use schema info to write accurate queries
   - Avoid guessing column names
   - Use correct data types

### For Database Administrators

1. **Organize Tables**
   - Use naming conventions
   - Group related tables
   - Document schema changes

2. **Set Appropriate Sync Intervals**
   - Frequent changes: 1-2 minutes
   - Stable schemas: 5-10 minutes
   - Consider performance impact

## Keyboard Shortcuts

- **Cmd/Ctrl + F**: Focus search bar
- **Esc**: Clear search
- **↑/↓**: Navigate table list
- **Enter**: Expand table columns

## Related Features

- **[SQL Editor Autocomplete](query-features.md)** - Use schema info for intelligent suggestions
- **[Query Features](query-features.md)** - EXPLAIN and Dry Run
- **[Admin Panel](admin-panel.md)** - Manage data sources

---

**Need help?** Check the [main documentation](../README.md) or [contact support](../README.md).
