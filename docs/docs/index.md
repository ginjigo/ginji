---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "Ginji"
  text: "Go Framework for Modern Web Services"
  tagline: High-performance, zero-dependency Go framework with real-time capabilities
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/ginjigo/ginji

features:
  - icon: ⚡
    title: Zero Dependencies
    details: 100% standard library only. No external dependencies means smaller binaries, faster builds, and maximum security.
  
  - icon: 🔒
    title: Production Security
    details: Built-in security headers, rate limiting, authentication (Basic/Bearer/API Key), RBAC, and more.
  
  - icon: 🚀
    title: Real-Time Ready
    details: Native WebSocket and Server-Sent Events support for chat, dashboards, and live notifications.
  
  - icon: ✅
    title: Advanced Validation
    details: 15+ built-in validators, nested struct support, custom validators, and structured error responses.
  
  - icon: 🧪
    title: Testing First
    details: Comprehensive testing utilities with request simulation, assertions, and 107 passing tests.
  
  - icon: 📦
    title: Complete Middleware
    details: Body limit, security headers, health checks, rate limiting, auth, timeout - all included.
---

## Quick Start

```bash
go get github.com/ginjigo/ginji
```

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())

    app.Get("/", func(c *ginji.Context) {
        c.JSON(ginji.StatusOK, ginji.H{
            "message": "Hello, Ginji!",
        })
    })

    app.Listen(":3000")
}
```

## Why Ginji?

- **Zero Dependencies** - Pure Go, no external packages
- **Production Ready** - Security, auth, rate limiting, health checks
- **Real-Time** - WebSocket & SSE built-in
- **Developer Friendly** - Clean API, great docs, 91% test coverage
- **High Performance** - Efficient routing, minimal allocations

[Get Started →](/guide/getting-started)
