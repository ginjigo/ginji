package ginji

import (
	"log"
	"time"
)

// Logger returns a middleware that logs HTTP requests.
func Logger() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) {
			t := time.Now()
			// Process request
			next(c)
			// Calculate resolution time
			log.Printf("[%d] %s %s in %v", c.StatusCode(), c.Req.Method, c.Req.URL.Path, time.Since(t))
		}
	}
}
