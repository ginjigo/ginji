package ginji

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

// RequestID adds a unique ID to the request context and header.
func RequestID() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) {
			id := generateRandomID()
			c.SetHeader("X-Request-ID", id)
			c.Set("request_id", id)
			next(c)
		}
	}
}

func generateRandomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(bytes)
}

// gzipResponseWriter wraps the http.ResponseWriter to support gzip compression.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Compress enables Gzip compression for responses.
func Compress() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) {
			if !strings.Contains(c.Req.Header.Get("Accept-Encoding"), "gzip") {
				next(c)
				return
			}

			w := c.Res
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Wrap the response writer
			gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}

			// We need to hack the context to use our new writer
			// But Context uses a custom responseWriter.
			// We should probably update Context to allow swapping the writer or just wrap it here.
			// The Context struct has `Res http.ResponseWriter`. We can update that.
			originalRes := c.Res
			c.Res = gzw
			c.SetHeader("Content-Encoding", "gzip")
			c.SetHeader("Vary", "Accept-Encoding")

			next(c)

			// Restore original writer (not strictly necessary but good practice)
			c.Res = originalRes
		}
	}
}
