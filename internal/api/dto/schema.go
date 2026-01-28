package dto

// DatabaseSchemaResponse represents the complete database schema
type DatabaseSchemaResponse struct {
	DataSourceID   string       `json:"data_source_id"`
	DataSourceName string       `json:"data_source_name"`
	DatabaseType   string       `json:"database_type"`
	DatabaseName   string       `json:"database_name"`
	Tables         []TableInfo  `json:"tables"`
	Schemas        []string     `json:"schemas,omitempty"`
}

// TableInfo represents a database table
type TableInfo struct {
	TableName string        `json:"table_name"`
	Schema    string        `json:"schema"`
	Columns   []ColumnInfo `json:"columns"`
	Indexes   []IndexInfo   `json:"indexes,omitempty"`
}

// IndexInfo represents an index on a table
type IndexInfo struct {
	IndexName string   `json:"index_name"`
	Columns   []string `json:"columns"`
	IsUnique  bool     `json:"is_unique"`
	IsPrimary bool     `json:"is_primary"`
}
