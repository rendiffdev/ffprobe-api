package errors

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code,omitempty"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// Common error codes
const (
	CodeValidationError    = "VALIDATION_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeUnauthorized      = "UNAUTHORIZED" 
	CodeForbidden         = "FORBIDDEN"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeBadRequest        = "BAD_REQUEST"
	CodeConflict          = "CONFLICT"
	CodeTooManyRequests   = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, code, message, details string) {
	requestID := ""
	if rid, exists := c.Get("request_id"); exists {
		if ridStr, ok := rid.(string); ok {
			requestID = ridStr
		}
	}

	response := ErrorResponse{
		Error:     message,
		Code:      code,
		Details:   details,
		Timestamp: time.Now(),
		RequestID: requestID,
	}

	c.JSON(statusCode, response)
}

// Common error response helpers
func BadRequest(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusBadRequest, CodeBadRequest, message, details)
}

func ValidationError(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusBadRequest, CodeValidationError, message, details)
}

func NotFound(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusNotFound, CodeNotFound, message, details)
}

func Unauthorized(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusUnauthorized, CodeUnauthorized, message, details)
}

func Forbidden(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusForbidden, CodeForbidden, message, details)
}

func InternalError(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusInternalServerError, CodeInternalError, message, details)
}

func Conflict(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusConflict, CodeConflict, message, details)
}

func TooManyRequests(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusTooManyRequests, CodeTooManyRequests, message, details)
}

func ServiceUnavailable(c *gin.Context, message, details string) {
	RespondWithError(c, http.StatusServiceUnavailable, CodeServiceUnavailable, message, details)
}