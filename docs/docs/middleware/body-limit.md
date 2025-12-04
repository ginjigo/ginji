# Body Limit

Prevent DOS attacks by limiting request body size.

## Basic Usage

```go
import "github.com/ginjigo/ginji/middleware"

// Limit to 10MB
app.Use(middleware.BodyLimit(10 << 20))
```

## Helper Functions

```go
app.Use(middleware.BodyLimit1MB())   // 1 MB
app.Use(middleware.BodyLimit5MB())   // 5 MB
app.Use(middleware.BodyLimit10MB())  // 10 MB
app.Use(middleware.BodyLimit50MB())  // 50 MB
```

## Custom Configuration

```go
config := middleware.BodyLimitConfig{
    MaxBytes:     5 << 20, // 5 MB
    ErrorMessage: "File size exceeds limit",
    StatusCode:   ginji.StatusRequestEntityTooLarge,
}

app.Use(middleware.BodyLimitWithConfig(config))
```

## How It Works

1. **Content-Length Check** - Fast fail if header exceeds limit
2. **Stream Validation** - Validates actual body size during read
3. **Error Response** - Returns clear error when limit exceeded

## Error Response

```json
{
  "error": "Request body too large",
  "limit": "10485760 bytes",
  "maxSize": "10 MB"
}
```

## Use Cases

### File Uploads

```go
// Limit file uploads to 50MB
app.Post("/upload", func(c *ginji.Context) {
    file, _ := c.FormFile("file")
    c.SaveUploadedFile(file, "./uploads/"+file.Filename)
    c.JSON(ginji.StatusOK, ginji.H{"uploaded": file.Filename})
})

app.Use(middleware.BodyLimit50MB())
```

### API Endpoints

```go
// Different limits for different endpoints
api := app.Group("/api")
api.Use(middleware.BodyLimit1MB()) // Small JSON payloads

uploads := app.Group("/uploads")
uploads.Use(middleware.BodyLimit50MB()) // Large file uploads
```

## Best Practices

1. **Set appropriate limits** - Balance security and functionality
2. **Different limits per route** - Use route groups
3. **Inform users** - Document upload limits
4. **Use with timeouts** - Prevent slow uploads from hanging

## Example

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Global limit for all routes
    app.Use(middleware.BodyLimit5MB())

    // API with smaller limit
    api := app.Group("/api")
    api.Use(middleware.BodyLimit1MB())
    api.Post("/data", handleData)

    // File uploads with larger limit
    uploads := app.Group("/uploads")
    uploads.Use(middleware.BodyLimit50MB())
    uploads.Post("/file", handleFileUpload)

    app.Listen(":8080")
}
```
