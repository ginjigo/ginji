package ginji

// Middleware is a function that wraps a handler.
// Middleware now returns error for consistent error handling.
type Middleware func(*Context) error
