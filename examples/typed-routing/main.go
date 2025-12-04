package main

import (
	"fmt"

	"github.com/ginjigo/ginji"
)

// Request/Response types for the API
type CreateUserRequest struct {
	Name  string `json:"name" ginji:"required"`
	Email string `json:"email" ginji:"required,email"`
	Age   int    `json:"age" ginji:"min=1,max=150"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type GetUserRequest struct {
	ID string `path:"id" ginji:"required"`
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty" ginji:"email"`
}

// In-memory storage
var users = make(map[int]*User)
var nextID = 1

func main() {
	app := ginji.New()

	// API v1 group
	v1 := app.Group("/api/v1")

	// Use typed handlers for all routes
	typed := v1.Typed()

	// Create user - POST /api/v1/users
	typed.Post("/users", func(c *ginji.Context, req CreateUserRequest) (User, error) {
		user := User{
			ID:    nextID,
			Name:  req.Name,
			Email: req.Email,
			Age:   req.Age,
		}
		users[nextID] = &user
		nextID++

		return user, nil
	})

	// Get user - GET /api/v1/users/:id
	typed.Get("/users/:id", func(c *ginji.Context, req GetUserRequest) (User, error) {
		// Parse ID from path
		id := 0
		fmt.Sscanf(req.ID, "%d", &id)

		user, ok := users[id]
		if !ok {
			return User{}, ginji.NewHTTPError(ginji.StatusNotFound, "User not found")
		}

		return *user, nil
	})

	// Update user - PUT /api/v1/users/:id
	v1.Put("/users/:id", func(c *ginji.Context) {
		// Example of mixing typed and regular handlers
		id := c.Param("id")

		var req UpdateUserRequest
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithError(ginji.StatusBadRequest, err)
			return
		}

		userID := 0
		fmt.Sscanf(id, "%d", &userID)

		user, ok := users[userID]
		if !ok {
			c.AbortWithError(ginji.StatusNotFound, ginji.NewHTTPError(ginji.StatusNotFound, "User not found"))
			return
		}

		if req.Name != "" {
			user.Name = req.Name
		}
		if req.Email != "" {
			user.Email = req.Email
		}

		_ = c.JSON(ginji.StatusOK, user)
	})

	// Delete user - DELETE /api/v1/users/:id
	typed.Delete("/users/:id", func(c *ginji.Context, req GetUserRequest) (ginji.EmptyRequest, error) {
		id := 0
		fmt.Sscanf(req.ID, "%d", &id)

		if _, ok := users[id]; !ok {
			return ginji.EmptyRequest{}, ginji.NewHTTPError(ginji.StatusNotFound, "User not found")
		}

		delete(users, id)
		return ginji.EmptyRequest{}, nil
	})

	// List all users - GET /api/v1/users
	typed.Get("/users", func(c *ginji.Context, req ginji.EmptyRequest) ([]User, error) {
		result := make([]User, 0, len(users))
		for _, user := range users {
			result = append(result, *user)
		}
		return result, nil
	})

	fmt.Println("🚀 Server running on http://localhost:3000")
	fmt.Println("Try these commands:")
	fmt.Println("  curl -X POST http://localhost:3000/api/v1/users -H 'Content-Type: application/json' -d '{\"name\":\"John Doe\",\"email\":\"john@example.com\",\"age\":30}'")
	fmt.Println("  curl http://localhost:3000/api/v1/users")
	fmt.Println("  curl http://localhost:3000/api/v1/users/1")
	fmt.Println("  curl -X DELETE http://localhost:3000/api/v1/users/1")

	app.Listen(":3000")
}
