package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ginjigo/ginji"
)

// bufferedResponseWriter buffers the response until we know if timeout occurred
type bufferedResponseWriter struct {
	header http.Header
	buf    *bytes.Buffer
	status int
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header: make(http.Header),
		buf:    new(bytes.Buffer),
		status: 200,
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

// copyTo copies the buffered response to the actual response writer
func (w *bufferedResponseWriter) copyTo(dst http.ResponseWriter) {
	// Copy headers
	for k, v := range w.header {
		for _, vv := range v {
			dst.Header().Add(k, vv)
		}
	}
	// Write status
	dst.WriteHeader(w.status)
	// Write body
	_, _ = dst.Write(w.buf.Bytes())
}

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

			// Replace response writer with buffered version
			originalRes := c.Res
			buffered := newBufferedResponseWriter()
			c.Res = buffered

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
				// Handler completed successfully - write buffered response
				buffered.copyTo(originalRes)
				return
			case <-ctx.Done():
				// Timeout occurred - write timeout response DIRECTLY to original writer
				// DO NOT restore c.Res - let handler continue writing to buffer which will be discarded
				if ctx.Err() == context.DeadlineExceeded {
					if !c.IsAborted() {
						// Write directly to original writer, bypassing the context
						originalRes.Header().Set("Content-Type", "application/json")
						originalRes.WriteHeader(config.StatusCode)
						jsonData, _ := json.Marshal(ginji.H{
							"error":   config.ErrorMessage,
							"timeout": config.Timeout.String(),
						})
						_, _ = originalRes.Write(jsonData)
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
