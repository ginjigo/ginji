package ginji

import (
	"reflect"
	"strconv"

	"github.com/ginjigo/schema"
)

// Route represents a chainable route for adding metadata.
type Route struct {
	engine       *Engine
	method       string
	pattern      string
	handler      Handler
	meta         *RouteMetadata
	middlewares  []Middleware
	bodySchema   *schema.Schema
	querySchema  *schema.Schema
	paramsSchema *schema.Schema
}

// RouteMetadata stores metadata about a route for documentation and validation.
type RouteMetadata struct {
	RequestType reflect.Type // Renamed from Request to RequestType
	Responses   map[string]reflect.Type
	Summary     string
	Description string // Added Description field
	Tags        []string
	OperationID string // Added OperationID field
	Deprecated  bool
}

// Summary sets the route summary.
func (r *Route) Summary(summary string) *Route {
	r.meta.Summary = summary
	return r
}

// Description sets the route description.
func (r *Route) Description(description string) *Route {
	r.meta.Description = description
	return r
}

// Tags sets the route tags.
func (r *Route) Tags(tags ...string) *Route {
	r.meta.Tags = tags
	return r
}

// OperationID sets the operation ID.
func (r *Route) OperationID(id string) *Route {
	r.meta.OperationID = id
	return r
}

// Request sets the request type for OpenAPI generation.
func (r *Route) Request(example interface{}) *Route {
	r.meta.RequestType = reflect.TypeOf(example)
	return r
}

// Response sets a response type for a status code.
func (r *Route) Response(code int, example interface{}) *Route {
	if r.meta.Responses == nil {
		r.meta.Responses = make(map[string]reflect.Type)
	}
	codeStr := strconv.Itoa(code)
	r.meta.Responses[codeStr] = reflect.TypeOf(example)
	return r
}

// Deprecated marks the route as deprecated.
func (r *Route) Deprecated() *Route {
	r.meta.Deprecated = true
	return r
}

// Middlewares adds middleware to this specific route.
func (r *Route) Middlewares(middlewares ...Middleware) *Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Body sets the request body schema for validation.
func (r *Route) Body(s *schema.Schema) *Route {
	r.bodySchema = s
	return r
}

// Query sets the query parameters schema for validation.
func (r *Route) Query(s *schema.Schema) *Route {
	r.querySchema = s
	return r
}

// Params sets the URL parameters schema for validation.
func (r *Route) Params(s *schema.Schema) *Route {
	r.paramsSchema = s
	return r
}

// build finalizes the route and adds it to the router.
func (r *Route) build() {
	// Add route to router
	r.engine.router.addRoute(r.method, r.pattern, r.handler)

	// Use consistent key for both metadata and middleware
	key := r.method + "-" + r.pattern

	// Store route-level middleware
	if len(r.middlewares) > 0 {
		r.engine.router.setRouteMiddleware(key, r.middlewares)
	}

	// Set metadata
	r.engine.router.setRouteMetadata(key, r.meta)
}
