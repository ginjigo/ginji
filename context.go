package ginji

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Context wraps the HTTP request and response writer.
type Context struct {
	Req     *http.Request
	Res     http.ResponseWriter
	Params  map[string]string
	writer  *responseWriter
	Keys    map[string]any
	error   error // error to be handled by error middleware
	written bool  // whether response has been written
	aborted bool  // whether request processing should stop
}

// NewContext creates a new Context instance.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	writer := &responseWriter{ResponseWriter: w, status: 200}
	return &Context{
		Req:     r,
		Res:     writer,
		Params:  make(map[string]string),
		writer:  writer,
		Keys:    make(map[string]any),
		written: false,
		aborted: false,
	}
}

// Set stores a new key/value pair exclusively for this context.
func (c *Context) Set(key string, value any) {
	c.Keys[key] = value
}

// Get returns the value for the given key.
func (c *Context) Get(key string) (any, bool) {
	value, exists := c.Keys[key]
	return value, exists
}

// Param returns the value of a URL parameter.
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Status sets the HTTP status code.
func (c *Context) Status(code int) *Context {
	c.Res.WriteHeader(code)
	return c
}

// StatusCode returns the HTTP status code.
func (c *Context) StatusCode() int {
	return c.writer.status
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) *Context {
	c.Res.Header().Set(key, value)
	return c
}

// Query returns the query parameter value.
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Header returns the header value.
func (c *Context) Header(key string) string {
	return c.Req.Header.Get(key)
}

// BindJSON binds the request body to a struct and validates it.
func (c *Context) BindJSON(v any) error {
	if err := json.NewDecoder(c.Req.Body).Decode(v); err != nil {
		return err
	}
	return validateStruct(v)
}

// Send writes a byte slice to the response.
func (c *Context) Send(body []byte) error {
	c.written = true
	_, err := c.Res.Write(body)
	return err
}

// Text writes a string to the response with a status code.
func (c *Context) Text(code int, text string) error {
	c.Status(code)
	c.SetHeader("Content-Type", "text/plain")
	return c.Send([]byte(text))
}

// HTML writes an HTML string to the response with a status code.
func (c *Context) HTML(code int, html string) error {
	c.Status(code)
	c.SetHeader("Content-Type", "text/html")
	return c.Send([]byte(html))
}

// JSON writes a JSON object to the response with a status code.
func (c *Context) JSON(code int, v any) error {
	c.Status(code)
	c.SetHeader("Content-Type", "application/json")
	return json.NewEncoder(c.Res).Encode(v)
}

// BindQuery binds query parameters to a struct.
func (c *Context) BindQuery(v any) error {
	// Simple implementation: iterate over struct fields and get query params
	// This requires reflection similar to validateStruct, but for setting values
	// For now, let's just support basic string binding or leave it for a more complex binder
	// A full binder is complex. Let's stick to the plan but maybe implement a simple version or use a library?
	// The plan said "Add BindQuery". I should implement it.
	// Since we want zero dependencies, I have to write it.
	return bindMap(c.Req.URL.Query(), v, "query")
}

// BindHeader binds headers to a struct.
func (c *Context) BindHeader(v any) error {
	return bindMap(c.Req.Header, v, "header")
}

// Cookie returns the named cookie.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Req.Cookie(name)
}

// SetCookie sets a cookie.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Res, cookie)
}

// Redirect redirects the request to a new location.
func (c *Context) Redirect(code int, location string) error {
	http.Redirect(c.Res, c.Req, location, code)
	return nil
}

// FormValue returns the form value for the given key.
func (c *Context) FormValue(key string) string {
	return c.Req.FormValue(key)
}

// FormFile returns the file for the given key.
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	_, fileHeader, err := c.Req.FormFile(key)
	return fileHeader, err
}

// Error sets an error and marks the context for error handling.
func (c *Context) Error(err error) *Context {
	c.error = err
	return c
}

// AbortWithError aborts the request with an error.
func (c *Context) AbortWithError(code int, err error) {
	c.aborted = true
	var httpErr *HTTPError
	if he, ok := err.(*HTTPError); ok {
		httpErr = he
	} else {
		httpErr = NewHTTPError(code, err.Error())
	}
	handleError(c, httpErr)
}

// AbortWithStatusJSON aborts the request with a JSON response.
func (c *Context) AbortWithStatusJSON(code int, data any) {
	c.aborted = true
	c.Status(code)
	_ = c.JSON(code, data)
}

// Abort aborts the request processing.
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted returns whether the request has been aborted.
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Next calls the next middleware/handler in the chain.
// This is useful for fine-grained middleware control.
func (c *Context) Next() {
	// This will be implemented when we refactor middleware execution
	// to use an index-based approach similar to Gin
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) string {
	if val, ok := c.Keys[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt returns the value associated with the key as an int.
func (c *Context) GetInt(key string) int {
	if val, ok := c.Keys[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool returns the value associated with the key as a bool.
func (c *Context) GetBool(key string) bool {
	if val, ok := c.Keys[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// MustGet returns the value for the given key or panics if not found.
func (c *Context) MustGet(key string) any {
	if val, exists := c.Get(key); exists {
		return val
	}
	panic("key \"" + key + "\" does not exist")
}
