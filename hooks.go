package ginji

// HookFunc represents a lifecycle hook function.
type HookFunc func(*Context)

// LifecycleHooks stores application lifecycle hooks.
type LifecycleHooks struct {
	onRequest  []HookFunc // Before routing
	onRoute    []HookFunc // After route match, before handler
	onResponse []HookFunc // After handler execution
	onError    []HookFunc // On error
}

// OnRequest registers a hook that runs before routing.
func (e *Engine) OnRequest(hook HookFunc) {
	e.hooks.onRequest = append(e.hooks.onRequest, hook)
}

// OnRoute registers a hook that runs after route matching.
func (e *Engine) OnRoute(hook HookFunc) {
	e.hooks.onRoute = append(e.hooks.onRoute, hook)
}

// OnResponse registers a hook that runs after handler execution.
func (e *Engine) OnResponse(hook HookFunc) {
	e.hooks.onResponse = append(e.hooks.onResponse, hook)
}

// OnError registers a hook that runs when an error occurs.
func (e *Engine) OnError(hook HookFunc) {
	e.hooks.onError = append(e.hooks.onError, hook)
}

// executeOnRequest runs all OnRequest hooks.
func (e *Engine) executeOnRequest(c *Context) {
	for _, hook := range e.hooks.onRequest {
		hook(c)
		if c.aborted {
			return
		}
	}
}

// executeOnRoute runs all OnRoute hooks.
func (e *Engine) executeOnRoute(c *Context) {
	for _, hook := range e.hooks.onRoute {
		hook(c)
		if c.aborted {
			return
		}
	}
}

// executeOnResponse runs all OnResponse hooks.
func (e *Engine) executeOnResponse(c *Context) {
	for _, hook := range e.hooks.onResponse {
		hook(c)
	}
}
