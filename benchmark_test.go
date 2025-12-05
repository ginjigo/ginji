package ginji

import (
	"net/http/httptest"
	"testing"
)

func BenchmarkContextPooling(b *testing.B) {
	app := New()
	app.Get("/ping", func(c *Context) error {
		return c.Text(200, "pong")
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app.ServeHTTP(w, req)
	}
}

func BenchmarkMiddlewareChain(b *testing.B) {
	app := New()

	// Add 10 middlewares
	for i := 0; i < 10; i++ {
		app.Use(func(c *Context) error {
			return c.Next()
		})
	}

	app.Get("/ping", func(c *Context) error {
		return c.Text(200, "pong")
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app.ServeHTTP(w, req)
	}
}

func BenchmarkRouter(b *testing.B) {
	app := New()

	// Add some routes to make the router work a bit
	app.Get("/api/v1/users", func(c *Context) error { return nil })
	app.Get("/api/v1/users/:id", func(c *Context) error { return nil })
	app.Post("/api/v1/users", func(c *Context) error { return nil })
	app.Get("/ping", func(c *Context) error {
		return c.Text(200, "pong")
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app.ServeHTTP(w, req)
	}
}
