# Quick Start

Get up and running with Ginji in 5 minutes.

## Installation

```bash
go get github.com/ginjigo/ginji
```

## Hello World

Create `main.go`:

```go
package main

import "github.com/ginjigo/ginji"

func main() {
    app := ginji.New()

    app.Get("/", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{
            "message": "Hello, World!",
        })
    })

    app.Listen(":3000")
}
```

Run it:

```bash
go run main.go
```

Visit http://localhost:3000

## Add Middleware

```go
import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Add middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.CORS(middleware.DefaultCORSOptions()))

    app.Get("/", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{"message": "Hello!"})
    })

    app.Listen(":3000")
}
```

## REST API

```go
type Todo struct {
    ID    int    `json:"id"`
    Title string `json:"title" ginji:"required,min=3"`
    Done  bool   `json:"done"`
}

var todos = []Todo{
    {ID: 1, Title: "Learn Ginji", Done: false},
}

func main() {
    app := ginji.New()
    app.Use(middleware.Logger())

    // List todos
    app.Get("/todos", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, todos)
    })

    // Get todo
    app.Get("/todos/:id", func(c *ginji.Context) {
        id := c.Param("id")
        c.JSON(ginji.StatusOK, ginji.H{"id": id})
    })

    // Create todo
    app.Post("/todos", func(c *ginji.Context) {
        var todo Todo
        if err := c.BindJSON(&todo); err != nil {
            return
        }
        todo.ID = len(todos) + 1
        todos = append(todos, todo)
        c.JSON(ginji.StatusCreated, todo)
    })

    app.Listen(":3000")
}
```

## What's Next?

- [Routing](/guide/routing) - Learn about route parameters and groups
- [Middleware](/middleware/) - Add security, rate limiting, auth
- [Validation](/guide/validation) - Validate request data
- [WebSocket](/realtime/websocket) - Add real-time features
