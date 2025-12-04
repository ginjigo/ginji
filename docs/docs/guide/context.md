# Context

The Context is the heart of Ginji, providing access to request/response and utilities.

## Request Data

### Query Parameters

```go
app.Get("/search", func(c *ginji.Context) {
    query := c.Query("q")
    page := c.Query("page")
    
    c.JSON(ginji.StatusOK, ginji.H{
        "query": query,
        "page": page,
    })
})
```

### Route Parameters

```go
app.Get("/users/:id", func(c *ginji.Context) {
    id := c.Param("id")
    c.JSON(ginji.StatusOK, ginji.H{"id": id})
})
```

### Headers

```go
app.Get("/", func(c *ginji.Context) {
    userAgent := c.Header("User-Agent")
    authorization := c.Header("Authorization")
    
    c.JSON(ginji.StatusOK, ginji.H{
        "userAgent": userAgent,
        "auth": authorization,
    })
})
```

### Request Body

#### JSON

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

app.Post("/users", func(c *ginji.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        return
    }
    
    c.JSON(ginji.StatusCreated, user)
})
```

#### Form Data

```go
app.Post("/form", func(c *ginji.Context) {
    name := c.FormValue("name")
    email := c.FormValue("email")
    
    c.JSON(ginji.StatusOK, ginji.H{
        "name": name,
        "email": email,
    })
})
```

#### Query Binding

```go
type SearchParams struct {
    Query string `query:"q"`
    Page  int    `query:"page"`
    Limit int    `query:"limit"`
}

app.Get("/search", func(c *ginji.Context) {
    var params SearchParams
    if err := c.BindQuery(&params); err != nil {
        c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
        return
    }
    
    c.JSON(ginji.StatusOK, params)
})
```

## Response

### JSON

```go
app.Get("/json", func(c *ginji.Context) {
    c.JSON(ginji.StatusOK, ginji.H{
        "message": "Hello!",
        "status": "success",
    })
})
```

### Text

```go
app.Get("/text", func(c *ginji.Context) {
    c.Text(ginji.StatusOK, "Hello, World!")
})
```

### HTML

```go
app.Get("/html", func(c *ginji.Context) {
    html := "<h1>Hello, World!</h1>"
    c.HTML(ginji.StatusOK, html)
})
```

### File

```go
app.Get("/file", func(c *ginji.Context) {
    c.File("./public/document.pdf")
})
```

### Download

```go
app.Get("/download", func(c *ginji.Context) {
    c.Attachment("./files/report.pdf", "monthly-report.pdf")
})
```

### Redirect

```go
app.Get("/redirect", func(c *ginji.Context) {
    c.Redirect(ginji.StatusMovedPermanently, "/new-location")
})
```

## Headers

### Set Header

```go
app.Get("/", func(c *ginji.Context) {
    c.SetHeader("X-Custom-Header", "value")
    c.SetHeader("Cache-Control", "no-cache")
    
    c.JSON(ginji.StatusOK, ginji.H{"message": "ok"})
})
```

### Cookies

```go
app.Get("/set-cookie", func(c *ginji.Context) {
    cookie := &http.Cookie{
        Name:  "session",
        Value: "abc123",
        Path:  "/",
        MaxAge: 3600,
    }
    c.SetCookie(cookie)
    
    c.JSON(ginji.StatusOK, ginji.H{"message": "Cookie set"})
})

app.Get("/get-cookie", func(c *ginji.Context) {
    cookie, err := c.Cookie("session")
    if err != nil {
        c.JSON(ginji.StatusOK, ginji.H{"session": "none"})
        return
    }
    
    c.JSON(ginji.StatusOK, ginji.H{"session": cookie.Value})
})
```

## Context Storage

Store and retrieve values in the context:

```go
app.Use(func(next ginji.Handler) ginji.Handler {
    return func(c *ginji.Context) {
        // Set value
        c.Set("user_id", "123")
        c.Set("role", "admin")
        
        next(c)
    }
})

app.Get("/profile", func(c *ginji.Context) {
    // Get value
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(ginji.StatusUnauthorized, ginji.H{"error": "Not authenticated"})
        return
    }
    
    // Typed getters
    role := c.GetString("role")
    
    c.JSON(ginji.StatusOK, ginji.H{
        "user_id": userID,
        "role": role,
    })
})
```

### Typed Getters

```go
userID := c.GetString("user_id")
count := c.GetInt("count")
active := c.GetBool("active")
```

### Must Get (panics if not found)

```go
userID := c.MustGet("user_id").(string)
```

## Error Handling

### Abort with Error

```go
app.Get("/users/:id", func(c *ginji.Context) {
    id := c.Param("id")
    
    user, err := findUser(id)
    if err != nil {
        c.AbortWithError(ginji.NewNotFoundError("User not found"))
        return
    }
    
    c.JSON(ginji.StatusOK, user)
})
```

### Abort with Status and JSON

```go
app.Post("/users", func(c *ginji.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.AbortWithStatusJSON(ginji.StatusBadRequest, ginji.H{
            "error": "Invalid request",
        })
        return
    }
    
    c.JSON(ginji.StatusCreated, user)
})
```

## File Upload

```go
app.Post("/upload", func(c *ginji.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(ginji.StatusBadRequest, ginji.H{"error": "No file"})
        return
    }
    
    // Save file
    err = c.SaveUploadedFile(file, "./uploads/"+file.Filename)
    if err != nil {
        c.JSON(ginji.StatusInternalServerError, ginji.H{"error": "Upload failed"})
        return
    }
    
    c.JSON(ginji.StatusOK, ginji.H{
        "filename": file.Filename,
        "size": file.Size,
    })
})
```

## Complete Example

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

type CreateUserRequest struct {
    Name  string `json:"name" ginji:"required,min=3"`
    Email string `json:"email" ginji:"required,email"`
}

func main() {
    app := ginji.New()
    app.Use(middleware.Logger())
    app.Use(ginji.DefaultErrorHandler())

    app.Get("/", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{"message": "API running"})
    })

    app.Post("/users", func(c *ginji.Context) {
        var req CreateUserRequest
        if err := c.BindJSON(&req); err != nil {
            return
        }
        
        user := createUser(req.Name, req.Email)
        c.JSON(ginji.StatusCreated, user)
    })

    app.Get("/users/:id", func(c *ginji.Context) {
        id := c.Param("id")
        user, err := findUser(id)
        
        if err != nil {
            c.AbortWithError(ginji.NewNotFoundError("User not found"))
            return
        }
        
        c.JSON(ginji.StatusOK, user)
    })

    app.Listen(":8080")
}
```

## Next Steps

- [Routing](/guide/routing) - Advanced routing patterns
- [Error Handling](/guide/error-handling) - Handle errors gracefully
- [Validation](/guide/validation) - Validate request data
