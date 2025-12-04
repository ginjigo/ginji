<div align="center">
  <a href="https://ginji.io">
    <img src="https://raw.githubusercontent.com/ginjigo/ginji/main/assets/logo-text.png" width="500" height="auto" alt="Ginji"/>
  </a>
</div>

<hr />

[![CI](https://github.com/ginjigo/ginji/actions/workflows/ci.yml/badge.svg)](https://github.com/ginjigo/ginji/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ginjigo/ginji)](https://goreportcard.com/report/github.com/ginjigo/ginji)
[![License](https://img.shields.io/github/license/ginjigo/ginji)](https://github.com/ginjigo/ginji/blob/main/LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/ginjigo/ginji.svg)](https://pkg.go.dev/github.com/ginjigo/ginji)

Ginji is a brand-new, ultra-fast, zero-dependency API framework for Go, inspired by Hono, Fiber, and Gin. It aims to provide a minimal, fast, and clean foundation for building web applications.

```go
package main

import (
    "github.com/ginjigo/ginji"
)

func main() {
    app := ginji.New()

    app.Get("/", func(c *ginji.Context) {
        c.Text(ginji.StatusOK, "Hello Ginji")
    })

    app.Listen(":3000")
}
```

## Quick Start

```bash
# Install CLI
brew install ginjigo/tap/ginji
# OR
go install github.com/ginjigo/ginji/cmd/ginji@latest

# Create a new project
ginji new my-app
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
