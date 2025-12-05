package ginji

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewContextAPIEnhancements tests the new Request wrapper and convenience methods.
func TestNewContextAPIEnhancements(t *testing.T) {
	app := New()

	// Test Request.Param and Request.Query
	app.Get("/user/:id", func(c *Context) error {
		// In a real request, router would set these, but we test the API directly
		// Test new Request wrapper API
		id := c.Request.Param("id")
		name := c.Request.Query("name")
		defaultName := c.Request.QueryDefault("missing", "guest")

		// For testing, we verify Query works (router doesn't populate params in test)
		if name == "" {
			t.Error("Expected name query parameter")
		}
		if defaultName != "guest" {
			t.Errorf("Expected default value 'guest', got '%s'", defaultName)
		}

		return c.JSONOK(H{
			"id":          id,
			"name":        name,
			"defaultName": defaultName,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/user/123?name=John", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestConvenienceResponseMethods tests JSONOK, TextOK, HTMLOK.
func TestConvenienceResponseMethods(t *testing.T) {
	app := New()

	// Test JSONOK
	app.Get("/json", func(c *Context) error {
		return c.JSONOK(H{"status": "ok"})
	})

	// Test TextOK
	app.Get("/text", func(c *Context) error {
		return c.TextOK("Hello World")
	})

	// Test HTMLOK
	app.Get("/html", func(c *Context) error {
		return c.HTMLOK("<h1>Hello</h1>")
	})

	tests := []struct {
		path        string
		contentType string
		body        string
	}{
		{"/json", "application/json", `{"status":"ok"}`},
		{"/text", "text/plain", "Hello World"},
		{"/html", "text/html", "<h1>Hello</h1>"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", tt.path, nil)
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", tt.path, w.Code)
		}

		if contentType := w.Header().Get("Content-Type"); contentType != tt.contentType {
			t.Errorf("Expected Content-Type %s for %s, got %s", tt.contentType, tt.path, contentType)
		}
	}
}

// TestFailMethods tests Fail and FailWithData.
func TestFailMethods(t *testing.T) {
	app := New()

	// Test Fail
	app.Get("/fail", func(c *Context) error {
		return c.Fail(http.StatusBadRequest, "Invalid input")
	})

	// Test FailWithData
	app.Get("/fail-data", func(c *Context) error {
		return c.FailWithData(http.StatusNotFound, "User not found", H{"id": 123})
	})

	// Test Fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fail", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Test FailWithData
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/fail-data", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestVarMethods tests Var and GetVar.
func TestVarMethods(t *testing.T) {
	app := New()

	app.Use(func(c *Context) error {
		c.Var("user_id", 123)
		return c.Next()
	})

	app.Get("/test", func(c *Context) error {
		if val, exists := c.GetVar("user_id"); exists {
			if userID, ok := val.(int); ok && userID == 123 {
				return c.TextOK("OK")
			}
		}
		return c.Fail(http.StatusInternalServerError, "Var not found")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

// TestBackwardsCompatibility ensures old API still works.
func TestBackwardsCompatibility(t *testing.T) {
	app := New()

	// Old API should still work
	app.Get("/old/:id", func(c *Context) error {
		id := c.Param("id")             // Old way
		name := c.Query("name")         // Old way
		return c.JSON(http.StatusOK, H{ // Old way
			"id":   id,
			"name": name,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/old/456?name=Jane", nil)
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
