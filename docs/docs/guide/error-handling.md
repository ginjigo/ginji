# Error Handling

Ginji provides a robust error handling system with structured errors and validation support.

## HTTPError

Create HTTP errors with status codes and messages:

```go
app.Get("/users/:id", func(c *ginji.Context) {
    id := c.Param("id")
    user, err := db.FindUser(id)
    
    if err != nil {
        c.AbortWithError(ginji.NewHTTPError(
            ginji.StatusNotFound,
            "User not found",
        ))
        return
    }
    
    c.JSON(ginji.StatusOK, user)
})
```

## Error with Details

Add structured error details:

```go
err := ginji.NewHTTPError(ginji.StatusBadRequest, "Invalid request").
    WithDetails(ginji.H{
        "field": "email",
        "reason": "already exists",
    })

c.AbortWithError(err)
```

## Error Handler Middleware

Use the default error handler:

```go
app.Use(ginji.DefaultErrorHandler())
```

Response format:

```json
{
  "error": "User not found",
  "status": 404
}
```

With details:

```json
{
  "error": "Invalid request",
  "status": 400,
  "details": {
    "field": "email",
    "reason": "already exists"
  }
}
```

## Validation Errors

Validation errors are automatically formatted:

```go
type CreateUserRequest struct {
    Email    string `json:"email" ginji:"required,email"`
    Password string `json:"password" ginji:"required,min=8"`
    Age      int    `json:"age" ginji:"required,gte=18"`
}

app.Post("/users", func(c *ginji.Context) {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        // Validation errors are automatically returned
        return
    }
    
    // Proceed with valid data
    createUser(req)
})
```

Validation error response:

```json
{
  "error": "Validation failed",
  "status": 400,
  "errors": [
    {
      "field": "email",
      "constraint": "email",
      "message": "email validation failed"
    },
    {
      "field": "password",
      "constraint": "min",
      "message": "min validation failed (expected: 8)"
    }
  ]
}
```

## Custom Error Handler

Create your own error handler:

```go
app.Use(func(next ginji.Handler) ginji.Handler {
    return func(c *ginji.Context) {
        next(c)
        
        if c.IsAborted() {
            if err := c.Error(); err != nil {
                // Handle HTTPError
                if httpErr, ok := err.(*ginji.HTTPError); ok {
                    c.JSON(httpErr.Status, ginji.H{
                        "success": false,
                        "error": httpErr.Message,
                        "code": httpErr.Status,
                    })
                    return
                }
                
                // Handle other errors
                c.JSON(ginji.StatusInternalServerError, ginji.H{
                    "success": false,
                    "error": "Internal server error",
                })
            }
        }
    }
})
```

## Debug Mode

Enable debug mode for detailed errors:

```go
ginji.SetMode(ginji.DebugMode)
```

In debug mode, errors include stack traces.

## Common HTTP Errors

Helper functions for common errors:

```go
// 400 Bad Request
ginji.NewBadRequestError("Invalid input")

// 401 Unauthorized
ginji.NewUnauthorizedError("Authentication required")

// 403 Forbidden
ginji.NewForbiddenError("Access denied")

// 404 Not Found
ginji.NewNotFoundError("Resource not found")

// 500 Internal Server Error
ginji.NewInternalServerError("Something went wrong")
```

## Complete Example

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

type User struct {
    Email    string `json:"email" ginji:"required,email"`
    Password string `json:"password" ginji:"required,min=8"`
    Age      int    `json:"age" ginji:"required,gte=18"`
}

func main() {
    app := ginji.New()
    
    // Enable error handler
    app.Use(ginji.DefaultErrorHandler())
    app.Use(middleware.Recovery())

    app.Post("/users", func(c *ginji.Context) {
        var user User
        
        // Validation errors are handled automatically
        if err := c.BindJSON(&user); err != nil {
            return
        }
        
        // Check if user exists
        if userExists(user.Email) {
            c.AbortWithError(
                ginji.NewHTTPError(ginji.StatusConflict, "User already exists").
                    WithDetails(ginji.H{"email": user.Email}),
            )
            return
        }
        
        // Create user
        if err := createUser(user); err != nil {
            c.AbortWithError(
                ginji.NewInternalServerError("Failed to create user"),
            )
            return
        }
        
        c.JSON(ginji.StatusCreated, ginji.H{"message": "User created"})
    })

    app.Listen(":8080")
}
```

## Best Practices

1. **Use HTTPError** - For predictable error responses
2. **Add details** - Include relevant context in errors
3. **Validate early** - Use validation tags on structs
4. **Log errors** - Log internal errors for debugging
5. **Don't expose internals** - Hide sensitive error details from clients
6. **Use appropriate status codes** - Follow HTTP conventions
