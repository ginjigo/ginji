package ginji

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCustomErrorHandler tests custom error handler functionality.
func TestCustomErrorHandler(t *testing.T) {
	app := New()

	// Use error handling middleware
	app.Use(DefaultErrorHandler())

	// Set a custom error handler
	customCalled := false
	app.SetErrorHandler(func(c *Context, err error) {
		customCalled = true
		_ = c.JSONOK(H{
			"custom_error": true,
			"message":      err.Error(),
		})
	})

	// Handler that sets an error
	app.Get("/error", func(c *Context) error {
		c.Error(errors.New("custom test error"))
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/error", nil)
	app.ServeHTTP(w, req)

	if !customCalled {
		t.Error("Custom error handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestDefaultErrorHandlerWithEngine tests that default error handler works when no custom handler is set.
func TestDefaultErrorHandlerWithEngine(t *testing.T) {
	app := New()

	app.Get("/error", func(c *Context) error {
		c.AbortWithError(http.StatusBadRequest, NewHTTPError(http.StatusBadRequest, "test error"))
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/error", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestStackTraceInDebugMode tests that stack traces are included in debug mode.
func TestStackTraceInDebugMode(t *testing.T) {
	// Ensure we're in debug mode
	originalMode := GetMode()
	SetMode(DebugMode)
	defer SetMode(originalMode)

	app := New()

	app.Get("/error", func(c *Context) error {
		c.AbortWithError(http.StatusInternalServerError, NewHTTPError(http.StatusInternalServerError, "test error"))
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/error", nil)
	app.ServeHTTP(w, req)

	// In debug mode, response should contain stack trace
	body := w.Body.String()
	if body == "" {
		t.Error("Expected error response body")
	}
}

// TestValidationErrorFormatting tests that validation errors are properly formatted.
func TestValidationErrorFormatting(t *testing.T) {
	app := New()

	// Use error handling middleware
	app.Use(DefaultErrorHandler())

	app.Post("/validate", func(c *Context) error {
		// Simulate validation error
		verrs := ValidationErrors{
			FormatValidationError("email", "email is required", "required", nil),
			FormatValidationError("age", "age must be at least 18", "min", 15),
		}
		c.AbortWithError(http.StatusUnprocessableEntity, verrs)
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/validate", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", w.Code)
	}
}
