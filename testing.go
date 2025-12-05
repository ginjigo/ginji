package ginji

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// TestContext creates a test context with the given request and response recorder.
func TestContext(w http.ResponseWriter, r *http.Request) *Context {
	if w == nil {
		w = httptest.NewRecorder()
	}
	if r == nil {
		r = httptest.NewRequest("GET", "/", nil)
	}
	return NewContext(w, r, nil)
}

// NewTestContext creates a new test context.
func NewTestContext(w http.ResponseWriter, r *http.Request) *Context {
	return NewContext(w, r, nil)
}

// NewTestContextWithRecorder creates a new test context with a response recorder.
func NewTestContextWithRecorder(method, path string) (*Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	return NewContext(w, r, nil), w
}

// PerformRequest simulates a request to the engine and returns the response recorder.
func PerformRequest(engine *Engine, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// PerformJSONRequest simulates a JSON request to the engine.
func PerformJSONRequest(engine *Engine, method, path string, payload any) *httptest.ResponseRecorder {
	jsonBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// PerformFormRequest simulates a form request to the engine.
func PerformFormRequest(engine *Engine, method, path string, formData url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// PerformMultipartRequest simulates a multipart/form-data request.
func PerformMultipartRequest(engine *Engine, method, path string, fields map[string]string, files map[string][]byte) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, val := range fields {
		_ = writer.WriteField(key, val)
	}

	// Add files
	for fieldname, content := range files {
		part, _ := writer.CreateFormFile(fieldname, fieldname)
		_, _ = part.Write(content)
	}

	_ = writer.Close()

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// Request is a fluent request builder for testing.
type Request struct {
	engine  *Engine
	method  string
	path    string
	headers map[string]string
	cookies []*http.Cookie
	body    io.Reader
}

// NewRequest creates a new request builder.
func NewRequest(engine *Engine, method, path string) *Request {
	return &Request{
		engine:  engine,
		method:  method,
		path:    path,
		headers: make(map[string]string),
	}
}

// Header sets a request header.
func (r *Request) Header(key, value string) *Request {
	r.headers[key] = value
	return r
}

// Cookie adds a cookie to the request.
func (r *Request) Cookie(cookie *http.Cookie) *Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

// JSON sets the request body to JSON.
func (r *Request) JSON(payload any) *Request {
	jsonBytes, _ := json.Marshal(payload)
	r.body = bytes.NewBuffer(jsonBytes)
	r.headers["Content-Type"] = "application/json"
	return r
}

// Body sets the request body.
func (r *Request) Body(body io.Reader) *Request {
	r.body = body
	return r
}

// Form sets the request body to URL-encoded form data.
func (r *Request) Form(formData url.Values) *Request {
	r.body = strings.NewReader(formData.Encode())
	r.headers["Content-Type"] = "application/x-www-form-urlencoded"
	return r
}

// Do executes the request and returns the response recorder.
func (r *Request) Do() *httptest.ResponseRecorder {
	req := httptest.NewRequest(r.method, r.path, r.body)

	// Set headers
	for key, val := range r.headers {
		req.Header.Set(key, val)
	}

	// Set cookies
	for _, cookie := range r.cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	r.engine.ServeHTTP(w, req)
	return w
}

// Response wraps httptest.ResponseRecorder with helper methods.
type Response struct {
	*httptest.ResponseRecorder
}

// JSON decodes the response body as JSON into v.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body.Bytes(), v)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return r.Body.String()
}

// Status returns the response status code.
func (r *Response) Status() int {
	return r.Code
}

// Header returns a response header value.
func (r *Response) Header(key string) string {
	return r.ResponseRecorder.Header().Get(key)
}

// NewResponse wraps a response recorder.
func NewResponse(w *httptest.ResponseRecorder) *Response {
	return &Response{ResponseRecorder: w}
}

// MockMiddleware creates a simple mock middleware for testing.
func MockMiddleware(key string, value any) Middleware {
	return func(c *Context) error {
		c.Set(key, value)
		return c.Next()
	}
}

// AssertStatus is a helper to assert response status.
func AssertStatus(t interface {
	Errorf(format string, args ...any)
}, w *httptest.ResponseRecorder, expected int) {
	if w.Code != expected {
		t.Errorf("Expected status %d, got %d", expected, w.Code)
	}
}

// AssertBody is a helper to assert response body contains a string.
func AssertBody(t interface {
	Errorf(format string, args ...any)
}, w *httptest.ResponseRecorder, expected string) {
	if !strings.Contains(w.Body.String(), expected) {
		t.Errorf("Expected body to contain '%s', got '%s'", expected, w.Body.String())
	}
}

// AssertHeader is a helper to assert response headers.
func AssertHeader(t interface {
	Errorf(format string, args ...any)
}, w *httptest.ResponseRecorder, key, expected string) {
	actual := w.Header().Get(key)
	if actual != expected {
		t.Errorf("Expected header %s to be '%s', got '%s'", key, expected, actual)
	}
}

// AssertJSON is a helper to assert JSON response matches expected.
func AssertJSON(t interface {
	Fatalf(format string, args ...any)
}, w *httptest.ResponseRecorder, expected any) {
	var actual any
	if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)

	if string(expectedJSON) != string(actualJSON) {
		t.Fatalf("Expected JSON:\n%s\nGot:\n%s", string(expectedJSON), string(actualJSON))
	}
}
