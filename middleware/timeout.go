package middleware

import (
	"context"
	"time"

	"github.com/ginjigo/ginji"
)

// TimeoutConfig defines the configuration for timeout middleware.
type TimeoutConfig struct {
	// Timeout is the duration before the request times out.
	Timeout time.Duration

	// ErrorMessage is the message returned when a timeout occurs.
	ErrorMessage string

	// StatusCode is the HTTP status code for timeout responses.
	// Default: 408 Request Timeout or 504 Gateway Timeout
	StatusCode int

	// SkipFunc allows skipping timeout for certain requests.
	SkipFunc func(*ginji.Context) bool
}

// DefaultTimeoutConfig returns default timeout configuration.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout:      30 * time.Second,
		ErrorMessage: "Request timeout",
		StatusCode:   ginji.StatusGatewayTimeout,
	}
}

// Timeout returns middleware that enforces a timeout on requests.
func Timeout(duration time.Duration) ginji.Middleware {
	config := DefaultTimeoutConfig()
	config.Timeout = duration
	return TimeoutWithConfig(config)
}

// TimeoutWithConfig returns middleware with custom timeout configuration.
func TimeoutWithConfig(config TimeoutConfig) ginji.Middleware {
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.StatusCode == 0 {
		config.StatusCode = ginji.StatusGatewayTimeout
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = "Request timeout"
	}

	return func(next ginji.Handler) ginji.Handler {
		return func(c *ginji.Context) {
			// Skip if skip function returns true
			if config.SkipFunc != nil && config.SkipFunc(c) {
				next(c)
				return
			}

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(c.Req.Context(), config.Timeout)
			defer cancel()

			// Replace request context
			c.Req = c.Req.WithContext(ctx)

			// Channel to signal completion
			done := make(chan struct{})

			// Run handler in goroutine
			go func() {
				defer func() {
					if r := recover(); r != nil {
						// If handler panics, let the recovery middleware handle it
						panic(r)
					}
				}()
				next(c)
				close(done)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Handler completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				if ctx.Err() == context.DeadlineExceeded {
					if !c.IsAborted() {
						c.AbortWithStatusJSON(config.StatusCode, ginji.H{
							"error":   config.ErrorMessage,
							"timeout": config.Timeout.String(),
						})
					}
				}
				return
			}
		}
	}
}

// TimeoutSeconds returns middleware with timeout in seconds.
func TimeoutSeconds(seconds int) ginji.Middleware {
	return Timeout(time.Duration(seconds) * time.Second)
}

// TimeoutMinutes returns middleware with timeout in minutes.
func TimeoutMinutes(minutes int) ginji.Middleware {
	return Timeout(time.Duration(minutes) * time.Minute)
}
