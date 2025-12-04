# Getting Started

## Installation

```bash
go get github.com/ginjigo/ginji
```

## Quick Start

Create your first Ginji application in under 2 minutes:

```go
package main

import (
    "github.com/ginjigo/ginji"
)

func main() {
    app := ginji.New()

    app.Get("/", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{
            "message": "Hello, Ginji!",
        })
    })

    app.Listen(":3000")
}
```

Run your application:

```bash
go run main.go
```

Visit `http://localhost:3000` and you'll see:

```json
{
  "message": "Hello, Ginji!"
}
```

## Your First API

Let's build a simple TODO API:

```go
package main

import (
    "github.com/ginjigo/ginji"
)

type Todo struct {
    ID    int    `json:"id"`
    Title string `json:"title" ginji:"required,min=3"`
    Done  bool   `json:"done"`
}

var todos = []Todo{
    {ID: 1, Title: "Learn Ginji", Done: false},
    {ID: 2, Title: "Build an API", Done: false},
}

func main() {
    app := ginji.New()

    // Get all todos
    app.Get("/todos", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, todos)
    })

    // Get single todo
    app.Get("/todos/:id", func(c *ginji.Context) {
        id := c.Param("id")
        c.JSON(ginji.StatusOK, ginji.H{
            "id": id,
            "todo": "Todo details here",
        })
    })

    // Create todo
    app.Post("/todos", func(c *ginji.Context) {
        var todo Todo
        if err := c.BindJSON(&todo); err != nil {
            c.JSON(ginji.StatusBadRequest, ginji.H{
                "error": err.Error(),
            })
            return
        }

        todo.ID = len(todos) + 1
        todos = append(todos, todo)

        c.JSON(ginji.StatusCreated, todo)
    })

    app.Listen(":3000")
}
```

## Add Middleware

Enhance your app with built-in middleware:

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
    app.Use(middleware.RateLimitPerMinute(100))

    // Your routes here...

    app.Listen(":3000")
}
```

## Next Steps

- [Learn about Routing](/guide/routing)
- [Explore Middleware](/middleware/)
- [Add Error Handling](/guide/error-handling)
- [Implement Validation](/guide/validation)
- [Build Real-Time Features](/realtime/websocket)
