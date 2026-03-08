package dto

// StartTransactionRequest represents a request to start a transaction for an approval
type StartTransactionRequest struct {
	ApprovalID string `json:"approval_id" binding:"required"`
	AuditMode  string `json:"audit_mode"` // full, sample, count_only (default: full)
}

// TransactionResponse represents the response when starting a transaction
type TransactionResponse struct {
	TransactionID string             `json:"transaction_id"`
	ApprovalID    string             `json:"approval_id"`
	Status        string             `json:"status"`
	QueryText     string             `json:"query_text"`
	DataSourceID  string             `json:"data_source_id"`
	StartedBy     string             `json:"started_by"`
	StartedAt     string             `json:"started_at"`
	Preview       TransactionPreview `json:"preview"`
}

// TransactionPreview represents the preview data from the transaction
type TransactionPreview struct {
	RowCount      int                      `json:"row_count"`
	EstimatedRows int                      `json:"estimated_rows"`
	Columns       []ColumnInfo             `json:"columns"`
	Data          []map[string]interface{} `json:"data"`
	Caution       bool                     `json:"caution"`
	CautionMsg    string                   `json:"caution_message,omitempty"`
	AuditMode     string                   `json:"audit_mode"`
}

// CommitTransactionRequest represents a request to commit a transaction
type CommitTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// CommitTransactionResponse represents the response when committing a transaction
type CommitTransactionResponse struct {
	TransactionID string                   `json:"transaction_id"`
	Status        string                   `json:"status"`
	Message       string                   `json:"message"`
	ApprovalID    string                   `json:"approval_id"`
	AffectedRows  int                      `json:"affected_rows"`
	AuditMode     string                   `json:"audit_mode"`
	BeforeData    []map[string]interface{} `json:"before_data,omitempty"`
	AfterData     []map[string]interface{} `json:"after_data,omitempty"`
}

// RollbackTransactionRequest represents a request to rollback a transaction
type RollbackTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// RollbackTransactionResponse represents the response when rolling back a transaction
type RollbackTransactionResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	ApprovalID    string `json:"approval_id"`
}

// TransactionStatusResponse represents the response for getting transaction status
type TransactionStatusResponse struct {
	TransactionID string                   `json:"transaction_id"`
	ApprovalID    string                   `json:"approval_id"`
	Status        string                   `json:"status"`
	QueryText     string                   `json:"query_text"`
	DataSourceID  string                   `json:"data_source_id"`
	StartedBy     string                   `json:"started_by"`
	StartedAt     string                   `json:"started_at"`
	CompletedAt   *string                  `json:"completed_at,omitempty"`
	ErrorMessage  string                   `json:"error_message,omitempty"`
	AffectedRows  int                      `json:"affected_rows"`
	EstimatedRows int                      `json:"estimated_rows"`
	AuditMode     string                   `json:"audit_mode"`
	BeforeData    []map[string]interface{} `json:"before_data,omitempty"`
	AfterData     []map[string]interface{} `json:"after_data,omitempty"`
}

// ValidateQueryRequest represents a request to validate a SQL query
type ValidateQueryRequest struct {
	QueryText    string `json:"query_text" binding:"required"`
	DataSourceID string `json:"data_source_id" binding:"required"` // Required for schema validation
}

// ValidateQueryResponse represents the response from query validation
type ValidateQueryResponse struct {
	Valid         bool   `json:"valid"`
	Error         string `json:"error,omitempty"`
	OperationType string `json:"operation_type"`
}
