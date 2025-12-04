# API Reference

Complete reference for the Ginji API.

## Engine

### New

```go
app := ginji.New()
```

Creates a new Ginji application.

### HTTP Methods

```go
app.Get(path string, handlers ...Handler)
app.Post(path string, handlers ...Handler)
app.Put(path string, handlers ...Handler)
app.Delete(path string, handlers ...Handler)
app.Patch(path string, handlers ...Handler)
app.Options(path string, handlers ...Handler)
app.Head(path string, handlers ...Handler)
```

### Middleware

```go
app.Use(middleware ...Middleware)
```

### Route Groups

```go
group := app.Group(prefix string)
```

### Listen

```go
app.Listen(addr string) error
app.ListenTLS(addr, certFile, keyFile string) error
```

## Context

### Request

```go
c.Param(key string) string
c.Query(key string) string
c.Header(key string) string
c.FormValue(key string) string
c.Cookie(name string) (*http.Cookie, error)
```

### Binding

```go
c.BindJSON(obj any) error
c.BindQuery(obj any) error
c.BindHeader(obj any) error
```

### Response

```go
c.JSON(code int, obj any) error
c.Text(code int, text string) error
c.HTML(code int, html string) error
c.File(filepath string) error
c.Attachment(filepath, filename string) error
c.Redirect(code int, location string) error
c.Status(code int)
```

### Headers

```go
c.SetHeader(key, value string)
c.SetCookie(cookie *http.Cookie)
```

### Context Storage

```go
c.Set(key string, value any)
c.Get(key string) (any, bool)
c.GetString(key string) string
c.GetInt(key string) int
c.GetBool(key string) bool
c.MustGet(key string) any
```

### Error Handling

```go
c.AbortWithError(err error)
c.AbortWithStatusJSON(code int, obj any)
c.Error(err error) *Context
c.IsAborted() bool
```

### Streaming

```go
c.Stream(contentType string, reader io.Reader) error
c.FileStream(filepath string) error
c.ChunkedJSON(obj any) error
c.StreamJSON(items <-chan any) error
```

## Error Types

### HTTPError

```go
ginji.NewHTTPError(status int, message string) *HTTPError
err.WithDetails(details H) *HTTPError
```

### Common Errors

```go
ginji.NewBadRequestError(message string) *HTTPError
ginji.NewUnauthorizedError(message string) *HTTPError
ginji.NewForbiddenError(message string) *HTTPError
ginji.NewNotFoundError(message string) *HTTPError
ginji.NewInternalServerError(message string) *HTTPError
```

## Validation

```go
ginji.Validate(obj any) error
ginji.RegisterValidator(name string, fn ValidatorFunc)
```

## HTTP Status Codes

```go
ginji.StatusOK                  // 200
ginji.StatusCreated             // 201
ginji.StatusNoContent           // 204
ginji.StatusBadRequest          // 400
ginji.StatusUnauthorized        // 401
ginji.StatusForbidden           // 403
ginji.StatusNotFound            // 404
ginji.StatusMethodNotAllowed    // 405
ginji.StatusRequestTimeout      // 408
ginji.StatusConflict            // 409
ginji.StatusTooManyRequests     // 429
ginji.StatusInternalServerError // 500
ginji.StatusGatewayTimeout      // 504
```

## Testing

```go
ginji.NewTestContext() (*Context, *httptest.ResponseRecorder)
ginji.PerformRequest(app *Engine, method, path string, body any) *httptest.ResponseRecorder
ginji.PerformJSONRequest(app *Engine, method, path string, body any) *httptest.ResponseRecorder
ginji.NewRequest(app *Engine, method, path string) *RequestBuilder
ginji.AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int)
ginji.AssertBody(t *testing.T, rec *httptest.ResponseRecorder, expected string)
ginji.AssertHeader(t *testing.T, rec *httptest.ResponseRecorder, key, expected string)
```

## Middleware

See [Middleware Overview](/middleware/) for all built-in middleware.

## Real-Time

### WebSocket

```go
c.WebSocket(handler func(*WebSocketConn))
c.IsWebSocket() bool

conn.ReadMessage() (messageType int, data []byte, err error)
conn.WriteMessage(messageType int, data []byte) error
conn.ReadJSON(obj any) error
conn.WriteJSON(obj any) error
conn.Close() error
```

### SSE

```go
c.SSE(handler func(*SSEStream))

stream.Send(event SSEEvent) error
stream.SendData(data string) error
stream.SendJSON(obj any) error
stream.SendEvent(eventType, data string) error
```

## Types

```go
type Handler func(*Context)
type Middleware func(Handler) Handler
type H map[string]any
```
