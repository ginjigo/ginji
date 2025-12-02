package ginji

import (
	"net/http"
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
	roots    map[string]*node
	handlers map[string]Handler
}

// newRouter creates a new Router instance.
func newRouter() *Router {
	return &Router{
		roots:    make(map[string]*node),
		handlers: make(map[string]Handler),
	}
}

// parsePattern splits a pattern into parts.
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			// Only break on wildcard if it's the start of a route definition part like *filepath
			// But we can't distinguish here easily.
			// However, for request paths, we MUST NOT break.
			// For route definitions, if user puts * in middle, it's their choice.
			if item[0] == '*' {
				// Check if it's a wildcard parameter (length > 1) or just a literal "*"
				// If it's just "*", it might be a valid part of a URL (though rare/discouraged)
				// If it's "*filepath", it's a wildcard param.
				// But wait, if I remove this break, then /files/*filepath/extra will be parsed as ["files", "*filepath", "extra"]
				// And inserted into trie.
				// But search stops at *filepath. So extra is unreachable. This is fine.

				// The original code broke on ANY item starting with *.
				// This breaks requests like /files/*/info.
				// So we remove the break.
			}
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

// handle executes the handler for a given context.
func (r *Router) handle(c *Context) {
	n, params := r.getRoute(c.Req.Method, c.Req.URL.Path)
	if n != nil {
		c.Params = params
		key := c.Req.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.Text(http.StatusNotFound, "404 NOT FOUND")
	}
}
