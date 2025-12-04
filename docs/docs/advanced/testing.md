# Testing

Comprehensive testing utilities for your Ginji applications.

## Test Context

Create test contexts for unit testing:

```go
func TestHandler(t *testing.T) {
    c, rec := ginji.NewTestContext()
    
    handler(c)
    
    if rec.Code != ginji.StatusOK {
        t.Errorf("Expected 200, got %d", rec.Code)
    }
}
```

## Perform Request

Simulate HTTP requests:

```go
func TestGetUsers(t *testing.T) {
    app := ginji.New()
    app.Get("/users", getUsers)
    
    w := ginji.PerformRequest(app, "GET", "/users", nil)
    
    ginji.AssertStatus(t, w, ginji.StatusOK)
    ginji.AssertBody(t, w, "users")
}
```

## JSON Requests

```go
func TestCreateUser(t *testing.T) {
    app := ginji.New()
    app.Post("/users", createUser)
    
    payload := ginji.H{
        "name": "John",
        "email": "john@example.com",
    }
    
    w := ginji.PerformJSONRequest(app, "POST", "/users", payload)
    
    ginji.AssertStatus(t, w, ginji.StatusCreated)
}
```

## Request Builder

Build complex requests:

```go
func TestAuthEndpoint(t *testing.T) {
    app := ginji.New()
    app.Get("/profile", getProfile)
    
    w := ginji.NewRequest(app, "GET", "/profile").
        Header("Authorization", "Bearer token123").
        Header("X-API-Key", "key456").
        Do()
    
    ginji.AssertStatus(t, w, ginji.StatusOK)
}
```

## Assertions

### Status Code

```go
ginji.AssertStatus(t, w, ginji.StatusOK)
```

### Body Content

```go
ginji.AssertBody(t, w, "expected content")
```

### Headers

```go
ginji.AssertHeader(t, w, "Content-Type", "application/json")
```

### JSON Response

```go
ginji.AssertJSON(t, w, "user.email", "test@example.com")
```

## Mock Middleware

Test middleware behavior:

```go
func TestMiddleware(t *testing.T) {
    called := false
    
    mw := ginji.MockMiddleware(func(c *ginji.Context) {
        called = true
    })
    
    app := ginji.New()
    app.Use(mw)
    app.Get("/test", func(c *ginji.Context) {
        c.Text(ginji.StatusOK, "ok")
    })
    
    ginji.PerformRequest(app, "GET", "/test", nil)
    
    if !called {
        t.Error("Middleware not called")
    }
}
```

## Complete Example

```go
package main

import (
    "testing"
    "github.com/ginjigo/ginji"
)

func TestUserAPI(t *testing.T) {
    app := setupApp()

    t.Run("List users", func(t *testing.T) {
        w := ginji.PerformRequest(app, "GET", "/users", nil)
        ginji.AssertStatus(t, w, ginji.StatusOK)
    })

    t.Run("Create user", func(t *testing.T) {
        payload := ginji.H{
            "name": "John",
            "email": "john@example.com",
        }
        
        w := ginji.PerformJSONRequest(app, "POST", "/users", payload)
        ginji.AssertStatus(t, w, ginji.StatusCreated)
        ginji.AssertBody(t, w, "john@example.com")
    })

    t.Run("Get user", func(t *testing.T) {
        w := ginji.PerformRequest(app, "GET", "/users/1", nil)
        ginji.AssertStatus(t, w, ginji.StatusOK)
        ginji.AssertJSON(t, w, "id", "1")
    })

    t.Run("Invalid request", func(t *testing.T) {
        w := ginji.PerformJSONRequest(app, "POST", "/users", ginji.H{})
        ginji.AssertStatus(t, w, ginji.StatusBadRequest)
    })
}
```

## Best Practices

1. **Test handlers separately** - Unit test business logic
2. **Use table-driven tests** - Test multiple scenarios
3. **Mock dependencies** - Database, external APIs
4. **Test middleware** - Verify auth, rate limiting
5. **Integration tests** - Test full request flow
