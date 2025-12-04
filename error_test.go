package ginji

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHTTPError(t *testing.T) {
	// Test with default message
	err := NewHTTPError(http.StatusNotFound)
	if err.Code != http.StatusNotFound {
		t.Errorf("Expected code %d, got %d", http.StatusNotFound, err.Code)
	}
	if err.Message != "Not Found" {
		t.Errorf("Expected message 'Not Found', got '%s'", err.Message)
	}

	// Test with custom message
	err = NewHTTPError(http.StatusBadRequest, "Invalid input")
	if err.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got '%s'", err.Message)
	}
}

func TestHTTPErrorWithDetails(t *testing.T) {
	err := NewHTTPError(http.StatusUnprocessableEntity).WithDetails(map[string]string{
		"field": "email",
		"issue": "invalid format",
	})

	if err.Details == nil {
		t.Error("Expected details to be set")
	}
}

func TestHTTPErrorError(t *testing.T) {
	err := NewHTTPError(http.StatusNotFound)
	expected := "HTTP 404: Not Found"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}

	// With details
	err = NewHTTPError(http.StatusBadRequest).WithDetails("extra info")
	if err.Error() != "HTTP 400: Bad Request (details: extra info)" {
		t.Errorf("Unexpected error string: %s", err.Error())
	}
}

func TestValidationErrors(t *testing.T) {
	errs := ValidationErrors{
		{Field: "email", Message: "invalid format"},
		{Field: "age", Message: "must be positive"},
	}

	errStr := errs.Error()
	if errStr != "validation failed on field 'email': invalid format" {
		t.Errorf("Unexpected error string: %s", errStr)
	}

	// Empty errors
	var emptyErrs ValidationErrors
	if emptyErrs.Error() != "validation failed" {
		t.Errorf("Expected 'validation failed', got '%s'", emptyErrs.Error())
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	// Test with HTTPError
	app := New()
	app.Use(DefaultErrorHandler())
	app.Get("/error", func(c *Context) {
		c.Error(NewHTTPError(http.StatusBadRequest, "Test error"))
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error != "Test error" {
		t.Errorf("Expected error 'Test error', got '%s'", response.Error)
	}
}

func TestContextAbortWithError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := NewContext(w, req)

	c.AbortWithError(http.StatusUnauthorized, errors.New("not authorized"))

	if !c.IsAborted() {
		t.Error("Expected context to be aborted")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestContextAbortWithStatusJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := NewContext(w, req)

	c.AbortWithStatusJSON(http.StatusForbidden, H{"message": "forbidden"})

	if !c.IsAborted() {
		t.Error("Expected context to be aborted")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestContextTypedGetters(t *testing.T) {
	c := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	c.Set("name", "John")
	c.Set("age", 30)
	c.Set("active", true)

	if c.GetString("name") != "John" {
		t.Errorf("Expected 'John', got '%s'", c.GetString("name"))
	}

	if c.GetInt("age") != 30 {
		t.Errorf("Expected 30, got %d", c.GetInt("age"))
	}

	if !c.GetBool("active") {
		t.Error("Expected true, got false")
	}

	// Test non-existent keys
	if c.GetString("nonexistent") != "" {
		t.Error("Expected empty string for non-existent key")
	}

	if c.GetInt("nonexistent") != 0 {
		t.Error("Expected 0 for non-existent key")
	}

	if c.GetBool("nonexistent") {
		t.Error("Expected false for non-existent key")
	}
}

func TestContextMustGet(t *testing.T) {
	c := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	c.Set("key", "value")
	val := c.MustGet("key")
	if val != "value" {
		t.Errorf("Expected 'value', got '%v'", val)
	}

	// Test panic on non-existent key
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-existent key")
		}
	}()
	c.MustGet("nonexistent")
}

func TestCommonHTTPErrors(t *testing.T) {
	tests := []struct {
		err  *HTTPError
		code int
		msg  string
	}{
		{ErrBadRequest, http.StatusBadRequest, "Bad Request"},
		{ErrUnauthorized, http.StatusUnauthorized, "Unauthorized"},
		{ErrForbidden, http.StatusForbidden, "Forbidden"},
		{ErrNotFound, http.StatusNotFound, "Not Found"},
		{ErrInternalServerError, http.StatusInternalServerError, "Internal Server Error"},
	}

	for _, tt := range tests {
		if tt.err.Code != tt.code {
			t.Errorf("Expected code %d, got %d", tt.code, tt.err.Code)
		}
		if tt.err.Message != tt.msg {
			t.Errorf("Expected message '%s', got '%s'", tt.msg, tt.err.Message)
		}
	}
}

func TestModeConfiguration(t *testing.T) {
	// Save original mode
	originalMode := GetMode()
	defer SetMode(originalMode)

	SetMode(ReleaseMode)
	if GetMode() != ReleaseMode {
		t.Errorf("Expected ReleaseMode, got %v", GetMode())
	}

	SetMode(DebugMode)
	if GetMode() != DebugMode {
		t.Errorf("Expected DebugMode, got %v", GetMode())
	}

	SetMode(TestMode)
	if GetMode() != TestMode {
		t.Errorf("Expected TestMode, got %v", GetMode())
	}
}

func TestFormatValidationError(t *testing.T) {
	ve := FormatValidationError("email", "invalid format", "email", "not-an-email")

	if ve.Field != "email" {
		t.Errorf("Expected field 'email', got '%s'", ve.Field)
	}
	if ve.Message != "invalid format" {
		t.Errorf("Expected message 'invalid format', got '%s'", ve.Message)
	}
	if ve.Tag != "email" {
		t.Errorf("Expected tag 'email', got '%s'", ve.Tag)
	}
}
