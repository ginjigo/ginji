package ginji

import (
	"net/http"
	"reflect"
	"strings"
)

// node represents a node in the routing trie.
type node struct {
	pattern  string
	part     string
	children []*node
	isWild   bool
}

// insert inserts a new pattern into the trie.
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// search searches for a node matching the parts.
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			// If we reached the end of parts but no pattern, check if we have a wildcard child
			// This allows /assets/ to match /assets/*filepath
			for _, child := range n.children {
				if child.isWild && strings.HasPrefix(child.part, "*") {
					return child
				}
			}
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// getRouteMetadata returns metadata for a route.
func (r *Router) getRouteMetadata(key string) *RouteMetadata {
	if meta, ok := r.metadata[key]; ok {
		return meta
	}
	return &RouteMetadata{
		Responses: make(map[string]reflect.Type),
	}
}

// setRouteMetadata sets metadata for a route.
func (r *Router) setRouteMetadata(key string, meta *RouteMetadata) {
	if meta.Responses == nil {
		meta.Responses = make(map[string]reflect.Type)
	}
	r.metadata[key] = meta
}

// setRouteMiddleware stores middleware for a specific route.
func (r *Router) setRouteMiddleware(key string, middlewares []Middleware) {
	r.routeMiddleware[key] = middlewares
}

// getRouteMiddleware retrieves middleware for a specific route.
func (r *Router) getRouteMiddleware(key string) []Middleware {
	return r.routeMiddleware[key]
}

func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// Router handles request routing.
type Router struct {
	roots           map[string]*node
	handlers        map[string]Handler
	metadata        map[string]*RouteMetadata
	routeMiddleware map[string][]Middleware
}

// newRouter creates a new Router instance.
func newRouter() *Router {
	return &Router{
		roots:           make(map[string]*node),
		handlers:        make(map[string]Handler),
		metadata:        make(map[string]*RouteMetadata),
		routeMiddleware: make(map[string][]Middleware),
	}
}

// parsePattern splits a pattern into parts.
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
		}
	}
	return parts
}

// addRoute adds a route to the router.
func (r *Router) addRoute(method string, pattern string, handler Handler) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// getRoute resolves a route and extracts parameters.
func (r *Router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

// handle dispatches the request to the matched handler.
func (r *Router) handle(c *Context, engine *Engine) {
	// Execute OnRequest hooks
	if engine != nil {
		engine.executeOnRequest(c)
		if c.aborted {
			return
		}
	}

	n, params := r.getRoute(c.Req.Method, c.Req.URL.Path)
	if n != nil {
		c.Params = params

		// Execute OnRoute hooks
		if engine != nil {
			engine.executeOnRoute(c)
			if c.aborted {
				return
			}
		}

		key := c.Req.Method + "-" + n.pattern
		handler := r.handlers[key]

		// Get route-specific middleware
		routeMW := r.getRouteMiddleware(key)

		// Append route middleware
		for _, mw := range routeMW {
			c.handlers = append(c.handlers, Handler(mw))
		}

		// Append final handler
		c.handlers = append(c.handlers, handler)

		// Note: OnResponse hooks are executed in ginji.go ServeHTTP as part of the middleware chain.
		// The first middleware added wraps c.Next() to execute hooks after all handlers complete.

	} else {
		c.handlers = append(c.handlers, func(c *Context) error {
			return c.Text(http.StatusNotFound, "404 NOT FOUND")
		})
	}
}
