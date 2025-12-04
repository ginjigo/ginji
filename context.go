package ginji

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
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
	Req      *http.Request
	Res      http.ResponseWriter
	Params   map[string]string
	writer   *responseWriter
	Keys     map[string]any
	error    error         // error to be handled by error middleware
	written  bool          // whether response has been written
	aborted  bool          // whether request processing should stop
	services *ServiceScope // service scope for DI
	handlers []Handler     // middleware chain
	index    int8          // current handler index
}

// NewContext creates a new Context instance.
func NewContext(w http.ResponseWriter, r *http.Request, engine *Engine) *Context {
	writer := &responseWriter{ResponseWriter: w, status: 200}
	ctx := &Context{
		Req:      r,
		Res:      writer,
		Params:   make(map[string]string),
		writer:   writer,
		Keys:     make(map[string]any),
		written:  false,
		aborted:  false,
		index:    -1,
		handlers: nil,
	}

	// Initialize service scope if engine is provided
	if engine != nil {
		ctx.services = NewServiceScope(engine.container, ctx)
	}

	return ctx
}

// Reset resets the context for reuse.
func (c *Context) Reset(w http.ResponseWriter, r *http.Request, engine *Engine) {
	c.writer.ResponseWriter = w
	c.writer.status = 200
	c.writer.size = 0
	c.Req = r
	c.Res = c.writer
	c.Params = make(map[string]string)
	c.Keys = make(map[string]any)
	c.written = false
	c.aborted = false
	c.error = nil
	c.index = -1
	c.handlers = c.handlers[:0]

	// Dispose old service scope before creating new one to prevent memory leaks
	if c.services != nil {
		c.services.Dispose()
	}

	if engine != nil {
		c.services = NewServiceScope(engine.container, c)
	} else {
		c.services = nil
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

// BindValidate is a convenience method that binds and validates in one call.
// It automatically detects the content type and binds accordingly.
func (c *Context) BindValidate(v any) error {
	contentType := c.Header("Content-Type")

	// Handle JSON content type
	if strings.Contains(contentType, "application/json") {
		return c.BindJSON(v)
	}

	// Handle form data
	if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data") {
		if err := bindForm(c.Req, v); err != nil {
			return err
		}
		return validateStruct(v)
	}

	// Default to JSON
	return c.BindJSON(v)
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

// BindQuery binds query parameters to a struct and validates.
func (c *Context) BindQuery(v any) error {
	if err := bindMap(c.Req.URL.Query(), v, "query"); err != nil {
		return err
	}
	return validateStruct(v)
}

// BindHeader binds headers to a struct and validates.
func (c *Context) BindHeader(v any) error {
	if err := bindMap(c.Req.Header, v, "header"); err != nil {
		return err
	}
	return validateStruct(v)
}

// BindPath binds path parameters to a struct and validates.
func (c *Context) BindPath(v any) error {
	if err := bindParams(c.Params, v); err != nil {
		return err
	}
	return validateStruct(v)
}

// BindAll binds from all sources (path, query, header, body) and validates.
func (c *Context) BindAll(v any) error {
	// Bind path parameters first
	if err := bindParams(c.Params, v); err != nil {
		return err
	}

	// Bind query parameters
	if err := bindMap(c.Req.URL.Query(), v, "query"); err != nil {
		return err
	}

	// Bind headers
	if err := bindMap(c.Req.Header, v, "header"); err != nil {
		return err
	}

	// Bind body if present
	contentType := c.Header("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(c.Req.Body).Decode(v); err != nil {
			return err
		}
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data") {
		if err := bindForm(c.Req, v); err != nil {
			return err
		}
	}

	// Validate the combined result
	return validateStruct(v)
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
	c.Abort()
}

// AbortWithStatusJSON aborts the request with a JSON response.
func (c *Context) AbortWithStatusJSON(code int, data any) {
	c.aborted = true
	c.Status(code)
	_ = c.JSON(code, data)
	c.Abort()
}

// Abort aborts the request processing.
func (c *Context) Abort() {
	c.aborted = true
	c.index = int8(len(c.handlers) + 100) // Ensure we skip remaining handlers
}

// IsAborted returns whether the request has been aborted.
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Next calls the next middleware/handler in the chain.
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
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

// GetService resolves a service from the scoped container.
func (c *Context) GetService(name string) (any, error) {
	if c.services == nil {
		return nil, fmt.Errorf("service scope not initialized")
	}
	return c.services.container.Resolve(name, c.services)
}

// MustGetService resolves a service or panics if not found.
func (c *Context) MustGetService(name string) any {
	service, err := c.GetService(name)
	if err != nil {
		panic(err)
	}
	return service
}

// GetServiceTyped resolves a service with type safety.
func GetServiceTyped[T any](c *Context, name string) (T, error) {
	var zero T
	instance, err := c.GetService(name)
	if err != nil {
		return zero, err
	}

	service, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is not of type %T", name, zero)
	}

	return service, nil
}

// MustGetServiceTyped is like GetServiceTyped but panics on error.
func MustGetServiceTyped[T any](c *Context, name string) T {
	service, err := GetServiceTyped[T](c, name)
	if err != nil {
		panic(err)
	}
	return service
}
