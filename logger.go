package ginji

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LoggerConfig defines configuration options for the Logger middleware.
type LoggerConfig struct {
	Output       io.Writer                             // Where to write logs (default: os.Stdout)
	Format       string                                // "json" or "text" (default: "text")
	SkipPaths    []string                              // Paths to skip logging
	CustomFields func(*Context) map[string]interface{} // Function to add custom fields
}

// DefaultLoggerConfig returns default logger configuration.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Output:    os.Stdout,
		Format:    "text",
		SkipPaths: []string{},
	}
}

// Logger returns a middleware that logs HTTP requests with default configuration.
func Logger() Middleware {
	return LoggerWithConfig(DefaultLoggerConfig())
}

// LoggerWithConfig returns a middleware that logs HTTP requests with custom configuration.
func LoggerWithConfig(config LoggerConfig) Middleware {
	// Set defaults
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.Format == "" {
		config.Format = "text"
	}

	return func(c *Context) {
		start := time.Now()
		path := c.Req.URL.Path
		query := c.Req.URL.RawQuery

		// Skip if path is in skip list
		for _, skip := range config.SkipPaths {
			if path == skip {
				c.Next()
				return
			}
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log entry
		logEntry := map[string]interface{}{
			"timestamp":  start.Format(time.RFC3339),
			"method":     c.Req.Method,
			"path":       path,
			"status":     c.StatusCode(),
			"latency_ms": latency.Milliseconds(),
			"client_ip":  c.Req.RemoteAddr,
			"user_agent": c.Req.UserAgent(),
		}

		// Add query if present
		if query != "" {
			logEntry["query"] = query
		}

		// Add request ID if present
		if reqID, exists := c.Get("request_id"); exists {
			logEntry["request_id"] = reqID
		}

		// Add custom fields if provided
		if config.CustomFields != nil {
			for k, v := range config.CustomFields(c) {
				logEntry[k] = v
			}
		}

		// Output based on format
		if config.Format == "json" {
			if err := json.NewEncoder(config.Output).Encode(logEntry); err != nil {
				log.Printf("Failed to encode log entry: %v", err)
			}
		} else {
			// Text format
			statusColor := getStatusColor(c.StatusCode())
			methodColor := getMethodColor(c.Req.Method)

			var logLine string
			if reqID, ok := logEntry["request_id"].(string); ok {
				logLine = fmt.Sprintf("[%s] %s%3d\033[0m | %13v | %15s | %s%-7s\033[0m | %s",
					logEntry["timestamp"],
					statusColor, c.StatusCode(),
					latency,
					logEntry["client_ip"],
					methodColor, c.Req.Method,
					path,
				)
				if query != "" {
					logLine += "?" + query
				}
				logLine += fmt.Sprintf(" | ID: %s", reqID)
			} else {
				logLine = fmt.Sprintf("[%s] %s%3d\033[0m | %13v | %15s | %s%-7s\033[0m | %s",
					logEntry["timestamp"],
					statusColor, c.StatusCode(),
					latency,
					logEntry["client_ip"],
					methodColor, c.Req.Method,
					path,
				)
				if query != "" {
					logLine += "?" + query
				}
			}

			fmt.Fprintln(config.Output, logLine)
		}
	}
}

// getStatusColor returns ANSI color code based on HTTP status.
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\033[32m" // Green
	case status >= 300 && status < 400:
		return "\033[36m" // Cyan
	case status >= 400 && status < 500:
		return "\033[33m" // Yellow
	default:
		return "\033[31m" // Red
	}
}

// getMethodColor returns ANSI color code based on HTTP method.
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // Blue
	case "POST":
		return "\033[32m" // Green
	case "PUT":
		return "\033[33m" // Yellow
	case "DELETE":
		return "\033[31m" // Red
	case "PATCH":
		return "\033[35m" // Magenta
	default:
		return "\033[37m" // White
	}
}
