package ginji

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

// HTTPError represents an HTTP error with status code, message, and details.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	stack   string // internal stack trace
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("HTTP %d: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// WithDetails adds details to the error.
func (e *HTTPError) WithDetails(details any) *HTTPError {
	e.Details = details
	return e
}

// StackTrace returns the stack trace if available.
func (e *HTTPError) StackTrace() string {
	return e.stack
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(code int, message ...string) *HTTPError {
	err := &HTTPError{
		Code:    code,
		Message: http.StatusText(code),
	}
	if len(message) > 0 {
		err.Message = message[0]
	}
	// Only capture stack traces in debug mode
	if mode == DebugMode {
		err.stack = captureStackTrace()
	}
	return err
}

// Common HTTP errors
var (
	ErrBadRequest          = NewHTTPError(http.StatusBadRequest)
	ErrUnauthorized        = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden           = NewHTTPError(http.StatusForbidden)
	ErrNotFound            = NewHTTPError(http.StatusNotFound)
	ErrMethodNotAllowed    = NewHTTPError(http.StatusMethodNotAllowed)
	ErrConflict            = NewHTTPError(http.StatusConflict)
	ErrUnprocessableEntity = NewHTTPError(http.StatusUnprocessableEntity)
	ErrTooManyRequests     = NewHTTPError(http.StatusTooManyRequests)
	ErrInternalServerError = NewHTTPError(http.StatusInternalServerError)
	ErrNotImplemented      = NewHTTPError(http.StatusNotImplemented)
	ErrServiceUnavailable  = NewHTTPError(http.StatusServiceUnavailable)
)

// ValidationError represents a validation error with field-level details.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   any    `json:"value,omitempty"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed on field '%s': %s", ve[0].Field, ve[0].Message)
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error   string           `json:"error"`
	Code    int              `json:"code"`
	Details any              `json:"details,omitempty"`
	Stack   string           `json:"stack,omitempty"`
	Errors  ValidationErrors `json:"errors,omitempty"`
}

// DefaultErrorHandler is the default error handler middleware.
func DefaultErrorHandler() Middleware {
	return func(c *Context) error {
		defer func() {
			if c.error != nil {
				handleError(c, c.error)
			}
		}()
		return c.Next()
	}
}

// handleError handles the error and sends an appropriate response.
// It uses the custom error handler if set, otherwise uses the default.
func handleError(c *Context, err error) {
	// Use custom error handler if set
	if c.engine != nil && c.engine.errorHandler != nil {
		c.engine.errorHandler(c, err)
		return
	}
	// Otherwise use default
	defaultErrorHandler(c, err)
}

// defaultErrorHandler is the default error handling logic.
func defaultErrorHandler(c *Context, err error) {
	// If already written, don't write again
	if c.written {
		return
	}

	var httpErr *HTTPError
	var validationErrs ValidationErrors

	// Check if it's an HTTPError
	if he, ok := err.(*HTTPError); ok {
		httpErr = he
	} else if ve, ok := err.(ValidationErrors); ok {
		// Validation error
		validationErrs = ve
		httpErr = NewHTTPError(http.StatusUnprocessableEntity, "Validation failed")
	} else {
		// Generic error - treat as 500
		httpErr = NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Build error response
	response := ErrorResponse{
		Error:   httpErr.Message,
		Code:    httpErr.Code,
		Details: httpErr.Details,
	}

	// Add validation errors if present
	if validationErrs != nil {
		response.Errors = validationErrs
	}

	// Only add stack trace in debug mode, never in production
	if mode == DebugMode && httpErr.stack != "" {
		response.Stack = httpErr.stack
	}

	// Send JSON response
	_ = c.JSON(httpErr.Code, response)
}

// captureStackTrace captures the current stack trace.
func captureStackTrace() string {
	const maxStackSize = 50
	pcs := make([]uintptr, maxStackSize)
	n := runtime.Callers(3, pcs) // skip first 3 frames

	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var trace string

	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("\n  %s:%d %s", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}

	return trace
}

// FormatValidationError formats a validation error into ValidationError.
func FormatValidationError(field, message, tag string, value any) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Tag:     tag,
		Value:   value,
	}
}

// MarshalJSON customizes JSON marshaling for HTTPError.
func (e *HTTPError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
	}{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	})
}
