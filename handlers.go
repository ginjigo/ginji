package ginji

// Handler is the function signature for route handlers.
// Handlers now return error for cleaner error handling.
type Handler func(*Context) error
