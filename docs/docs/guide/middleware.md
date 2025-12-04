# Middleware

Learn how to use and create middleware in Ginji.

## What is Middleware?

Middleware intercepts requests before they reach your handlers, allowing you to add cross-cutting concerns like logging, authentication, and rate limiting.

## Using Middleware

### Global Middleware

Apply to all routes:

```go
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
```

### Route-Specific

Apply to a single route:

```go
app.Get("/admin", middleware.BasicAuth(users), adminHandler)
```

### Group Middleware

Apply to route groups:

```go
admin := app.Group("/admin")
admin.Use(middleware.BasicAuth(users))
admin.Use(middleware.RequireRole("admin"))
{
    admin.Get("/dashboard", dashboard)
    admin.Get("/users", listUsers)
}
```

## Creating Custom Middleware

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

app.Use(CustomLogger())
```

## Built-in Middleware

See [Middleware Overview](/middleware/) for all available middleware.

## Next Steps

- [Middleware Overview](/middleware/) - All built-in middleware
- [Authentication](/middleware/auth) - Secure your API
- [Rate Limiting](/middleware/rate-limit) - Prevent abuse
