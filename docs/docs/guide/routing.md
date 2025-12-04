# Routing

Ginji provides a powerful and flexible routing system based on a trie data structure.

## Basic Routes

```go
app.Get("/", handler)
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)
app.Patch("/users/:id", patchUser)
```

## Route Parameters

```go
app.Get("/users/:id", func(c *ginji.Context) {
    id := c.Param("id")
    c.JSON(ginji.StatusOK, ginji.H{"id": id})
})

app.Get("/posts/:category/:id", func(c *ginji.Context) {
    category := c.Param("category")
    id := c.Param("id")
    c.JSON(ginji.StatusOK, ginji.H{
        "category": category,
        "id": id,
    })
})
```

## Wildcard Routes

```go
// Matches any path starting with /static/
app.Get("/static/*filepath", func(c *ginji.Context) {
    filepath := c.Param("filepath")
    c.File("./public/" + filepath)
})
```

## Query Parameters

```go
app.Get("/search", func(c *ginji.Context) {
    query := c.Query("q")
    page := c.Query("page")
    
    c.JSON(ginji.StatusOK, ginji.H{
        "query": query,
        "page": page,
    })
})

// GET /search?q=golang&page=2
```

## Route Groups

Group related routes with a common prefix:

```go
api := app.Group("/api/v1")
{
    api.Get("/users", listUsers)
    api.Post("/users", createUser)
    api.Get("/users/:id", getUser)
}

// Creates routes:
// GET    /api/v1/users
// POST   /api/v1/users
// GET    /api/v1/users/:id
```

## Group Middleware

Apply middleware to specific route groups:

```go
admin := app.Group("/admin")
admin.Use(middleware.BasicAuth(users))
admin.Use(middleware.RequireRole("admin"))
{
    admin.Get("/dashboard", adminDashboard)
    admin.Get("/users", adminUsers)
}
```

## Multiple Handlers

Chain multiple handlers for a route:

```go
app.Get("/protected",
    middleware.BearerAuth(validateToken),
    middleware.RequireRole("user"),
    handler,
)
```

## HTTP Methods

All standard HTTP methods are supported:

```go
app.Get("/resource", handler)
app.Post("/resource", handler)
app.Put("/resource", handler)
app.Delete("/resource", handler)
app.Patch("/resource", handler)
app.Options("/resource", handler)
app.Head("/resource", handler)
```

## Static Files

Serve static files:

```go
// Serve single file
app.Get("/favicon.ico", func(c *ginji.Context) {
    c.File("./public/favicon.ico")
})

// Serve directory
app.Get("/static/*filepath", func(c *ginji.Context) {
    filepath := c.Param("filepath")
    c.File("./public" + filepath)
})
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
    app.Get("/", home)
    app.Get("/about", about)

    // API v1
    v1 := app.Group("/api/v1")
    v1.Use(middleware.RateLimitPerMinute(100))
    {
        // Public endpoints
        v1.Get("/status", status)
        v1.Post("/register", register)
        v1.Post("/login", login)

        // Protected endpoints
        protected := v1.Group("")
        protected.Use(middleware.BearerAuth(validateToken))
        {
            protected.Get("/profile", getProfile)
            protected.Put("/profile", updateProfile)
            protected.Get("/posts", getPosts)
            protected.Post("/posts", createPost)
        }

        // Admin endpoints
        admin := v1.Group("/admin")
        admin.Use(middleware.BearerAuth(validateToken))
        admin.Use(middleware.RequireRole("admin"))
        {
            admin.Get("/users", listAllUsers)
            admin.Delete("/users/:id", deleteUser)
        }
    }

    // Static files
    app.Get("/static/*filepath", serveStatic)

    app.Listen(":8080")
}
```

## Best Practices

1. **Use route groups** - Organize related routes
2. **Apply middleware at group level** - More efficient than per-route
3. **Use meaningful names** - Clear parameter names like `:id`, `:category`
4. **Version your APIs** - Use `/api/v1`, `/api/v2` prefixes
5. **Keep routes RESTful** - Follow REST conventions when possible
