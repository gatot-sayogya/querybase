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
	QueryID      string        `json:"query_id"`
	Status       string        `json:"status"`
	RowCount     *int          `json:"row_count"`
	ExecutionTime *int         `json:"execution_time_ms"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	Columns      []ColumnInfo  `json:"columns"`
	ErrorMessage string        `json:"error_message,omitempty"`
	RequiresApproval bool       `json:"requires_approval"`
	ApprovalID   string        `json:"approval_id,omitempty"`
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
	ID            string `json:"id"`
	Name          string `json:"name"`
	DataSourceID  string `json:"data_source_id"`
	DataSourceName string `json:"data_source_name"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	OperationType string `json:"operation_type"`
	Status        string `json:"status"`
	RowCount      *int   `json:"row_count"`
	CreatedAt     string `json:"created_at"`
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

