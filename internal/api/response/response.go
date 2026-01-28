package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/yourorg/querybase/internal/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SendError sends an error response with appropriate status code
func SendError(c *gin.Context, err error) {
	// Check if it's an AppError
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.Code, ErrorResponse{
			Error:   appErr.Message,
			Details: appErr.Details,
			Code:    appErr.Code,
		})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "Internal server error",
		Code:  http.StatusInternalServerError,
	})
}

// SendBadRequest sends a 400 Bad Request error
func SendBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: message,
		Code:  http.StatusBadRequest,
	})
}

// SendUnauthorized sends a 401 Unauthorized error
func SendUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error: message,
		Code:  http.StatusUnauthorized,
	})
}

// SendForbidden sends a 403 Forbidden error
func SendForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error: message,
		Code:  http.StatusForbidden,
	})
}

// SendNotFound sends a 404 Not Found error
func SendNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: message,
		Code:  http.StatusNotFound,
	})
}

// SendConflict sends a 409 Conflict error
func SendConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Error: message,
		Code:  http.StatusConflict,
	})
}

// SendInternalError sends a 500 Internal Server Error
func SendInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: message,
		Code:  http.StatusInternalServerError,
	})
}

// SendSuccess sends a successful response with data
func SendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// SendSuccessMessage sends a successful response with a message
func SendSuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
	})
}

// SendCreated sends a 201 Created response with data
func SendCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// SendNoContent sends a 204 No Content response
func SendNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
