package ginji

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRouterConflictingPatterns tests handling of potentially conflicting route patterns
func TestRouterConflictingPatterns(t *testing.T) {
	app := New()

	// Register /users/new first
	app.Get("/users/new", func(c *Context) error {
		return c.Text(200, "new user form")
	})

	// Then register /users/:id - should not conflict with /users/new
	app.Get("/users/:id", func(c *Context) error {
		return c.Text(200, "user: "+c.Param("id"))
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/users/new", "new user form"},
		{"/users/123", "user: 123"},
		{"/users/abc", "user: abc"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, body)
			}
		})
	}
}

// TestRouterEmptyPattern tests handling of empty route patterns
func TestRouterEmptyPattern(t *testing.T) {
	app := New()

	app.Get("", func(c *Context) error {
		return c.Text(200, "root")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200 for empty pattern, got %d", w.Code)
	}
}

// TestRouterMultipleWildcards tests patterns with multiple parameter segments
func TestRouterMultipleWildcards(t *testing.T) {
	app := New()

	app.Get("/api/:version/users/:id", func(c *Context) error {
		version := c.Param("version")
		id := c.Param("id")
		return c.Text(200, version+":"+id)
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := "v1:123"
	actual := strings.TrimSpace(w.Body.String())
	if actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

// TestRouterCatchAllWildcard tests * wildcard for catching all remaining path segments
func TestRouterCatchAllWildcard(t *testing.T) {
	app := New()

	app.Get("/files/*filepath", func(c *Context) error {
		filepath := c.Param("filepath")
		return c.Text(200, "filepath: "+filepath)
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/files/a.txt", "filepath: a.txt"},
		{"/files/dir/subdir/file.txt", "filepath: dir/subdir/file.txt"},
		{"/files/", "filepath: "},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, body)
			}
		})
	}
}

// TestRouterSpecialCharactersInParams tests route params with special characters
func TestRouterSpecialCharactersInParams(t *testing.T) {
	app := New()

	app.Get("/users/:id", func(c *Context) error {
		return c.Text(200, "id: "+c.Param("id"))
	})

	tests := []struct {
		path string
		id   string
	}{
		{"/users/user-123", "user-123"},
		{"/users/user_123", "user_123"},
		{"/users/user.123", "user.123"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			expected := "id: " + tt.id
			actual := strings.TrimSpace(w.Body.String())
			if actual != expected {
				t.Errorf("Expected %q, got %q", expected, actual)
			}
		})
	}
}

// TestRouterDeepNesting tests very deeply nested routes
func TestRouterDeepNesting(t *testing.T) {
	app := New()

	// Create a deeply nested route (10 levels)
	pattern := "/a/b/c/d/e/f/g/h/i/j"
	app.Get(pattern, func(c *Context) error {
		return c.Text(200, "deep")
	})

	req := httptest.NewRequest("GET", pattern, nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200 for deep route, got %d", w.Code)
	}

	if strings.TrimSpace(w.Body.String()) != "deep" {
		t.Error("Deep route not matched correctly")
	}
}

// TestRouterMethodHandling tests that different HTTP methods are distinguished
func TestRouterMethodHandling(t *testing.T) {
	app := New()

	app.Get("/resource", func(c *Context) error {
		return c.Text(200, "GET")
	})

	app.Post("/resource", func(c *Context) error {
		return c.Text(200, "POST")
	})

	tests := []struct {
		method   string
		expected string
	}{
		{"GET", "GET"},
		{"POST", "POST"},
		{"PUT", "404 NOT FOUND"}, // No PUT handler registered
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/resource", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expected {
				t.Errorf("Expected %q for %s, got %q", tt.expected, tt.method, body)
			}
		})
	}
}

// TestRouterTrailingSlashHandling tests handling of trailing slashes
func TestRouterTrailingSlashHandling(t *testing.T) {
	app := New()

	app.Get("/users", func(c *Context) error {
		return c.Text(200, "users")
	})

	tests := []struct {
		path       string
		shouldWork bool
	}{
		{"/users", true},
		{"/users/", false}, // Trailing slash is different route
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if tt.shouldWork {
				if w.Code != 200 {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
			} else {
				if w.Code == 200 {
					t.Error("Expected non-200 status for unregistered route")
				}
			}
		})
	}
}

// TestRouterParamPersistence tests that params don't leak between requests
func TestRouterParamPersistence(t *testing.T) {
	app := New()

	app.Get("/users/:id", func(c *Context) error {
		id := c.Param("id")
		return c.Text(200, id)
	})

	// First request
	req1 := httptest.NewRequest("GET", "/users/123", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)

	// Second request with different ID
	req2 := httptest.NewRequest("GET", "/users/456", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)

	if strings.TrimSpace(w1.Body.String()) != "123" {
		t.Error("First request did not get correct param")
	}

	if strings.TrimSpace(w2.Body.String()) != "456" {
		t.Error("Second request did not get correct param (param leakage)")
	}
}

// TestRouterGroupPrefix tests that group prefixes work correctly
func TestRouterGroupPrefix(t *testing.T) {
	app := New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/users", func(c *Context) error {
		return c.Text(200, "v1 users")
	})

	api.Group("/v2").Get("/users", func(c *Context) error {
		return c.Text(200, "v2 users")
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/users", "v1 users"},
		{"/api/v2/users", "v2 users"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, body)
			}
		})
	}
}

// TestRouterNotFound tests 404 handling
func TestRouterNotFound(t *testing.T) {
	app := New()

	app.Get("/exists", func(c *Context) error {
		return c.Text(200, "found")
	})

	req := httptest.NewRequest("GET", "/doesnotexist", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "404") {
		t.Error("404 response should contain '404'")
	}
}
