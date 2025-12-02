package ginji

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWildcardRoute(t *testing.T) {
	app := New()
	app.Get("/files/*filepath", func(c *Context) {
		_ = c.Text(http.StatusOK, c.Param("filepath"))
	})

	// Test simple file
	req := httptest.NewRequest("GET", "/files/test.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "test.txt" {
		t.Errorf("Expected test.txt, got %s", w.Body.String())
	}

	// Test nested file
	req = httptest.NewRequest("GET", "/files/css/style.css", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "css/style.css" {
		t.Errorf("Expected css/style.css, got %s", w.Body.String())
	}
}

func TestRequestWithWildcard(t *testing.T) {
	app := New()
	app.Get("/files/:name/info", func(c *Context) {
		_ = c.Text(http.StatusOK, c.Param("name"))
	})

	// Request with * in path
	req := httptest.NewRequest("GET", "/files/*/info", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "*" {
		t.Errorf("Expected *, got %s", w.Body.String())
	}
}
