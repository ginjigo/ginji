<div align="center">
  <h1>Ginji</h1>
  <p><strong>Ultra-fast, zero-dependency API framework for Go</strong></p>
</div>

<hr />

[![CI](https://github.com/kalana/ginji/actions/workflows/ci.yml/badge.svg)](https://github.com/kalana/ginji/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kalana/ginji)](https://goreportcard.com/report/github.com/kalana/ginji)
[![License](https://img.shields.io/github/license/kalana/ginji)](https://github.com/kalana/ginji/blob/main/LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/kalana/ginji.svg)](https://pkg.go.dev/github.com/kalana/ginji)

Ginji is a brand-new, ultra-fast, zero-dependency API framework for Go, inspired by Hono, Fiber, and Gin. It aims to provide a minimal, fast, and clean foundation for building web applications.

```go
package main

import (
    "ginji/ginji"
    "net/http"
)

func main() {
    app := ginji.New()

    app.Get("/", func(c *ginji.Context) {
        c.Text(http.StatusOK, "Hello Ginji")
    })

    app.Listen(":3000")
}
```

## Quick Start

```bash
go get github.com/kalana/ginji
```

## Features

- **Ultrafast** 🚀 - Built for performance with minimal overhead.
- **Hono-like Routing** 🛣️ - Simple and expressive routing with dynamic parameter support (`/users/:id`).
- **Middleware Support** 🧩 - Easy-to-use middleware system for global and per-route logic.
- **Zero Dependencies** 📦 - Uses only the Go standard library.
- **Production Ready** 🛠️ - Clean architecture designed for scalability.

## Documentation

See the [examples](examples) directory for more usage examples.

## Contributing

Contributions Welcome! You can contribute in the following ways.

- Create an Issue - Propose a new feature. Report a bug.
- Pull Request - Fix a bug and typo. Refactor the code.

For more details, see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.
