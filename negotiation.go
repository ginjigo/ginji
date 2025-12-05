package ginji

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// NegotiateFormat represents content negotiation handlers.
type NegotiateFormat struct {
	JSON func() error
	XML  func() error
	HTML func() error
	Text func() error
}

// Negotiate performs content negotiation based on Accept header.
func (c *Context) Negotiate(code int, data interface{}, formats NegotiateFormat) error {
	accept := c.Header("Accept")

	// Determine preferred content type
	switch {
	case strings.Contains(accept, "application/json") || accept == "*/*" || accept == "":
		if formats.JSON != nil {
			return formats.JSON()
		}
		return c.JSON(code, data)

	case strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml"):
		if formats.XML != nil {
			return formats.XML()
		}
		return c.Text(code, "XML not supported")

	case strings.Contains(accept, "text/html"):
		if formats.HTML != nil {
			return formats.HTML()
		}
		return c.Text(code, fmt.Sprintf("%v", data))

	case strings.Contains(accept, "text/plain"):
		if formats.Text != nil {
			return formats.Text()
		}
		return c.Text(code, fmt.Sprintf("%v", data))

	default:
		// Default to JSON
		if formats.JSON != nil {
			return formats.JSON()
		}
		return c.JSON(code, data)
	}
}

// CacheConfig represents cache configuration.
type CacheConfig struct {
	MaxAge         time.Duration
	SMaxAge        time.Duration
	MustRevalidate bool
	NoStore        bool
	NoCache        bool
	Public         bool
	Private        bool
	NoTransform    bool
}

// Cache sets cache-control headers and returns the context for chaining.
func (c *Context) Cache(maxAge time.Duration) *CacheContext {
	return &CacheContext{
		ctx: c,
		config: CacheConfig{
			MaxAge: maxAge,
		},
	}
}

// CacheContext wraps Context for cache operations.
type CacheContext struct {
	ctx    *Context
	config CacheConfig
}

// Public marks the response as publicly cacheable.
func (cc *CacheContext) Public() *CacheContext {
	cc.config.Public = true
	return cc
}

// Private marks the response as privately cacheable.
func (cc *CacheContext) Private() *CacheContext {
	cc.config.Private = true
	return cc
}

// NoStore prevents caching entirely.
func (cc *CacheContext) NoStore() *CacheContext {
	cc.config.NoStore = true
	return cc
}

// MustRevalidate forces revalidation with origin server.
func (cc *CacheContext) MustRevalidate() *CacheContext {
	cc.config.MustRevalidate = true
	return cc
}

// JSON sends JSON with cache headers.
func (cc *CacheContext) JSON(code int, data interface{}) error {
	cc.applyCacheHeaders()
	return cc.ctx.JSON(code, data)
}

// HTML sends HTML with cache headers.
func (cc *CacheContext) HTML(code int, html string) error {
	cc.applyCacheHeaders()
	return cc.ctx.HTML(code, html)
}

// Text sends text with cache headers.
func (cc *CacheContext) Text(code int, text string) error {
	cc.applyCacheHeaders()
	return cc.ctx.Text(code, text)
}

// applyCacheHeaders applies cache-control headers.
func (cc *CacheContext) applyCacheHeaders() {
	var directives []string

	if cc.config.NoStore {
		directives = append(directives, "no-store")
	}
	if cc.config.NoCache {
		directives = append(directives, "no-cache")
	}
	if cc.config.Public {
		directives = append(directives, "public")
	}
	if cc.config.Private {
		directives = append(directives, "private")
	}
	if cc.config.MaxAge > 0 {
		directives = append(directives, fmt.Sprintf("max-age=%d", int(cc.config.MaxAge.Seconds())))
	}
	if cc.config.SMaxAge > 0 {
		directives = append(directives, fmt.Sprintf("s-maxage=%d", int(cc.config.SMaxAge.Seconds())))
	}
	if cc.config.MustRevalidate {
		directives = append(directives, "must-revalidate")
	}
	if cc.config.NoTransform {
		directives = append(directives, "no-transform")
	}

	if len(directives) > 0 {
		cc.ctx.SetHeader("Cache-Control", strings.Join(directives, ", "))
	}
}

// ETag sets the ETag header based on content.
func (c *Context) ETag(content string) *Context {
	hash := md5.Sum([]byte(content))
	etag := hex.EncodeToString(hash[:])
	c.SetHeader("ETag", `"`+etag+`"`)

	// Check If-None-Match header
	if c.Header("If-None-Match") == `"`+etag+`"` {
		c.writer.WriteHeader(304)
	}

	return c
}

// LastModified sets the Last-Modified header.
func (c *Context) LastModified(t time.Time) *Context {
	c.SetHeader("Last-Modified", t.UTC().Format(time.RFC1123))

	// Check If-Modified-Since header
	if ims := c.Header("If-Modified-Since"); ims != "" {
		if modTime, err := time.Parse(time.RFC1123, ims); err == nil {
			if !t.After(modTime) {
				c.Status(304)
				c.written = true
			}
		}
	}

	return c
}

// NoCache prevents caching.
func (c *Context) NoCache() *Context {
	c.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	c.SetHeader("Pragma", "no-cache")
	c.SetHeader("Expires", "0")
	return c
}
