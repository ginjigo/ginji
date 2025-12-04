package ginji

import "strings"

// ConditionFunc is a function that determines if middleware should run.
type ConditionFunc func(*Context) bool

// ConditionalMiddleware wraps middleware with a condition.
type ConditionalMiddleware struct {
	condition  ConditionFunc
	middleware Middleware
}

// If creates conditional middleware that runs only if condition is true.
func If(condition ConditionFunc, middleware Middleware) Middleware {
	return func(c *Context) {
		if condition(c) {
			middleware(c)
		} else {
			c.Next()
		}
	}
}

// Unless creates conditional middleware that runs only if condition is false.
func Unless(condition ConditionFunc, middleware Middleware) Middleware {
	return func(c *Context) {
		if !condition(c) {
			middleware(c)
		} else {
			c.Next()
		}
	}
}

// PathMatches returns a condition that checks if path matches a pattern.
func PathMatches(pattern string) ConditionFunc {
	return func(c *Context) bool {
		return strings.HasPrefix(c.Req.URL.Path, pattern)
	}
}

// MethodIs returns a condition that checks if HTTP method matches.
func MethodIs(method string) ConditionFunc {
	return func(c *Context) bool {
		return c.Req.Method == method
	}
}

// HeaderEquals returns a condition that checks if header equals value.
func HeaderEquals(key, value string) ConditionFunc {
	return func(c *Context) bool {
		return c.Header(key) == value
	}
}

// HeaderExists returns a condition that checks if header exists.
func HeaderExists(key string) ConditionFunc {
	return func(c *Context) bool {
		return c.Header(key) != ""
	}
}

// And combines multiple conditions with AND logic.
func And(conditions ...ConditionFunc) ConditionFunc {
	return func(c *Context) bool {
		for _, cond := range conditions {
			if !cond(c) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple conditions with OR logic.
func Or(conditions ...ConditionFunc) ConditionFunc {
	return func(c *Context) bool {
		for _, cond := range conditions {
			if cond(c) {
				return true
			}
		}
		return false
	}
}

// Not negates a condition.
func Not(condition ConditionFunc) ConditionFunc {
	return func(c *Context) bool {
		return !condition(c)
	}
}

// Combine merges multiple middlewares into one.
// Note: This implementation dynamically inserts middlewares into the chain.
func Combine(middlewares ...Middleware) Middleware {
	return func(c *Context) {
		// We need to insert these middlewares into the current execution chain
		// immediately after the current index.

		// Create a new slice with enough capacity
		newHandlers := make([]Handler, 0, len(c.handlers)+len(middlewares))

		// Append handlers up to current index (inclusive)
		// Note: c.index points to the current handler (Combine's returned func)
		// We want to keep previous handlers and the current one (which is finishing)
		// But wait, if we are inside Combine's returned func, we are executing it.
		// We want the NEXT handlers to be [mw1, mw2, ... original_next...]

		// c.handlers[:c.index+1] includes the current handler.
		newHandlers = append(newHandlers, c.handlers[:c.index+1]...)

		// Append new middlewares
		newHandlers = append(newHandlers, middlewares...)

		// Append remaining handlers
		if int(c.index)+1 < len(c.handlers) {
			newHandlers = append(newHandlers, c.handlers[c.index+1:]...)
		}

		// Update context handlers
		c.handlers = newHandlers

		// Continue execution
		c.Next()
	}
}

// Skip creates middleware that skips if condition is true.
func Skip(condition ConditionFunc, middleware Middleware) Middleware {
	return Unless(condition, middleware)
}

// Only creates middleware that only runs if condition is true.
func Only(condition ConditionFunc, middleware Middleware) Middleware {
	return If(condition, middleware)
}
