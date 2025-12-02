package main

import (
	"fmt"

	"github.com/ginjigo/ginji"
	"{{.Name}}/internal/database"
	"{{.Name}}/internal/handlers"
)

func main() {
	// Initialize Database
	database.Connect()

	app := ginji.New()

	// Middlewares
	app.Use(ginji.Logger())
	app.Use(ginji.Recovery())
	app.Use(ginji.CORS(ginji.DefaultCORS()))

	// Routes
	app.Get("/", handlers.HelloHandler)
	app.Get("/health", handlers.HealthCheck)

	fmt.Println("Server running on :8080")
	app.Run(":8080")
}
