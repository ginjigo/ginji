# Examples

Explore practical examples of building applications with Ginji.

## REST API

A complete CRUD API with validation and auth:

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name" ginji:"required,min=3"`
    Email string `json:"email" ginji:"required,email"`
}

var users = []User{
    {ID: 1, Name: "John Doe", Email: "john@example.com"},
}

func main() {
    app := ginji.New()
    
    // Middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.CORS(middleware.DefaultCORSOptions()))
    app.Use(ginji.DefaultErrorHandler())
    
    // Routes
    app.Get("/users", listUsers)
    app.Get("/users/:id", getUser)
    app.Post("/users", createUser)
    app.Put("/users/:id", updateUser)
    app.Delete("/users/:id", deleteUser)
    
    app.Listen(":8080")
}

func listUsers(c *ginji.Context) {
    c.JSON(ginji.StatusOK, users)
}

func getUser(c *ginji.Context) {
    id := c.Param("id")
    c.JSON(ginji.StatusOK, ginji.H{"id": id})
}

func createUser(c *ginji.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return
    }
    
    user.ID = len(users) + 1
    users = append(users, user)
    
    c.JSON(ginji.StatusCreated, user)
}

func updateUser(c *ginji.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return
    }
    
    c.JSON(ginji.StatusOK, user)
}

func deleteUser(c *ginji.Context) {
    c.JSON(ginji.StatusNoContent, nil)
}
```

## Chat Application

WebSocket-based chat with Hub:

```go
package main

import (
    "github.com/ginjigo/ginji"
)

func main() {
    app := ginji.New()
    hub := ginji.NewHub()
    go hub.Run()

    app.Get("/chat", func(c *ginji.Context) {
        upgrader := ginji.NewWebSocketUpgrader(ginji.DefaultWebSocketConfig())
        conn, _ := upgrader.Upgrade(c)
        
        hub.Register(conn)
        defer hub.Unregister(conn)
        
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                break
            }
            hub.Broadcast(msg)
        }
    })

    app.Listen(":8080")
}
```

## Live Dashboard

SSE for real-time updates:

```go
package main

import (
    "fmt"
    "time"
    "github.com/ginjigo/ginji"
)

var broadcaster = ginji.NewSSEBroadcaster()

func main() {
    app := ginji.New()

    app.Get("/dashboard", func(c *ginji.Context) {
        broadcaster.ServeSSE(c)
    })

    go sendUpdates()

    app.Listen(":8080")
}

func sendUpdates() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        broadcaster.Broadcast(ginji.SSEEvent{
            Event: "stats",
            Data:  fmt.Sprintf(`{"cpu": %d, "memory": %d}`, rand.Intn(100), rand.Intn(100)),
        })
    }
}
```

## Microservice

Production-ready microservice template:

```go
package main

import (
    "time"
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    ginji.SetMode(ginji.ReleaseMode)
    
    app := ginji.New()

    // Security
    app.Use(middleware.SecureStrict())
    app.Use(middleware.BodyLimit10MB())
    app.Use(middleware.CORS(middleware.DefaultCORSOptions()))

    // Operations
    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())
    app.Use(middleware.RequestID())
    app.Use(middleware.Health())

    // Performance
    app.Use(middleware.Timeout(30 * time.Second))
    app.Use(middleware.RateLimitPerMinute(1000))
    app.Use(middleware.Compress())

    // Public routes
    public := app.Group("/api/v1")
    public.Get("/status", status)

    // Protected routes
    protected := app.Group("/api/v1")
    protected.Use(middleware.BearerAuth(validateToken))
    protected.Get("/data", getData)
    protected.Post("/data", createData)

    app.Listen(":8080")
}

func status(c *ginji.Context) {
    c.JSON(ginji.StatusOK, ginji.H{"status": "ok"})
}

func getData(c *ginji.Context) {
    c.JSON(ginji.StatusOK, ginji.H{"data": []string{"item1", "item2"}})
}

func createData(c *ginji.Context) {
    c.JSON(ginji.StatusCreated, ginji.H{"created": true})
}

func validateToken(token string) (any, bool) {
    // Validate JWT token
    return ginji.H{"user_id": "123"}, true
}
```

## More Examples

Check the [examples directory](https://github.com/ginjigo/ginji/tree/main/examples) for more:

- File upload/download
- Authentication patterns
- Database integration
- Testing examples
- And more!
