package ginji

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBindValidate(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
	}

	// Test JSON binding
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","email":"john@example.com"}`))
	req.Header.Set("Content-Type", "application/json")

	c := NewContext(w, req, nil)
	var data TestStruct
	err := c.BindValidate(&data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if data.Name != "John" {
		t.Errorf("Expected name 'John', got %s", data.Name)
	}
}

func TestBindPath(t *testing.T) {
	type PathParams struct {
		ID int `path:"id"`
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/123", nil)

	c := NewContext(w, req, nil)
	c.Params = map[string]string{"id": "123"}

	var params PathParams
	err := c.BindPath(&params)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if params.ID != 123 {
		t.Errorf("Expected ID 123, got %d", params.ID)
	}
}

func TestBindAll(t *testing.T) {
	type AllParams struct {
		ID    int    `path:"id"`
		Query string `query:"q"`
		Name  string `json:"name"`
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/123?q=test", strings.NewReader(`{"name":"John"}`))
	req.Header.Set("Content-Type", "application/json")

	c := NewContext(w, req, nil)
	c.Params = map[string]string{"id": "123"}

	var params AllParams
	err := c.BindAll(&params)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if params.ID != 123 {
		t.Errorf("Expected ID 123, got %d", params.ID)
	}

	if params.Query != "test" {
		t.Errorf("Expected query 'test', got %s", params.Query)
	}

	if params.Name != "John" {
		t.Errorf("Expected name 'John', got %s", params.Name)
	}
}

func TestNegotiate(t *testing.T) {
	tests := []struct {
		name        string
		accept      string
		expectCalls string
	}{
		{"JSON", "application/json", "json"},
		{"Default", "", "json"},
		{"Text", "text/plain", "text"},
		{"HTML", "text/html", "html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", tt.accept)

			c := NewContext(w, req, nil)

			called := ""
			_ = c.Negotiate(200, map[string]string{"test": "data"}, NegotiateFormat{
				JSON: func() error {
					called = "json"
					return c.JSON(200, map[string]string{"test": "data"})
				},
				Text: func() error {
					called = "text"
					return c.Text(200, "test data")
				},
				HTML: func() error {
					called = "html"
					return c.HTML(200, "<p>test data</p>")
				},
			})

			if called != tt.expectCalls {
				t.Errorf("Expected %s to be called, got %s", tt.expectCalls, called)
			}
		})
	}
}

func TestCache(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req, nil)

	_ = c.Cache(1*time.Hour).Public().JSON(200, map[string]string{"test": "data"})

	cacheControl := w.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "public") {
		t.Errorf("Expected Cache-Control to contain 'public', got %s", cacheControl)
	}

	if !strings.Contains(cacheControl, "max-age=3600") {
		t.Errorf("Expected Cache-Control to contain 'max-age=3600', got %s", cacheControl)
	}
}

func TestETag(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req, nil)

	content := "test content"
	c.ETag(content)

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("Expected ETag header to be set")
	}

	// Test that matching ETag returns 304
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("If-None-Match", etag)
	c2 := NewContext(w2, req2, nil)
	c2.ETag(content)

	if w2.Code != 304 {
		t.Errorf("Expected status 304 for matching ETag, got %d", w2.Code)
	}
}

func TestLastModified(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req, nil)

	modTime := time.Now().UTC()
	c.LastModified(modTime)

	lastMod := w.Header().Get("Last-Modified")
	if lastMod == "" {
		t.Error("Expected Last-Modified header to be set")
	}
}

func TestNoCache(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req, nil)

	c.NoCache()

	cacheControl := w.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "no-cache") {
		t.Errorf("Expected Cache-Control to contain 'no-cache', got %s", cacheControl)
	}

	if !strings.Contains(cacheControl, "no-store") {
		t.Errorf("Expected Cache-Control to contain 'no-store', got %s", cacheControl)
	}
}

func TestBindQueryWithValidation(t *testing.T) {
	type QueryParams struct {
		Page  int `query:"page" validate:"gte=1"`
		Limit int `query:"limit" validate:"gte=1,lte=100"`
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=2&limit=50", nil)
	c := NewContext(w, req, nil)

	var params QueryParams
	err := c.BindQuery(&params)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if params.Page != 2 {
		t.Errorf("Expected page 2, got %d", params.Page)
	}

	if params.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", params.Limit)
	}
}

func TestBindForm(t *testing.T) {
	type FormData struct {
		Username string `form:"username" validate:"required"`
		Password string `form:"password" validate:"required"`
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/login", strings.NewReader("username=john&password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := NewContext(w, req, nil)

	var data FormData
	err := c.BindValidate(&data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if data.Username != "john" {
		t.Errorf("Expected username 'john', got %s", data.Username)
	}

	if data.Password != "secret" {
		t.Errorf("Expected password 'secret', got %s", data.Password)
	}
}
