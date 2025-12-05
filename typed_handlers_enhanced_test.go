package ginji

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBindingErrorMessages tests that the enhanced error messages are informative.
func TestBindingErrorMessages(t *testing.T) {
	app := New()

	type Request struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	app.Post("/test", TypedHandlerFunc(func(c *Context, req Request) (string, error) {
		return "OK", nil
	}))

	tests := []struct {
		name           string
		body           string
		contentType    string
		expectContains string
	}{
		{
			name:           "invalid JSON",
			body:           `{invalid json}`,
			contentType:    "application/json",
			expectContains: "JSON body",
		},
		{
			name:           "type mismatch in JSON",
			body:           `{"name": 123, "age": "not a number"}`,
			contentType:    "application/json",
			expectContains: "Failed to bind request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			if rec.Code != 400 {
				t.Errorf("Expected status 400, got %d", rec.Code)
			}

			body := rec.Body.String()
			if tt.expectContains != "" {
				if !strings.Contains(body, tt.expectContains) {
					t.Errorf("Expected error message to contain %q, got: %s", tt.expectContains, body)
				}
			}
		})
	}
}

// TestTypeCaching tests that type caching works correctly.
func TestTypeCaching(t *testing.T) {
	app := New()

	type Request struct {
		Message string `json:"message"`
	}

	// Register multiple handlers with the same types
	handler := func(c *Context, req Request) (Request, error) {
		return Request{Message: "Echo: " + req.Message}, nil
	}

	app.Post("/test1", TypedHandlerFunc(handler))
	app.Post("/test2", TypedHandlerFunc(handler))
	app.Post("/test3", TypedHandlerFunc(handler))

	// Test that all handlers work correctly (type caching shouldn't break anything)
	testData := Request{Message: "hello"}
	body, _ := json.Marshal(testData)

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("POST", "/test"+string(rune(i+'0')), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("Handler %d: Expected status 200, got %d", i, rec.Code)
		}
	}
}

// TestEmptyRequestResponse tests EmptyRequest and response handling.
func TestEmptyRequestResponse(t *testing.T) {
	app := New()

	// Handler with no input, has output
	app.Get("/no-input", TypedHandlerFunc(func(c *Context, _ EmptyRequest) (map[string]string, error) {
		return map[string]string{"message": "hello"}, nil
	}))

	// Handler with input, no output
	app.Delete("/no-output", TypedHandlerFunc(func(c *Context, req struct{ ID string }) (EmptyRequest, error) {
		return EmptyRequest{}, nil
	}))

	// Test no input
	req := httptest.NewRequest("GET", "/no-input", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "hello" {
		t.Errorf("Expected message 'hello', got %q", response["message"])
	}

	// Test no output
	req = httptest.NewRequest("DELETE", "/no-output?ID=123", nil)
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != 204 {
		t.Errorf("Expected status 204, got %d", rec.Code)
	}

	if rec.Body.Len() > 0 {
		t.Errorf("Expected empty body, got: %s", rec.Body.String())
	}
}

// TestBindingErrorType tests the BindingError type.
func TestBindingErrorType(t *testing.T) {
	err := &BindingError{
		Source:      "JSON body",
		ContentType: "application/json",
		Cause:       http.ErrMissingFile,
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "JSON body") {
		t.Errorf("Error message should contain source: %s", errMsg)
	}
	if !strings.Contains(errMsg, "application/json") {
		t.Errorf("Error message should contain content type: %s", errMsg)
	}

	// Test Unwrap
	if err.Unwrap() != http.ErrMissingFile {
		t.Error("Unwrap should return the cause")
	}
}

// TestTypedHandlerWithStatusFunc tests the status code variant.
func TestTypedHandlerWithStatusFunc(t *testing.T) {
	app := New()

	type Request struct {
		Name string `json:"name"`
	}

	type Response struct {
		Message string `json:"message"`
	}

	app.Post("/custom-status", TypedHandlerWithStatusFunc(
		func(c *Context, req Request) (int, Response, error) {
			if req.Name == "admin" {
				return 201, Response{Message: "Created with admin"}, nil
			}
			return 200, Response{Message: "OK"}, nil
		},
	))

	// Test 201 status
	body, _ := json.Marshal(Request{Name: "admin"})
	req := httptest.NewRequest("POST", "/custom-status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	// Test 200 status
	body, _ = json.Marshal(Request{Name: "user"})
	req = httptest.NewRequest("POST", "/custom-status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// BenchmarkTypedHandler benchmarks the performance of typed handlers.
func BenchmarkTypedHandler(b *testing.B) {
	app := New()

	type Request struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Response struct {
		Message string `json:"message"`
	}

	app.Post("/bench", TypedHandlerFunc(func(c *Context, req Request) (Response, error) {
		return Response{Message: "Hello " + req.Name}, nil
	}))

	body, _ := json.Marshal(Request{Name: "John", Age: 30})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/bench", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)
	}
}

// BenchmarkTypeCaching benchmarks the type caching mechanism.
func BenchmarkTypeCaching(b *testing.B) {
	type Request struct {
		Value int `json:"value"`
	}

	b.Run("WithCaching", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = TypedHandlerFunc(func(c *Context, req Request) (Request, error) {
				return req, nil
			})
		}
	})
}
