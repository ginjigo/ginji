package main

import (
	"fmt"

	"github.com/ginjigo/ginji"
)

type User struct {
	Name  string `json:"name" ginji:"required,min=3"`
	Email string `json:"email" ginji:"required,email"`
	Age   int    `json:"age" ginji:"min=18"`
}

type SearchQuery struct {
	Q    string `query:"q" ginji:"required"`
	Page int    `query:"page"`
}

func main() {
	app := ginji.New()

	app.Post("/users", func(c *ginji.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			_ = c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
			return
		}
		_ = c.JSON(ginji.StatusCreated, user)
	})

	app.Get("/search", func(c *ginji.Context) {
		var query SearchQuery
		if err := c.BindQuery(&query); err != nil {
			_ = c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
			return
		}
		_ = c.JSON(ginji.StatusOK, query)
	})

	fmt.Println("Server running on :8082")
	if err := app.Listen(":8082"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
