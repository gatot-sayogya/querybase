package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error constructors
func BadRequest(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

func Conflict(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: message,
		Err:     err,
	}
}

func InternalError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

func ServiceUnavailable(message string) *AppError {
	return &AppError{
		Code:    http.StatusServiceUnavailable,
		Message: message,
	}
}

// Specific domain errors
func InvalidCredentials() *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: "Invalid username or password",
	}
}

func InvalidToken() *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: "Invalid or expired token",
	}
}

func InsufficientPermissions() *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: "Insufficient permissions to perform this action",
	}
}

func DataSourceNotFound(id string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("Data source %s not found", id),
	}
}

func QueryNotFound(id string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("Query %s not found", id),
	}
}

func UserNotFound(id string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("User %s not found", id),
	}
}

func GroupNotFound(id string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("Group %s not found", id),
	}
}

func ApprovalNotFound(id string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("Approval request %s not found", id),
	}
}

func InvalidSQL(details string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid SQL syntax",
		Details: details,
		Err:     err,
	}
}

func DataSourceConnectionFailed(err error) *AppError {
	return &AppError{
		Code:    http.StatusServiceUnavailable,
		Message: "Failed to connect to data source",
		Err:     err,
	}
}

func QueryExecutionFailed(err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Query execution failed",
		Err:     err,
	}
}

func ValidationError(field, message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Validation failed for field '%s': %s", field, message),
	}
}
