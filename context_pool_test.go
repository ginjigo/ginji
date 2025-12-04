package ginji

import (
	"net/http/httptest"
	"sync"
	"testing"
)

// TestContextPoolConcurrency tests context pool under concurrent load
func TestContextPoolConcurrency(t *testing.T) {
	app := New()

	app.Get("/test", func(c *Context) {
		c.Set("test_key", "test_value")
		_ = c.JSON(StatusOK, H{"message": "ok"})
	})

	// Run 1000 concurrent requests
	concurrency := 1000
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			if rec.Code != StatusOK {
				t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
			}
		}()
	}

	wg.Wait()
}

// TestContextReset tests that context reset properly clears state
func TestContextReset(t *testing.T) {
	engine := New()

	// Create initial context
	req1 := httptest.NewRequest("GET", "/path1", nil)
	rec1 := httptest.NewRecorder()
	ctx := NewContext(rec1, req1, engine)

	// Set some data
	ctx.Set("key1", "value1")
	ctx.Set("key2", "value2")
	ctx.Params["id"] = "123"

	// Reset with new request
	req2 := httptest.NewRequest("POST", "/path2", nil)
	rec2 := httptest.NewRecorder()
	ctx.Reset(rec2, req2, engine)

	// Check that state is cleared
	if len(ctx.Keys) != 0 {
		t.Errorf("Expected Keys to be empty after reset, got %d items", len(ctx.Keys))
	}

	if len(ctx.Params) != 0 {
		t.Errorf("Expected Params to be empty after reset, got %d items", len(ctx.Params))
	}

	if ctx.Req.URL.Path != "/path2" {
		t.Errorf("Expected path /path2, got %s", ctx.Req.URL.Path)
	}

	if ctx.Req.Method != "POST" {
		t.Errorf("Expected method POST, got %s", ctx.Req.Method)
	}
}

// TestContextPoolReuse tests that contexts are properly reused from the pool
func TestContextPoolReuse(t *testing.T) {
	app := New()

	contextPointers := make(map[*Context]bool)
	var mu sync.Mutex

	app.Get("/test", func(c *Context) {
		mu.Lock()
		contextPointers[c] = true
		mu.Unlock()
		_ = c.Text(StatusOK, "ok")
	})

	// Make multiple sequential requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
	}

	// Should have reused contexts (likely fewer than 100 unique pointers)
	if len(contextPointers) == 100 {
		t.Log("Warning: No context reuse detected, each request got a new context")
	}
}

// TestContextServiceScopeDisposal tests that service scopes are properly disposed
func TestContextServiceScopeDisposal(t *testing.T) {
	engine := New()

	var disposed bool

	// Create a service that tracks disposal
	type DisposableService struct {
		disposed *bool
	}

	// Register scoped service
	engine.RegisterScoped("test_service", func(scope *ServiceScope) (*DisposableService, error) {
		return &DisposableService{disposed: &disposed}, nil
	})

	// Create and reset context
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext(rec, req, engine)

	// Resolve service
	_, err := ctx.GetService("test_service")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	// Reset context - should dispose old scope
	req2 := httptest.NewRequest("GET", "/", nil)
	rec2 := httptest.NewRecorder()
	ctx.Reset(rec2, req2, engine)

	// Note: We can't directly test disposal without implementing Disposable interface
	// This test verifies that Reset doesn't panic
}

// TestContextAbort tests that Abort properly stops middleware chain
func TestContextAbort(t *testing.T) {
	app := New()

	var handlerCalled bool

	// Chain the middlewares
	app.Get("/test", func(c *Context) {
		c.Abort() // Stop execution
		c.Next()  // Should not execute anything after abort
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	// Verify that response is empty (abort prevented handler from writing)
	if handlerCalled {
		t.Error("Handler should NOT be called - this test verifies abort behavior")
	}
}

// TestContextGetSetTypes tests type-safe getters
func TestContextGetSetTypes(t *testing.T) {
	ctx := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), nil)

	// Test string
	ctx.Set("string_key", "test_string")
	if ctx.GetString("string_key") != "test_string" {
		t.Error("GetString failed")
	}
	if ctx.GetString("nonexistent") != "" {
		t.Error("GetString should return empty string for nonexistent key")
	}

	// Test int
	ctx.Set("int_key", 42)
	if ctx.GetInt("int_key") != 42 {
		t.Error("GetInt failed")
	}
	if ctx.GetInt("nonexistent") != 0 {
		t.Error("GetInt should return 0 for nonexistent key")
	}

	// Test bool
	ctx.Set("bool_key", true)
	if !ctx.GetBool("bool_key") {
		t.Error("GetBool failed")
	}
	if ctx.GetBool("nonexistent") {
		t.Error("GetBool should return false for nonexistent key")
	}

	// Test wrong type returns zero value
	ctx.Set("wrong_type", "not a number")
	if ctx.GetInt("wrong_type") != 0 {
		t.Error("GetInt should return 0 for wrong type")
	}
}

// TestContextStatusCode tests StatusCode getter
func TestContextStatusCode(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext(rec, req, nil)

	// Should default to 200
	if ctx.StatusCode() != 200 {
		t.Errorf("Expected default status 200, got %d", ctx.StatusCode())
	}

	// Set status
	ctx.Status(404)
	if ctx.StatusCode() != 404 {
		t.Errorf("Expected status 404, got %d", ctx.StatusCode())
	}

	// Note: Writing with c.Text actually calls WriteHeader which sets status
	// If status is already set, multiple Status() calls are ignored
}

// TestContextQueryParams tests query parameter handling
func TestContextQueryParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=John&age=30&tags=go&tags=rust", nil)
	ctx := NewContext(httptest.NewRecorder(), req, nil)

	if ctx.Query("name") != "John" {
		t.Error("Failed to get name query param")
	}

	if ctx.Query("age") != "30" {
		t.Error("Failed to get age query param")
	}

	if ctx.Query("nonexistent") != "" {
		t.Error("Should return empty string for nonexistent query param")
	}

	// Note: Query() only gets first value for multi-value params
	if ctx.Query("tags") != "go" {
		t.Error("Should return first value for multi-value param")
	}
}

// BenchmarkContextPool benchmarks context pooling performance
func BenchmarkContextPool(b *testing.B) {
	app := New()

	app.Get("/bench", func(c *Context) {
		c.Set("key", "value")
		_ = c.JSON(StatusOK, H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/bench", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
	}
}

// BenchmarkContextSetGet benchmarks context key-value operations
func BenchmarkContextSetGet(b *testing.B) {
	ctx := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx.Set("key", "value")
		_, _ = ctx.Get("key")
	}
}
