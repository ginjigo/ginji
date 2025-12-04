# Timeout

Prevent resource exhaustion with context-based request timeouts.

## Basic Usage

```go
import (
    "time"
    "github.com/ginjigo/ginji/middleware"
)

app.Use(middleware.Timeout(30 * time.Second))
```

## Helper Functions

```go
app.Use(middleware.TimeoutSeconds(30))  // 30 seconds
app.Use(middleware.TimeoutMinutes(5))   // 5 minutes
```

## Custom Configuration

```go
config := middleware.TimeoutConfig{
    Timeout:      10 * time.Second,
    ErrorMessage: "Request took too long",
    StatusCode:   ginji.StatusRequestTimeout,
    
    SkipFunc: func(c *ginji.Context) bool {
        // Skip timeout for file uploads
        return strings.HasPrefix(c.Req.URL.Path, "/upload")
    },
}

app.Use(middleware.TimeoutWithConfig(config))
```

## Using Context in Handlers

The timeout middleware sets a deadline on the request context:

```go
app.Get("/long-task", func(c *ginji.Context) {
    ctx := c.Req.Context()
    
    // Start long-running operation
    resultChan := make(chan Result)
    go func() {
        result := doExpensiveWork()
        resultChan <- result
    }()
    
    // Wait for result or timeout
    select {
    case result := <-resultChan:
        c.JSON(ginji.StatusOK, result)
    case <-ctx.Done():
        // Timeout occurred, cleanup
        cleanup()
        return
    }
})
```

## Error Response

When timeout occurs (504 Gateway Timeout):

```json
{
  "error": "Request timeout",
  "timeout": "30s"
}
```

## Different Timeouts for Different Routes

```go
// Quick API endpoints
api := app.Group("/api")
api.Use(middleware.TimeoutSeconds(5))
api.Get("/status", getStatus)

// Long-running operations
reports := app.Group("/reports")
reports.Use(middleware.TimeoutMinutes(5))
reports.Get("/generate", generateReport)

// File uploads (no timeout)
uploads := app.Group("/uploads")
uploads.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
    SkipFunc: func(c *ginji.Context) bool {
        return true // Skip all timeouts
    },
}))
uploads.Post("/file", handleUpload)
```

## Database Queries

Respect timeouts in database queries:

```go
app.Get("/users", func(c *ginji.Context) {
    ctx := c.Req.Context()
    
    // Pass context to database query
    users, err := db.QueryContext(ctx, "SELECT * FROM users")
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            c.JSON(ginji.StatusGatewayTimeout, ginji.H{
                "error": "Query timeout",
            })
            return
        }
        c.JSON(ginji.StatusInternalServerError, ginji.H{
            "error": "Database error",
        })
        return
    }
    
    c.JSON(ginji.StatusOK, users)
})
```

## External API Calls

```go
app.Get("/external", func(c *ginji.Context) {
    ctx := c.Req.Context()
    
    // Create request with context
    req, _ := http.NewRequestWithContext(
        ctx,
        "GET",
        "https://api.example.com/data",
        nil,
    )
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            c.JSON(ginji.StatusGatewayTimeout, ginji.H{
                "error": "External API timeout",
            })
            return
        }
        c.JSON(ginji.StatusBadGateway, ginji.H{
            "error": "External API error",
        })
        return
    }
    defer resp.Body.Close()
    
    // Process response
})
```

## Complete Example

```go
package main

import (
    "context"
    "time"
    
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Global timeout
    app.Use(middleware.Timeout(30 * time.Second))

    // Quick endpoints
    app.Get("/health", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{"status": "ok"})
    })

    // Long-running task
    app.Get("/process", func(c *ginji.Context) {
        ctx := c.Req.Context()
        
        result := make(chan string)
        go func() {
            // Simulate long work
            time.Sleep(20 * time.Second)
            result <- "completed"
        }()
        
        select {
        case r := <-result:
            c.JSON(ginji.StatusOK, ginji.H{"result": r})
        case <-ctx.Done():
            c.JSON(ginji.StatusGatewayTimeout, ginji.H{
                "error": "Processing timeout",
            })
        }
    })

    app.Listen(":8080")
}
```

## Best Practices

1. **Set reasonable timeouts** - Balance UX and resource protection
2. **Use context** - Pass request context to long operations
3. **Different timeouts** - Vary by endpoint type
4. **Cleanup on timeout** - Cancel work, close connections
5. **Monitor timeouts** - Track frequency to adjust limits
6. **Skip when needed** - File uploads, webhooks, etc.

## Status Codes

- **408 Request Timeout** - Client took too long to send request
- **504 Gateway Timeout** - Server took too long to process (default)

Use `StatusCode` in config to choose which one fits your use case.
