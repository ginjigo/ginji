package main

import (
	"fmt"
	"net/http"

	"github.com/ginjigo/ginji"
)

func main() {
	app := ginji.New()

	// Global Middleware
	app.Use(ginji.Logger())
	app.Use(ginji.Recovery())
	app.Use(ginji.CORS(ginji.DefaultCORS()))

	// Static Files
	app.Static("/assets", "./examples/basic/public")

	app.Get("/", func(c *ginji.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Ginji</h1>")
	})

	// Group Routing
	v1 := app.Group("/v1")
	{
		v1.Get("/hello", func(c *ginji.Context) {
			name := c.Query("name")
			if name == "" {
				name = "Guest"
			}
			c.JSON(http.StatusOK, ginji.H{
				"message": fmt.Sprintf("Hello %s", name),
			})
		})

		v1.Post("/login", func(c *ginji.Context) {
			var json struct {
				User     string `json:"user" ginji:"required"`
				Password string `json:"password" ginji:"required"`
			}
			if err := c.BindJSON(&json); err != nil {
				c.JSON(http.StatusBadRequest, ginji.H{"error": err.Error()})
				return
			}
			if json.User != "admin" || json.Password != "123456" {
				c.JSON(http.StatusUnauthorized, ginji.H{"status": "unauthorized"})
				return
			}
			c.JSON(http.StatusOK, ginji.H{"status": "authorized"})
		})

		v1.Get("/panic", func(c *ginji.Context) {
			panic("something went wrong")
		})
	}

	fmt.Println("Server is running on :3000")
	app.Run(":3000")
}
