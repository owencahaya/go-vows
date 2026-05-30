package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard JSON envelope returned by all endpoints.
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success writes a 2xx JSON response.
func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error writes a non-2xx JSON response. The errCode is a short machine-readable
// code (e.g. "invitation_not_found"), message is human readable.
func Error(c *gin.Context, code int, errCode, message string) {
	c.JSON(code, APIResponse{
		Status:  "error",
		Error:   errCode,
		Message: message,
	})
}

// BadRequest is a convenience helper for validation errors.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "bad_request", message)
}

// InternalError is a convenience helper for unexpected errors.
func InternalError(c *gin.Context, err error) {
	Error(c, http.StatusInternalServerError, "internal_error", err.Error())
}
