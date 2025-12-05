package ginji

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// Test request types
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=1,max=150"`
}

type CreateUserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetUserParams struct {
	ID string `path:"id" validate:"required"`
}

type UpdateUserRequest struct {
	ID    string `path:"id" validate:"required"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestTypedHandlerPOST(t *testing.T) {
	app := New()

	// Register a typed POST handler
	app.Typed().Post("/users", func(c *Context, req CreateUserRequest) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	// Valid request
	reqBody := CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
	}

	var res CreateUserResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res.Name != reqBody.Name {
		t.Errorf("Expected name %s, got %s", reqBody.Name, res.Name)
	}
}

func TestTypedHandlerGET(t *testing.T) {
	app := New()

	// Register a typed GET handler with path params
	app.Typed().Get("/users/:id", func(c *Context, req GetUserParams) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		}, nil
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
	}

	var res CreateUserResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res.Name != "John Doe" {
		t.Errorf("Expected name John Doe, got %s", res.Name)
	}
}

func TestTypedHandlerValidation(t *testing.T) {
	app := New()

	// Define a struct with ginji validation tags
	type ValidatedRequest struct {
		Email string `json:"email" validate:"email"`
		Age   int    `json:"age" validate:"max=150"`
	}

	app.Typed().Post("/users", func(c *Context, req ValidatedRequest) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:    1,
			Name:  "Test",
			Email: req.Email,
		}, nil
	})

	// Invalid request (invalid email format and age out of range)
	reqBody := map[string]any{
		"email": "invalid-email", // Invalid email format
		"age":   200,             // Age too high (max=150)
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	// Should return validation error
	if rec.Code == StatusOK {
		t.Errorf("Expected validation error, got status %d", rec.Code)
	}
}

func TestTypedHandlerPUT(t *testing.T) {
	app := New()

	// Register a typed PUT handler
	app.Typed().Put("/users/:id", func(c *Context, req UpdateUserRequest) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	reqBody := map[string]string{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/users/123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
	}

	var res CreateUserResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res.Name != "Jane Doe" {
		t.Errorf("Expected name Jane Doe, got %s", res.Name)
	}
}

func TestTypedHandlerDELETE(t *testing.T) {
	app := New()

	// Register a typed DELETE handler that returns EmptyRequest (no response body)
	app.Typed().Delete("/users/:id", func(c *Context, req GetUserParams) (EmptyRequest, error) {
		// Simulate deletion
		return EmptyRequest{}, nil
	})

	req := httptest.NewRequest("DELETE", "/users/123", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusNoContent {
		t.Errorf("Expected status %d, got %d", StatusNoContent, rec.Code)
	}
}

func TestTypedHandlerError(t *testing.T) {
	app := New()

	// Register a typed handler that returns an error
	app.Typed().Get("/error", func(c *Context, req EmptyRequest) (CreateUserResponse, error) {
		return CreateUserResponse{}, NewHTTPError(StatusNotFound, "User not found")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusNotFound {
		t.Errorf("Expected status %d, got %d", StatusNotFound, rec.Code)
	}
}

func TestTypedHandlerWithGroup(t *testing.T) {
	app := New()

	// Create a group and use typed handlers
	api := app.Group("/api/v1")

	api.Typed().Get("/users/:id", func(c *Context, req GetUserParams) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}, nil
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
	}

	var res CreateUserResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res.Name != "Test User" {
		t.Errorf("Expected name Test User, got %s", res.Name)
	}
}

func TestTypedHandlerFunc(t *testing.T) {
	// Test the generic TypedHandlerFunc wrapper directly
	handler := TypedHandlerFunc(func(c *Context, req CreateUserRequest) (CreateUserResponse, error) {
		return CreateUserResponse{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	reqBody := CreateUserRequest{
		Name:  "Direct Test",
		Email: "direct@example.com",
		Age:   25,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := NewContext(rec, req, nil)
	if err := handler(c); err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
	}

	var res CreateUserResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res.Name != reqBody.Name {
		t.Errorf("Expected name %s, got %s", reqBody.Name, res.Name)
	}
}

func TestEmptyRequestAndResponse(t *testing.T) {
	app := New()

	// Handler with no request or response body
	app.Typed().Get("/ping", func(c *Context, req EmptyRequest) (EmptyRequest, error) {
		return EmptyRequest{}, nil
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != StatusNoContent {
		t.Errorf("Expected status %d for empty response, got %d", StatusNoContent, rec.Code)
	}
}
