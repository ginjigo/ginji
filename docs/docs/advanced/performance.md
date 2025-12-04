# Performance

Optimize your Ginji applications for maximum performance.

## Routing Performance

Ginji uses a trie-based router for O(log n) route matching.

```go
// Efficient routing
app.Get("/users/:id", getUser)
app.Get("/posts/:id", getPost)
app.Get("/api/v1/*path", handleAPI)
```

## Middleware Ordering

Order middleware for best performance:

```go
// Fast checks first
app.Use(middleware.RequestID())      // Cheap
app.Use(middleware.Recovery())       // Only on panic
app.Use(middleware.RateLimit(...))  // Before expensive auth
app.Use(middleware.BearerAuth(...)) // Expensive validation
```

## Response Compression

Enable gzip for text responses:

```go
app.Use(middleware.Compress())
```

## Connection Pooling

Reuse database connections:

```go
var db *sql.DB

func init() {
    db, _ = sql.Open("postgres", dsn)
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
}
```

## Caching

Cache expensive operations:

```go
var cache = make(map[string]any)
var mu sync.RWMutex

app.Get("/data/:id", func(c *ginji.Context) {
    id := c.Param("id")
    
    // Check cache
    mu.RLock()
    if cached, ok := cache[id]; ok {
        mu.RUnlock()
        c.JSON(ginji.StatusOK, cached)
        return
    }
    mu.RUnlock()
    
    // Fetch and cache
    data := fetchData(id)
    
    mu.Lock()
    cache[id] = data
    mu.Unlock()
    
    c.JSON(ginji.StatusOK, data)
})
```

## JSON Performance

Use efficient JSON encoding:

```go
// Reuse encoder
var jsonEncoder = json.NewEncoder(writer)

// Or use Pre-marshaled responses
type CachedResponse struct {
    data []byte
}

func (r *CachedResponse) MarshalJSON() ([]byte, error) {
    return r.data, nil
}
```

## Benchmarking

```go
func BenchmarkHandler(b *testing.B) {
    app := ginji.New()
    app.Get("/test", handler)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ginji.PerformRequest(app, "GET", "/test", nil)
    }
}
```

## Production Config

```go
func main() {
    // Set release mode
    ginji.SetMode(ginji.ReleaseMode)
    
    app := ginji.New()
    
    // Production middleware stack
    app.Use(middleware.Recovery())
    app.Use(middleware.Compress())
    app.Use(middleware.RateLimitPerMinute(1000))
    app.Use(middleware.Timeout(30 * time.Second))
    
    // Your routes
    
    app.Listen(":8080")
}
```

## Best Practices

1. **Use connection pools** - Database, Redis, HTTP clients
2. **Enable compression** - For text responses
3. **Cache aggressively** - Static and semi-static data
4. **Minimize allocations** - Reuse buffers and objects
5. **Profile in production** - Use pprof to find bottlenecks
6. **Set timeouts** - Prevent resource exhaustion
7. **Rate limit** - Protect against abuse
