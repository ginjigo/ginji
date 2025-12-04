# Middleware

Ginji comes with a comprehensive set of production-ready middleware.

## Overview

Middleware in Ginji intercepts requests before they reach your handlers, allowing you to add cross-cutting concerns like logging, authentication, rate limiting, and more.

```go
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
```

## Built-in Middleware

### Security

- **[Body Limit](/middleware/body-limit)** - Prevent DOS attacks by limiting request body size
- **[Security Headers](/middleware/security)** - Add security headers (XSS, HSTS, CSP, CORS)
- **[Authentication](/middleware/auth)** - Basic Auth, Bearer tokens, API keys, RBAC
- **[Rate Limiting](/middleware/rate-limit)** - Token bucket-based rate limiting

### Operations

- **[Health Checks](/middleware/health)** - Kubernetes-style liveness/readiness probes
- **[Timeout](/middleware/timeout)** - Context-based request timeouts
- **Logger** - HTTP request/response logging
- **Recovery** - Panic recovery with stack traces

### Performance

- **Compress** - Gzip compression for responses
- **Request ID** - Add unique request IDs for tracing

## Quick Example

```go
package main

import (
    "time"
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Security middleware
    app.Use(middleware.SecureStrict())
    app.Use(middleware.BodyLimit10MB())
    app.Use(middleware.CORS(middleware.DefaultCORSOptions()))

    // Operations
    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())
    app.Use(middleware.Health())

    // Performance
    app.Use(middleware.Timeout(30 * time.Second))
    app.Use(middleware.RateLimitPerMinute(100))
    app.Use(middleware.Compress())

    // Your routes...
    app.Get("/", handler)

    app.Listen(":8080")
}
```

## Custom Middleware

Creating custom middleware is easy:

```go
func CustomLogger() ginji.Middleware {
    return func(next ginji.Handler) ginji.Handler {
        return func(c *ginji.Context) {
            start := time.Now()

            // Before request
            log.Printf("Started %s %s", c.Req.Method, c.Req.URL.Path)

            // Process request
            next(c)

            // After request
            duration := time.Since(start)
            log.Printf("Completed in %v", duration)
        }
    }
}

// Use it
app.Use(CustomLogger())
```

## Middleware Chaining

Middleware executes in the order it's registered:

```go
app.Use(middleware.RequestID())      // 1. Add request ID
app.Use(middleware.Logger())         // 2. Log with ID
app.Use(middleware.Recovery())       // 3. Catch panics
app.Use(middleware.RateLimit(...))   // 4. Check rate limit
app.Use(middleware.BasicAuth(...))   // 5. Authenticate
```

## Route-Specific Middleware

Apply middleware to specific routes or groups:

```go
// Single route
app.Get("/admin", middleware.RequireRole("admin"), adminHandler)

// Route group
admin := app.Group("/api/admin")
admin.Use(middleware.BasicAuth(users))
admin.Use(middleware.RequireRole("admin"))
admin.Get("/stats", getStats)
admin.Post("/users", createUser)
```

## Next Steps

Explore individual middleware:

- [Body Limit](/middleware/body-limit) - DOS protection
- [Security Headers](/middleware/security) - XSS, HSTS, CSP
- [Health Checks](/middleware/health) - K8s probes
- [Rate Limiting](/middleware/rate-limit) - Abuse prevention
- [Authentication](/middleware/auth) - Auth strategies
- [Timeout](/middleware/timeout) - Request timeouts
