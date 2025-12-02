package main

import (
	"fmt"

	"github.com/ginjigo/ginji/ginji"
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
			c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
			return
		}
		c.JSON(ginji.StatusCreated, user)
	})

	app.Get("/search", func(c *ginji.Context) {
		var query SearchQuery
		if err := c.BindQuery(&query); err != nil {
			c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
			return
		}
		c.JSON(ginji.StatusOK, query)
	})

	fmt.Println("Server running on :8082")
	app.Listen(":8082")
}
