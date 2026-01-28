package dto

// ApprovalRequestResponse represents an approval request response
type ApprovalRequestResponse struct {
	ID            string `json:"id"`
	QueryID       *string `json:"query_id,omitempty"`
	OperationType string `json:"operation_type"`
	QueryText     string `json:"query_text"`
	DataSourceID  string `json:"data_source_id"`
	DataSourceName string `json:"data_source_name"`
	RequestedBy   string `json:"requested_by"`
	RequesterName string `json:"requester_name"`
	Status        string `json:"status"`
	RejectionReason string `json:"rejection_reason,omitempty"`
	CreatedAt     string `json:"created_at"`
	Reviews       []ApprovalReviewResponse `json:"reviews,omitempty"`
}

// ApprovalReviewResponse represents an approval review response
type ApprovalReviewResponse struct {
	ID         string `json:"id"`
	ReviewedBy string `json:"reviewed_by"`
	ReviewerName string `json:"reviewer_name"`
	Status     string `json:"status"`
	Comments   string `json:"comments"`
	ReviewedAt string `json:"reviewed_at"`
}

// ReviewApprovalRequest represents a review approval request
type ReviewApprovalRequest struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected"`
	Comments string `json:"comments"`
}

// ApprovalCommentRequest represents a comment creation request
type ApprovalCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1,max=5000"`
}

// ApprovalCommentResponse represents a comment response
type ApprovalCommentResponse struct {
	ID                string `json:"id"`
	ApprovalRequestID string `json:"approval_request_id"`
	UserID            string `json:"user_id"`
	Username          string `json:"username"`
	FullName          string `json:"full_name"`
	Comment           string `json:"comment"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// ApprovalCommentsListResponse represents a paginated list of comments
type ApprovalCommentsListResponse struct {
	Comments []ApprovalCommentResponse `json:"comments"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PerPage  int                       `json:"per_page"`
}
