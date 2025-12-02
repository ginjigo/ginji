package ginji

// Middleware defines a function that wraps a Handler.
type Middleware func(Handler) Handler

// applyMiddleware applies a list of middlewares to a handler.
// The middlewares are applied in reverse order so that the first middleware
// in the list is the first one to be executed.
func applyMiddleware(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
