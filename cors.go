package ginji

import "net/http"

// CORSOptions defines the configuration for CORS middleware.
type CORSOptions struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// DefaultCORS returns a default CORS configuration.
func DefaultCORS() CORSOptions {
	return CORSOptions{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}
}

// CORS returns a middleware that handles CORS headers.
func CORS(options CORSOptions) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) {
			origin := c.Req.Header.Get("Origin")
			allowOrigin := ""

			// Check allowed origins
			for _, o := range options.AllowOrigins {
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
			}

			if allowOrigin != "" {
				c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
			}

			// Set other headers
			if len(options.AllowMethods) > 0 {
				c.SetHeader("Access-Control-Allow-Methods", join(options.AllowMethods, ", "))
			}
			if len(options.AllowHeaders) > 0 {
				c.SetHeader("Access-Control-Allow-Headers", join(options.AllowHeaders, ", "))
			}

			// Handle preflight request
			if c.Req.Method == "OPTIONS" {
				c.Status(http.StatusNoContent)
				return
			}

			next(c)
		}
	}
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	n := len(sep) * (len(strs) - 1)
	for i := 0; i < len(strs); i++ {
		n += len(strs[i])
	}

	b := make([]byte, n)
	bp := copy(b, strs[0])
	for _, s := range strs[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}
