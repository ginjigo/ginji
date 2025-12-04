package ginji

import (
	"net/http"
	"strings"
)

// Mode represents the application mode.
type Mode string

const (
	// DebugMode indicates development/debug mode.
	DebugMode Mode = "debug"
	// ReleaseMode indicates production/release mode.
	ReleaseMode Mode = "release"
	// TestMode indicates test mode.
	TestMode Mode = "test"
)

// mode is the current application mode.
var mode = DebugMode

// Engine is the main instance of the Ginji framework.
type Engine struct {
	*RouterGroup
	router *Router
	groups []*RouterGroup // store all groups
}

// RouterGroup defines a group of routes.
type RouterGroup struct {
	prefix      string
	middlewares []Middleware
	parent      *RouterGroup
	engine      *Engine
}

// New creates a new Ginji engine.
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group creates a new router group.
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Use adds middleware to the group.
func (group *RouterGroup) Use(middlewares ...Middleware) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// addRoute registers a route with the router.
func (group *RouterGroup) addRoute(method string, comp string, handler Handler) {
	pattern := group.prefix + comp
	group.engine.router.addRoute(method, pattern, handler)
}

// Get registers a GET request handler.
func (group *RouterGroup) Get(pattern string, handler Handler) {
	group.addRoute("GET", pattern, handler)
}

// Post registers a POST request handler.
func (group *RouterGroup) Post(pattern string, handler Handler) {
	group.addRoute("POST", pattern, handler)
}

// Put registers a PUT request handler.
func (group *RouterGroup) Put(pattern string, handler Handler) {
	group.addRoute("PUT", pattern, handler)
}

// Delete registers a DELETE request handler.
func (group *RouterGroup) Delete(pattern string, handler Handler) {
	group.addRoute("DELETE", pattern, handler)
}

// Patch registers a PATCH request handler.
func (group *RouterGroup) Patch(pattern string, handler Handler) {
	group.addRoute("PATCH", pattern, handler)
}

// Static registers a route to serve static files.
func (group *RouterGroup) Static(prefix, root string) {
	fs := http.StripPrefix(group.prefix+prefix, http.FileServer(http.Dir(root)))
	handler := func(c *Context) {
		fs.ServeHTTP(c.Res, c.Req)
	}
	// Register route for both the prefix and subpaths
	// Note: Trie router needs wildcard support for this to work perfectly for subpaths
	// My current router supports * wildcard.
	// So we register /prefix/*
	pattern := prefix + "/*filepath"
	group.addRoute("GET", pattern, handler)
}

// Run starts the HTTP server (alias for Listen).
func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

// Listen starts the HTTP server.
func (engine *Engine) Listen(addr string) error {
	return engine.Run(addr)
}

// ServeHTTP implements the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []Middleware
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := NewContext(w, req)

	// Define the final handler that executes the router logic
	finalHandler := func(c *Context) {
		engine.router.handle(c)
	}

	h := applyMiddleware(finalHandler, middlewares...)
	h(c)
}

// SetMode sets the application mode (debug, release, test).
func SetMode(m Mode) {
	mode = m
}

// GetMode returns the current application mode.
func GetMode() Mode {
	return mode
}
