package ginji

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions defines the configuration for CORS middleware.
type CORSOptions struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string // Headers that browsers are allowed to access
	AllowCredentials bool     // Whether to allow cookies/credentials
	MaxAge           int      // How long preflight requests can be cached (seconds)
}

// DefaultCORS returns a default CORS configuration.
func DefaultCORS() CORSOptions {
	return CORSOptions{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a middleware that handles CORS headers.
func CORS(options CORSOptions) Middleware {
	// Validate configuration for security issues
	if options.AllowCredentials && containsWildcard(options.AllowOrigins) {
		panic("CORS: cannot use credentials with wildcard origin '*'. This is a security vulnerability. Either disable credentials or specify explicit origins.")
	}

	return func(c *Context) error {
		origin := c.Req.Header.Get("Origin")

		// If no origin header, skip CORS (not a CORS request)
		if origin == "" {
			return c.Next()
		}

		// Check if origin is allowed
		if isOriginAllowed(origin, options.AllowOrigins) {
			// When credentials are enabled, we must return the specific origin, not "*"
			if options.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Origin", origin)
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			} else if len(options.AllowOrigins) == 1 && options.AllowOrigins[0] == "*" {
				// Only use wildcard when credentials are not enabled
				c.SetHeader("Access-Control-Allow-Origin", "*")
			} else {
				// Return specific origin
				c.SetHeader("Access-Control-Allow-Origin", origin)
			}
		} else {
			// Origin not allowed - don't set CORS headers, browser will block
			// Continue processing but don't set CORS headers
			return c.Next()
		}

		// Set allowed methods
		if len(options.AllowMethods) > 0 {
			c.SetHeader("Access-Control-Allow-Methods", strings.Join(options.AllowMethods, ", "))
		}

		// Set allowed headers
		if len(options.AllowHeaders) > 0 {
			c.SetHeader("Access-Control-Allow-Headers", strings.Join(options.AllowHeaders, ", "))
		}

		// Set exposed headers
		if len(options.ExposeHeaders) > 0 {
			c.SetHeader("Access-Control-Expose-Headers", strings.Join(options.ExposeHeaders, ", "))
		}

		// Set max age for preflight caching
		if options.MaxAge > 0 {
			c.SetHeader("Access-Control-Max-Age", strconv.Itoa(options.MaxAge))
		}

		// Handle preflight request
		if c.Req.Method == "OPTIONS" {
			c.Status(http.StatusNoContent)
			return nil
		}

		return c.Next()
	}
}

// isOriginAllowed checks if an origin is allowed, supporting wildcards.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard patterns like *.example.com
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// containsWildcard checks if the allowed origins list contains a wildcard.
func containsWildcard(allowedOrigins []string) bool {
	for _, origin := range allowedOrigins {
		if origin == "*" || strings.HasPrefix(origin, "*.") {
			return true
		}
	}
	return false
}
