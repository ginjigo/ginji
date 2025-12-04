# Authentication

Secure your API with built-in authentication middleware.

## Basic Auth

```go
users := map[string]string{
    "admin": "secret123",
    "user":  "password456",
}

app.Use(middleware.BasicAuth(users))
```

## Bearer Token

```go
app.Use(middleware.BearerAuth(func(token string) (any, bool) {
    user, err := validateJWT(token)
    if err != nil {
        return nil, false
    }
    return user, true
}))
```

## API Key

### Header-based

```go
app.Use(middleware.APIKey("X-API-Key", func(key string) (any, bool) {
    return db.FindUserByAPIKey(key)
}))
```

### Query parameter

```go
config := middleware.APIKeyConfig{
    Header:    "X-API-Key",
    Query:     "api_key",
    Validator: validateKey,
}
app.Use(middleware.APIKeyWithConfig(config))
```

## Role-Based Access Control

```go
// Require specific role
admin := app.Group("/admin")
admin.Use(middleware.BearerAuth(validateToken))
admin.Use(middleware.RequireRole("admin"))
admin.Get("/dashboard", adminDashboard)
```

## Complete Example

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Public routes
    public := app.Group("/api/public")
    public.Get("/status", getStatus)

    // Basic auth routes
    basic := app.Group("/api/basic")
    basic.Use(middleware.BasicAuth(map[string]string{
        "user": "pass",
    }))
    basic.Get("/profile", getProfile)

    // Bearer token routes
    protected := app.Group("/api")
    protected.Use(middleware.BearerAuth(validateToken))
    protected.Get("/users", listUsers)
    protected.Post("/users", createUser)

    // Admin routes
    admin := app.Group("/api/admin")
    admin.Use(middleware.BearerAuth(validateToken))
    admin.Use(middleware.RequireRole("admin"))
    admin.Get("/stats", getStats)
    admin.Delete("/users/:id", deleteUser)

    // API key routes
    external := app.Group("/api/external")
    external.Use(middleware.APIKey("X-API-Key", validateAPIKey))
    external.Get("/data", getData)

    app.Listen(":8080")
}

func validateToken(token string) (any, bool) {
    // Validate JWT or session token
    // Return user object and validity
    return map[string]any{
        "id":   "123",
        "role": "admin",
    }, true
}

func validateAPIKey(key string) (any, bool) {
    // Validate API key
    // Return client info and validity
    return map[string]any{
        "client": "CompanyA",
    }, true
}
```

## Custom Validators

### LDAP Authentication

```go
app.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
    Validator: func(username, password string) bool {
        return ldap.Authenticate(username, password)
    },
}))
```

### Database Validation

```go
app.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
    Validator: func(username, password string) bool {
        user, err := db.FindUser(username)
        if err != nil {
            return false
        }
        return bcrypt.CompareHashAndPassword(
            user.PasswordHash,
            []byte(password),
        ) == nil
    },
}))
```

## Access User in Handler

```go
app.Get("/profile", func(c *ginji.Context) {
    // Get authenticated user
    user := c.Get("user")
    
    // Or with type assertion
    username := c.GetString("user")
    
    c.JSON(ginji.StatusOK, ginji.H{
        "user": username,
    })
})
```

## Custom Context Key

```go
config := middleware.BasicAuthConfig{
    Users:      users,
    ContextKey: "authenticated_user",
}
app.Use(middleware.BasicAuthWithConfig(config))

// In handler
user := c.GetString("authenticated_user")
```

## Multiple Auth Strategies

```go
// Try Bearer first, fall back to API key
app.Use(func(next ginji.Handler) ginji.Handler {
    return func(c *ginji.Context) {
        // Try Bearer
        if token := c.Header("Authorization"); token != "" {
            if user, ok := validateBearer(token); ok {
                c.Set("user", user)
                next(c)
                return
            }
        }
        
        // Try API Key
        if key := c.Header("X-API-Key"); key != "" {
            if user, ok := validateAPIKey(key); ok {
                c.Set("user", user)
                next(c)
                return
            }
        }
        
        c.AbortWithStatusJSON(ginji.StatusUnauthorized, ginji.H{
            "error": "Authentication required",
        })
    }
})
```

## Security Best Practices

1. **Use HTTPS** - Always in production
2. **Hash passwords** - Use bcrypt or argon2
3. **Validate tokens** - Check expiration and signature
4. **Use strong secrets** - For JWT signing
5. **Rotate API keys** - Periodically
6. **Rate limit** - Prevent brute force
7. **Audit logs** - Track auth attempts

## Examples

### JWT Validation

```go
import "github.com/golang-jwt/jwt/v5"

func validateJWT(tokenString string) (any, bool) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte("your-secret-key"), nil
    })
    
    if err != nil || !token.Valid {
        return nil, false
    }
    
    claims := token.Claims.(jwt.MapClaims)
    return claims, true
}
```

### With Rate Limiting

```go
app.Use(middleware.RateLimitPerMinute(5)) // Limit login attempts
app.Post("/login", loginHandler)
```

## Features

- ✅ Basic Auth with constant-time comparison
- ✅ Bearer token support
- ✅ API Key (header or query)
- ✅ Role-based access control
- ✅ Custom validators
- ✅ Flexible context storage
- ✅ Multiple auth strategies
- ✅ WWW-Authenticate headers
