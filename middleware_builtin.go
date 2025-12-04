package ginji

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"
)

// RequestID adds a unique ID to the request context and header.
func RequestID() Middleware {
	return func(c *Context) {
		id := generateRandomID()
		c.SetHeader("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
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
	return func(c *Context) {
		if !strings.Contains(c.Req.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		w := c.Res
		gz := gzip.NewWriter(w)
		defer func() {
			if err := gz.Close(); err != nil {
				log.Printf("Failed to close gzip writer: %v", err)
			}
		}()

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

		c.Next()

		// Restore original writer (not strictly necessary but good practice)
		c.Res = originalRes
	}
}
