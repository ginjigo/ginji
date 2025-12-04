package main

import (
	"github.com/ginjigo/ginji"
)

// User represents a user model.
type User struct {
	ID    int    `json:"id" description:"Unique user identifier" example:"1"`
	Name  string `json:"name" validate:"required" description:"User's full name" example:"John Doe"`
	Email string `json:"email" validate:"required,email" description:"User's email address" example:"john@example.com"`
	Age   int    `json:"age" validate:"gte=0,lte=150" description:"User's age" example:"30"`
}

// CreateUserRequest represents the request for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required" description:"User's full name" example:"John Doe"`
	Email string `json:"email" validate:"required,email" description:"User's email address" example:"john@example.com"`
	Age   int    `json:"age" validate:"gte=0,lte=150" description:"User's age" example:"30"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error" description:"Error message"`
	Message string `json:"message" description:"Detailed error message"`
}

func main() {
	app := ginji.New()

	// Define routes with OpenAPI metadata
	app.Get("/", func(c *ginji.Context) {
		_ = c.JSON(200, map[string]string{
			"message": "Welcome to Ginji API with OpenAPI/Swagger documentation!",
			"docs":    "Visit /docs for interactive API documentation",
		})
	}).
		Summary("Welcome endpoint").
		Description("Returns a welcome message and link to API documentation").
		Tags("general").
		Response(200, map[string]string{})

	// List users
	app.Get("/users", listUsers).
		Summary("List all users").
		Description("Returns a list of all registered users").
		Tags("users").
		Response(200, []User{}).
		Response(500, ErrorResponse{})

	// Get user by ID
	app.Get("/users/:id", getUser).
		Summary("Get user by ID").
		Description("Returns a single user by their unique identifier").
		Tags("users").
		Response(200, User{}).
		Response(404, ErrorResponse{})

	// Create user
	app.Post("/users", createUser).
		Summary("Create a new user").
		Description("Creates a new user with the provided information").
		Tags("users").
		Request(CreateUserRequest{}).
		Response(201, User{}).
		Response(400, ErrorResponse{})

	// Update user
	app.Put("/users/:id", updateUser).
		Summary("Update user").
		Description("Updates an existing user's information").
		Tags("users").
		Request(CreateUserRequest{}).
		Response(200, User{}).
		Response(404, ErrorResponse{})

	// Delete user
	app.Delete("/users/:id", deleteUser).
		Summary("Delete user").
		Description("Deletes a user by their unique identifier").
		Tags("users").
		Response(204, nil).
		Response(404, ErrorResponse{})

	// Configure and serve OpenAPI documentation
	app.SwaggerUI("/docs", ginji.OpenAPIConfig{
		Title:       "Ginji Demo API",
		Description: "A demonstration API built with Ginji framework showcasing OpenAPI/Swagger integration",
		Version:     "1.0.0",
		Contact: &ginji.OpenAPIContact{
			Name:  "API Support",
			Email: "support@example.com",
		},
		License: &ginji.OpenAPILicense{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
		Servers: []ginji.OpenAPIServer{
			{
				URL:         "http://localhost:3000",
				Description: "Development server",
			},
		},
		Tags: []ginji.OpenAPITag{
			{
				Name:        "general",
				Description: "General endpoints",
			},
			{
				Name:        "users",
				Description: "User management endpoints",
			},
		},
	})

	println("Server starting on :3000")
	println("📚 API Documentation: http://localhost:3000/docs")
	println("📄 OpenAPI Spec: http://localhost:3000/docs/openapi.json")
	_ = app.Listen(":3000")
}

// Sample handlers
var users = []User{
	{ID: 1, Name: "Alice Smith", Email: "alice@example.com", Age: 28},
	{ID: 2, Name: "Bob Johnson", Email: "bob@example.com", Age: 35},
}

func listUsers(c *ginji.Context) {
	_ = c.JSON(200, users)
}

func getUser(c *ginji.Context) {
	id := c.Param("id")
	for _, user := range users {
		if user.ID == parseID(id) {
			_ = c.JSON(200, user)
			return
		}
	}
	_ = c.JSON(404, ErrorResponse{
		Error:   "not_found",
		Message: "User not found",
	})
}

func createUser(c *ginji.Context) {
	var req CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		_ = c.JSON(400, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	user := User{
		ID:    len(users) + 1,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	users = append(users, user)

	_ = c.JSON(201, user)
}

func updateUser(c *ginji.Context) {
	id := c.Param("id")
	var req CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		_ = c.JSON(400, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	for i, user := range users {
		if user.ID == parseID(id) {
			users[i].Name = req.Name
			users[i].Email = req.Email
			users[i].Age = req.Age
			_ = c.JSON(200, users[i])
			return
		}
	}

	_ = c.JSON(404, ErrorResponse{
		Error:   "not_found",
		Message: "User not found",
	})
}

func deleteUser(c *ginji.Context) {
	id := c.Param("id")
	for i, user := range users {
		if user.ID == parseID(id) {
			users = append(users[:i], users[i+1:]...)
			c.Status(204)
			return
		}
	}

	_ = c.JSON(404, ErrorResponse{
		Error:   "not_found",
		Message: "User not found",
	})
}

func parseID(id string) int {
	var result int
	for _, ch := range id {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		}
	}
	return result
}
