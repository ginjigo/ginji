# What is Ginji?

Ginji is a high-performance, lightweight Go framework for building modern web services and APIs.

## Why Ginji?

### Zero Dependencies
100% standard library only. No external dependencies means:
- ✅ Smaller binary sizes
- ✅ Faster builds
- ✅ No dependency hell
- ✅ Maximum security

### Production-Ready
- **Error Handling** - HTTPError with validation errors
- **Security** - XSS, HSTS, CSP, CORS, rate limiting
- **Authentication** - Basic Auth, Bearer tokens, API keys, RBAC
- **Real-Time** - WebSocket, SSE, streaming
- **Testing** - Comprehensive testing utilities
- **Operations** - Health checks, timeouts, recovery

### High Performance
- Trie-based routing
- Minimal allocations
- Efficient middleware chaining
- Optional compression

### Developer Experience
- Clean, intuitive API
- Comprehensive documentation
- 107 tests, 91% coverage
- TypeScript-like chaining

## Feature Comparison

| Feature | Ginji | Gin | Fiber |
|---------|-------|-----|-------|
| Zero Dependencies | ✅ | ❌ | ❌ |
| Error Handling | ✅✅ | ✅ | ✅ |
| Validation | ✅✅ | ✅ | ✅ |
| WebSocket | ✅ | ❌ | ✅ |
| SSE | ✅ | ❌ | ❌ |
| Rate Limiting | ✅ | ❌* | ✅ |
| Auth Middleware | ✅ | ✅ | ✅ |
| Testing Utils | ✅✅ | ✅ | ✅ |

*Requires external package

## Quick Example

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Security
    app.Use(middleware.SecureStrict())
    app.Use(middleware.RateLimitPerMinute(100))

    // Operations
    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())
    app.Use(middleware.Health())

    // Routes
    app.Get("/", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{
            "message": "Hello, Ginji!",
        })
    })

    app.Listen(":3000")
}
```

## Use Cases

- **REST APIs** - Build scalable backend services
- **Microservices** - Lightweight, fast, zero deps
- **Real-Time Apps** - Chat, dashboards, notifications
- **File Services** - Upload, download, streaming
- **API Gateways** - Rate limiting, auth, routing

## Community

- **GitHub**: [github.com/ginjigo/ginji](https://github.com/ginjigo/ginji)
- **Docs**: https://ginjigo.github.io/ginji
- **License**: MIT

## Next Steps

- [Getting Started](/guide/getting-started)
- [Core Concepts](/guide/routing)
- [Middleware Guide](/middleware/)
- [Real-Time Features](/realtime/websocket)
