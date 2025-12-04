package ginji

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	app := New()
	app.Use(RequestID())
	app.Get("/", func(c *Context) {
		_ = c.Text(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header")
	}
}

func TestCompress(t *testing.T) {
	app := New()
	app.Use(Compress())
	app.Get("/", func(c *Context) {
		_ = c.Text(http.StatusOK, "compressed content")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected Content-Encoding: gzip")
	}

	// Verify content
	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gr.Close() }()

	body, _ := io.ReadAll(gr)
	if string(body) != "compressed content" {
		t.Errorf("Expected compressed content, got %s", string(body))
	}
}
