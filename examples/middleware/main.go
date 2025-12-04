package main

import (
	"fmt"
	"time"

	"github.com/ginjigo/ginji"
)

func main() {
	app := ginji.New()

	// Use built-in middleware
	app.Use(ginji.RequestID())
	app.Use(ginji.Compress())

	// Custom Logger Middleware
	app.Use(func(next ginji.Handler) ginji.Handler {
		return func(c *ginji.Context) {
			start := time.Now()
			next(c)
			fmt.Printf("[%s] %s %s %v\n",
				time.Now().Format(time.RFC3339),
				c.Req.Method,
				c.Req.URL.Path,
				time.Since(start),
			)
		}
	})

	app.Get("/", func(c *ginji.Context) {
		_ = c.Text(ginji.StatusOK, "Hello Middleware! Check headers for X-Request-ID and Content-Encoding.")
	})

	app.Get("/id", func(c *ginji.Context) {
		id, _ := c.Get("request_id")
		_ = c.JSON(ginji.StatusOK, ginji.H{
			"request_id": id,
		})
	})

	fmt.Println("Server running on :8081")
	if err := app.Listen(":8081"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
