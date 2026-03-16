package dto

// ExecuteQueryRequest represents a query execution request
type ExecuteQueryRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

// ExecuteQueryResponse represents a query execution response
type ExecuteQueryResponse struct {
	QueryID          string                   `json:"query_id"`
	Status           string                   `json:"status"`
	RowCount         *int                     `json:"row_count"`
	ExecutionTime    *int                     `json:"execution_time_ms"`
	Data             []map[string]interface{} `json:"data"`
	Columns          []ColumnInfo             `json:"columns"`
	ErrorMessage     string                   `json:"error_message,omitempty"`
	RequiresApproval bool                     `json:"requires_approval"`
	ApprovalID       string                   `json:"approval_id,omitempty"`
	Validation       *ValidationResult        `json:"validation,omitempty"`
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// SaveQueryRequest represents a save query request
type SaveQueryRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

// QueryListResponse represents a list of queries
type QueryListResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DataSourceID   string `json:"data_source_id"`
	DataSourceName string `json:"data_source_name"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	OperationType  string `json:"operation_type"`
	Status         string `json:"status"`
	RowCount       *int   `json:"row_count"`
	CreatedAt      string `json:"created_at"`
}

// ExplainQueryRequest represents an EXPLAIN query request
type ExplainQueryRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
	Analyze      bool   `json:"analyze"` // If true, use EXPLAIN ANALYZE
}

// DryRunRequest represents a dry run request for DELETE queries
type DryRunRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	TotalPages int  `json:"total_pages"`
	TotalRows  int  `json:"total_rows"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// PaginatedResultDTO represents paginated query results
type PaginatedResultDTO struct {
	QueryID       string                   `json:"query_id"`
	RowCount      int                      `json:"row_count"`
	Columns       []ColumnInfo             `json:"columns"`
	Data          []map[string]interface{} `json:"data"`
	Metadata      PaginationMeta           `json:"metadata"`
	SortColumn    string                   `json:"sort_column,omitempty"`
	SortDirection string                   `json:"sort_direction,omitempty"`
}

// ExportFormat represents the export format type
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
)

// ExportQueryRequest represents a query export request
type ExportQueryRequest struct {
	QueryID string       `json:"query_id" binding:"required"`
	Format  ExportFormat `json:"format" binding:"required,oneof=csv json"`
}

// ExportQueryResponse represents a query export response
type ExportQueryResponse struct {
	QueryID     string `json:"query_id"`
	Format      string `json:"format"`
	RowCount    int    `json:"row_count"`
	FileSize    int64  `json:"file_size"`
	DownloadURL string `json:"download_url"`
}

// PreviewWriteQueryRequest represents a request to preview affected rows for a write query
type PreviewWriteQueryRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
}

// PreviewWriteQueryResponse represents the preview result for a write query
type PreviewWriteQueryResponse struct {
	TotalAffected int                      `json:"total_affected"`
	PreviewRows   []map[string]interface{} `json:"preview_rows"`
	Columns       []string                 `json:"columns"`
	PreviewLimit  int                      `json:"preview_limit"`
	SelectQuery   string                   `json:"select_query"`
	OperationType string                   `json:"operation_type"`
}

// ValidationResult represents the result of validating a write query before creating approval
type ValidationResult struct {
	Valid        bool                     `json:"valid"`
	Status       string                   `json:"status"` // "ok", "no_match", "error"
	Message      string                   `json:"message"`
	AffectedRows int                      `json:"affected_rows"`
	PreviewRows  []map[string]interface{} `json:"preview_rows,omitempty"`
	Columns      []string                 `json:"columns,omitempty"`
	Suggestion   string                   `json:"suggestion,omitempty"`
}

// ValidateWriteQueryResponse represents the response for write query validation
type ValidateWriteQueryResponse struct {
	QueryID    string           `json:"query_id"`
	Status     string           `json:"status"` // "validated", "no_match", "error"
	Validation ValidationResult `json:"validation"`
}
