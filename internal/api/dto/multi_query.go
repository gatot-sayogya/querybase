package dto

// MultiQueryPreviewRequest represents a preview request for multiple queries
type MultiQueryPreviewRequest struct {
	DataSourceID string   `json:"data_source_id" binding:"required"`
	QueryTexts   []string `json:"query_texts" binding:"required,min=1,max=50"`
}

// StatementPreview represents a preview for a single statement
type StatementPreview struct {
	Sequence      int                      `json:"sequence"`
	QueryText     string                   `json:"query_text"`
	OperationType string                   `json:"operation_type"`
	EstimatedRows int                      `json:"estimated_rows"`
	PreviewRows   []map[string]interface{} `json:"preview_rows,omitempty"`
	Columns       []ColumnInfo             `json:"columns,omitempty"`
	Error         string                   `json:"error,omitempty"`
}

// MultiQueryPreviewResponse represents the preview response
type MultiQueryPreviewResponse struct {
	StatementCount     int                `json:"statement_count"`
	TotalEstimatedRows int                `json:"total_estimated_rows"`
	Statements         []StatementPreview `json:"statements"`
	RequiresApproval   bool               `json:"requires_approval"`
}

// MultiQueryExecuteRequest represents an execution request
type MultiQueryExecuteRequest struct {
	DataSourceID string   `json:"data_source_id" binding:"required"`
	QueryTexts   []string `json:"query_texts" binding:"required,min=1,max=50"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
}

// StatementResult represents the result of a single statement
type StatementResult struct {
	Sequence        int                      `json:"sequence"`
	QueryText       string                   `json:"query_text"`
	OperationType   string                   `json:"operation_type"`
	Status          string                   `json:"status"`
	AffectedRows    int                      `json:"affected_rows"`
	RowCount        int                      `json:"row_count"`
	Columns         []ColumnInfo             `json:"columns,omitempty"`
	Data            []map[string]interface{} `json:"data,omitempty"`
	ErrorMessage    string                   `json:"error_message,omitempty"`
	ExecutionTimeMs int                      `json:"execution_time_ms"`
}

// MultiQueryResponse represents the execution response
type MultiQueryResponse struct {
	QueryID           string            `json:"query_id"`
	TransactionID     string            `json:"transaction_id"`
	Status            string            `json:"status"`
	IsMultiQuery      bool              `json:"is_multi_query"`
	StatementCount    int               `json:"statement_count"`
	TotalAffectedRows int               `json:"total_affected_rows"`
	ExecutionTimeMs   int               `json:"execution_time_ms"`
	Statements        []StatementResult `json:"statements"`
	ErrorMessage      string            `json:"error_message,omitempty"`
	RequiresApproval  bool              `json:"requires_approval"`
	ApprovalID        string            `json:"approval_id,omitempty"`
}

// GetMultiQueryStatementsRequest represents a request to get statement details
type GetMultiQueryStatementsRequest struct {
	TransactionID string `uri:"id" binding:"required"`
}

// CommitMultiQueryRequest represents a commit request
type CommitMultiQueryRequest struct {
	TransactionID string `uri:"id" binding:"required"`
}

// RollbackMultiQueryRequest represents a rollback request
type RollbackMultiQueryRequest struct {
	TransactionID string `uri:"id" binding:"required"`
}
