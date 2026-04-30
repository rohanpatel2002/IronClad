package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a standardized application error.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Error codes
const (
	ErrInvalidRequest   = "invalid_request"
	ErrUnauthorized     = "unauthorized"
	ErrForbidden       = "forbidden"
	ErrNotFound        = "not_found"
	ErrInternal        = "internal_error"
	ErrRateLimited     = "rate_limited"
	ErrCircuitOpen     = "circuit_breaker_open"
	ErrExternalService = "external_service_error"
)

// New creates a new AppError.
func New(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// Respond sends a standardized JSON error response.
func Respond(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		c.AbortWithStatusJSON(appErr.Status, appErr)
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, AppError{
		Code:    ErrInternal,
		Message: "An unexpected error occurred",
	})
}
