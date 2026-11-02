package common

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AppError is a domain error carrying the HTTP status to return, the Go analogue of the Kotlin
// ApiException. Services return it; handlers translate it via Fail.
type AppError struct {
	Status int
	Msg    string
}

func (e *AppError) Error() string { return e.Msg }

func NewAppError(status int, msg string) *AppError { return &AppError{Status: status, Msg: msg} }

// ApiError is the JSON body returned for controller-level errors — identical shape to the Kotlin
// GlobalExceptionHandler's ApiError: {timestamp, status, error, message, path}.
type ApiError struct {
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
}

// AbortError writes the full ApiError body and aborts the request.
func AbortError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, ApiError{
		Timestamp: time.Now().UTC(),
		Status:    status,
		Error:     http.StatusText(status),
		Message:   message,
		Path:      c.Request.URL.Path,
	})
}

// Fail maps an error to a response: *AppError uses its status/message, anything else is a 500.
func Fail(c *gin.Context, err error) {
	if ae, ok := err.(*AppError); ok {
		AbortError(c, ae.Status, ae.Msg)
		return
	}
	AbortError(c, http.StatusInternalServerError, "Internal server error")
}

// The security layer (auth + role middleware) uses a terser body, matching the Kotlin
// SecurityConfig entryPoint/accessDeniedHandler: {status, error, message} with no timestamp/path.

func AbortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"status":  401,
		"error":   "Unauthorized",
		"message": "Authentication required",
	})
}

func AbortForbidden(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status":  403,
		"error":   "Forbidden",
		"message": "You do not have access to this resource",
	})
}
