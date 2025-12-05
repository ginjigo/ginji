package ginji

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"
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

// ErrorHandler is a function that handles errors in the application.
type ErrorHandler func(*Context, error)

// Engine is the core of the framework.
type Engine struct {
	*RouterGroup
	router       *Router
	groups       []*RouterGroup // store all groups
	hooks        LifecycleHooks
	plugins      *PluginRegistry
	container    *Container   // DI container
	pool         sync.Pool    // context pool
	Logger       *slog.Logger // structured logger
	errorHandler ErrorHandler // custom error handler
}

// RouterGroup defines a group of routes.
type RouterGroup struct {
	prefix      string
	middlewares []Middleware
	parent      *RouterGroup
	engine      *Engine
}

// New creates a new Engine instance.
func New() *Engine {
	engine := &Engine{
		router:    newRouter(),
		hooks:     LifecycleHooks{},
		plugins:   newPluginRegistry(),
		container: NewContainer(),
	}

	// Initialize logger with appropriate handler based on mode
	var handler slog.Handler
	if mode == DebugMode {
		// JSON handler for debug mode with debug level
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		// JSON handler for production with info level
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	engine.Logger = slog.New(handler)

	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	engine.pool.New = func() any {
		return NewContext(nil, nil, engine)
	}
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
func (group *RouterGroup) Get(pattern string, handler Handler) *Route {
	fullPattern := group.prefix + pattern
	route := &Route{
		engine:  group.engine,
		method:  "GET",
		pattern: fullPattern,
		handler: handler,
		meta: &RouteMetadata{
			Responses: make(map[string]reflect.Type),
		},
	}
	route.build()
	return route
}

// Post registers a POST request handler.
func (group *RouterGroup) Post(pattern string, handler Handler) *Route {
	fullPattern := group.prefix + pattern
	route := &Route{
		engine:  group.engine,
		method:  "POST",
		pattern: fullPattern,
		handler: handler,
		meta: &RouteMetadata{
			Responses: make(map[string]reflect.Type),
		},
	}
	route.build()
	return route
}

// Put registers a PUT request handler.
func (group *RouterGroup) Put(pattern string, handler Handler) *Route {
	fullPattern := group.prefix + pattern
	route := &Route{
		engine:  group.engine,
		method:  "PUT",
		pattern: fullPattern,
		handler: handler,
		meta: &RouteMetadata{
			Responses: make(map[string]reflect.Type),
		},
	}
	route.build()
	return route
}

// Delete registers a DELETE request handler.
func (group *RouterGroup) Delete(pattern string, handler Handler) *Route {
	fullPattern := group.prefix + pattern
	route := &Route{
		engine:  group.engine,
		method:  "DELETE",
		pattern: fullPattern,
		handler: handler,
		meta: &RouteMetadata{
			Responses: make(map[string]reflect.Type),
		},
	}
	route.build()
	return route
}

// Patch registers a PATCH request handler.
func (group *RouterGroup) Patch(pattern string, handler Handler) *Route {
	fullPattern := group.prefix + pattern
	route := &Route{
		engine:  group.engine,
		method:  "PATCH",
		pattern: fullPattern,
		handler: handler,
		meta: &RouteMetadata{
			Responses: make(map[string]reflect.Type),
		},
	}
	route.build()
	return route
}

// Static registers a route to serve static files.
func (group *RouterGroup) Static(prefix, root string) {
	fs := http.StripPrefix(group.prefix+prefix, http.FileServer(http.Dir(root)))
	handler := func(c *Context) error {
		fs.ServeHTTP(c.Res, c.Req)
		return nil
	}
	// Register route for both the prefix and subpaths
	// Note: Trie router needs wildcard support for this to work perfectly for subpaths
	// My current router supports * wildcard.
	// So we register /prefix/*
	pattern := prefix + "/*filepath"
	group.addRoute("GET", pattern, handler)
}

// Typed creates a typed route builder for this router group.
// This avoids the limitation of Go not allowing generic methods.
func (group *RouterGroup) Typed() *TypedRouteBuilder {
	return &TypedRouteBuilder{group: group}
}

// TypedRouteBuilder provides type-safe route registration.
type TypedRouteBuilder struct {
	group *RouterGroup
}

// Get registers a type-safe GET request handler.
func (t *TypedRouteBuilder) Get(pattern string, handler any) *Route {
	return t.group.Get(pattern, wrapTypedHandler(handler))
}

// Post registers a type-safe POST request handler.
func (t *TypedRouteBuilder) Post(pattern string, handler any) *Route {
	return t.group.Post(pattern, wrapTypedHandler(handler))
}

// Put registers a type-safe PUT request handler.
func (t *TypedRouteBuilder) Put(pattern string, handler any) *Route {
	return t.group.Put(pattern, wrapTypedHandler(handler))
}

// Delete registers a type-safe DELETE request handler.
func (t *TypedRouteBuilder) Delete(pattern string, handler any) *Route {
	return t.group.Delete(pattern, wrapTypedHandler(handler))
}

// Patch registers a type-safe PATCH request handler.
func (t *TypedRouteBuilder) Patch(pattern string, handler any) *Route {
	return t.group.Patch(pattern, wrapTypedHandler(handler))
}

// wrapTypedHandler wraps any typed handler into a regular Handler.
// This uses reflection to detect and wrap the handler appropriately.
func wrapTypedHandler(handler any) Handler {
	handlerVal := reflect.ValueOf(handler)
	handlerType := handlerVal.Type()

	// Check if it's already a regular Handler
	if handlerType == reflect.TypeOf((Handler)(nil)) {
		return handler.(Handler)
	}

	// Handler should be a function with signature: func(*Context, Req) (Res, error)
	if handlerType.Kind() != reflect.Func {
		panic("handler must be a function")
	}

	if handlerType.NumIn() != 2 || handlerType.NumOut() != 2 {
		panic("handler must have signature: func(*Context, Req) (Res, error)")
	}

	// Extract request and response types
	reqType := handlerType.In(1)
	resType := handlerType.Out(0)

	return func(c *Context) error {
		// Create request value
		var reqVal reflect.Value
		isEmptyReq := reqType == reflect.TypeOf(EmptyRequest{})

		if !isEmptyReq {
			reqPtr := reflect.New(reqType)
			if err := bindTypedRequest(c, reqPtr.Interface()); err != nil {
				c.AbortWithError(StatusBadRequest, NewHTTPError(StatusBadRequest, "Invalid request: "+err.Error()))
				return nil
			}

			if err := validateStruct(reqPtr.Elem().Interface()); err != nil {
				c.AbortWithError(StatusBadRequest, err)
				return nil
			}

			reqVal = reqPtr.Elem()
		} else {
			reqVal = reflect.Zero(reqType)
		}

		// Call handler
		results := handlerVal.Call([]reflect.Value{reflect.ValueOf(c), reqVal})

		// Check error (second return value)
		if errInterface := results[1].Interface(); errInterface != nil {
			if err, ok := errInterface.(error); ok {
				if httpErr, ok := err.(*HTTPError); ok {
					c.AbortWithError(httpErr.Code, httpErr)
				} else {
					c.AbortWithError(StatusInternalServerError, err)
				}
				return nil
			}
		}

		// Handle response (first return value)
		isEmptyRes := resType == reflect.TypeOf(EmptyRequest{})
		if !isEmptyRes {
			res := results[0].Interface()
			_ = c.JSON(StatusOK, res)
		} else {
			if c.StatusCode() == StatusOK {
				c.Status(StatusNoContent)
			}
		}

		return nil
	}
}

// Run starts the HTTP server (alias for Listen).
func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

// Listen starts the HTTP server.
func (engine *Engine) Listen(addr string) error {
	return engine.Run(addr)
}

// ListenTLS starts the HTTPS server.
func (engine *Engine) ListenTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, engine)
}

// ListenWithShutdown starts the HTTP server with graceful shutdown support.
// It listens for SIGINT/SIGTERM signals and gracefully shuts down the server
// with the specified timeout.
func (engine *Engine) ListenWithShutdown(addr string, timeout time.Duration) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		engine.Logger.Info("Server starting", slog.String("addr", addr))
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return err

	case sig := <-shutdown:
		engine.Logger.Info("Received shutdown signal", slog.String("signal", sig.String()))

		// Stop plugins first to allow them to clean up resources
		if err := engine.StopPlugins(); err != nil {
			engine.Logger.Error("Failed to stop plugins", slog.String("error", err.Error()))
		}

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			// Force close after timeout
			engine.Logger.Error("Graceful shutdown failed, forcing close", slog.String("error", err.Error()))
			if closeErr := srv.Close(); closeErr != nil {
				return closeErr
			}
			return err
		}

		engine.Logger.Info("Server gracefully stopped")
		return nil
	}
}

// ListenTLSWithShutdown starts the HTTPS server with graceful shutdown support.
func (engine *Engine) ListenTLSWithShutdown(addr, certFile, keyFile string, timeout time.Duration) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		engine.Logger.Info("HTTPS server starting", slog.String("addr", addr))
		serverErrors <- srv.ListenAndServeTLS(certFile, keyFile)
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return err

	case sig := <-shutdown:
		engine.Logger.Info("Received shutdown signal", slog.String("signal", sig.String()))

		// Stop plugins first to allow them to clean up resources
		if err := engine.StopPlugins(); err != nil {
			engine.Logger.Error("Failed to stop plugins", slog.String("error", err.Error()))
		}

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			// Force close after timeout
			engine.Logger.Error("Graceful shutdown failed, forcing close", slog.String("error", err.Error()))
			if closeErr := srv.Close(); closeErr != nil {
				return closeErr
			}
			return err
		}

		engine.Logger.Info("HTTPS server gracefully stopped")
		return nil
	}
}

// ServeHTTP makes the router implement the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.Reset(w, req, engine)

	// Add system middleware to handle OnResponse hooks
	// This must be the first handler in the chain to ensure it runs last on the way back
	c.handlers = append(c.handlers, func(c *Context) error {
		err := c.Next()
		engine.executeOnResponse(c)
		return err
	})

	// Collect all middleware
	// Note: In a real high-perf scenario, we should pre-calculate this or optimize it
	for _, group := range engine.groups {
		if len(group.prefix) == 0 || (len(req.URL.Path) >= len(group.prefix) && req.URL.Path[:len(group.prefix)] == group.prefix) {
			for _, mw := range group.middlewares {
				c.handlers = append(c.handlers, Handler(mw))
			}
		}
	}

	// Dispatch to router to find route handlers
	engine.router.handle(c, engine)

	// Execute the chain
	c.Next()

	// Return to pool
	engine.pool.Put(c)
}

// SetMode sets the application mode (debug, release, test).
func SetMode(m Mode) {
	mode = m
}

// GetMode returns the current application mode.
func GetMode() Mode {
	return mode
}

// RegisterService registers a service with the DI container.
func (e *Engine) RegisterService(name string, factory any, lifetime ServiceLifetime) error {
	return e.container.Register(name, factory, lifetime)
}

// RegisterSingleton registers a singleton service.
func (e *Engine) RegisterSingleton(name string, factory any) error {
	return e.container.RegisterSingleton(name, factory)
}

// RegisterTransient registers a transient service.
func (e *Engine) RegisterTransient(name string, factory any) error {
	return e.container.RegisterTransient(name, factory)
}

// RegisterScoped registers a scoped service.
func (e *Engine) RegisterScoped(name string, factory any) error {
	return e.container.RegisterScoped(name, factory)
}

// RegisterInstance registers a pre-created instance as a singleton.
func (e *Engine) RegisterInstance(name string, instance any) error {
	return e.container.RegisterInstance(name, instance)
}

// Container returns the DI container.
func (e *Engine) Container() *Container {
	return e.container
}

// SetErrorHandler sets a custom error handler for the application.
// The custom handler will be called instead of the default error handler.
func (e *Engine) SetErrorHandler(handler ErrorHandler) {
	e.errorHandler = handler
}

// GetErrorHandler returns the current error handler (custom or default).
func (e *Engine) GetErrorHandler() ErrorHandler {
	if e.errorHandler != nil {
		return e.errorHandler
	}
	return defaultErrorHandler
}
