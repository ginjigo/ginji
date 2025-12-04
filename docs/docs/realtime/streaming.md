# Streaming

Efficiently stream files and data with Ginji's streaming capabilities.

## Stream from Reader

```go
app.Get("/stream", func(c *ginji.Context) {
    reader := getDataReader() // io.Reader
    c.Stream("application/octet-stream", reader)
})
```

## File Streaming

```go
app.Get("/video/:id", func(c *ginji.Context) {
    id := c.Param("id")
    c.FileStream("/videos/" + id + ".mp4")
})
```

## File Download

```go
app.Get("/download/:file", func(c *ginji.Context) {
    filename := c.Param("file")
    c.Attachment("/files/"+filename, filename)
})
```

## Chunked JSON

Stream large JSON responses:

```go
app.Get("/large-data", func(c *ginji.Context) {
    data := getLargeDataset()
    c.ChunkedJSON(data)
})
```

## JSON Array Streaming

Stream items one by one:

```go
app.Get("/stream-users", func(c *ginji.Context) {
    usersChan := make(chan any, 10)
    
    go func() {
        users := getAllUsers()
        for _, user := range users {
            usersChan <- user
        }
        close(usersChan)
    }()
    
    c.StreamJSON(usersChan)
})
```

## File Upload

```go
app.Post("/upload", func(c *ginji.Context) {
    file, _ := c.FormFile("file")
    
    // Save uploaded file
    c.SaveUploadedFile(file, "./uploads/"+file.Filename)
    
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

func main() {
    app := ginji.New()
    app.Use(middleware.BodyLimit50MB())

    // Stream video file
    app.Get("/video/:id", func(c *ginji.Context) {
        id := c.Param("id")
        c.FileStream("./videos/" + id + ".mp4")
    })

    // Download file
    app.Get("/download/:file", func(c *ginji.Context) {
        file := c.Param("file")
        c.Attachment("./files/"+file, file)
    })

    // Stream large dataset
    app.Get("/data", func(c *ginji.Context) {
        dataChan := make(chan any)
        
        go func() {
            for i := 0; i < 10000; i++ {
                dataChan <- ginji.H{"id": i, "value": i * 2}
            }
            close(dataChan)
        }()
        
        c.StreamJSON(dataChan)
    })

    // File upload
    app.Post("/upload", func(c *ginji.Context) {
        file, _ := c.FormFile("file")
        c.SaveUploadedFile(file, "./uploads/"+file.Filename)
        c.JSON(ginji.StatusOK, ginji.H{"uploaded": true})
    })

    app.Listen(":8080")
}
```

## Best Practices

1. **Use streaming for large files** - Don't buffer entire file in memory
2. **Set appropriate body limits** - Prevent DOS with `BodyLimit` middleware
3. **Handle errors** - Check for upload/download errors
4. **Use context** - Respect timeouts during streaming
5. **Content types** - Set correct Content-Type headers
