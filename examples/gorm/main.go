package main

import (
	"fmt"
	"ginji/ginji"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User model
type User struct {
	gorm.Model
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Initialize GORM with SQLite
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&User{})

	app := ginji.New()
	app.Use(ginji.Logger())
	app.Use(ginji.Recovery())

	// Middleware to inject DB into context
	app.Use(func(next ginji.Handler) ginji.Handler {
		return func(c *ginji.Context) {
			c.Set("db", db)
			next(c)
		}
	})

	app.Post("/users", func(c *ginji.Context) {
		db, _ := c.Get("db")
		conn := db.(*gorm.DB)

		var user User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, ginji.H{"error": err.Error()})
			return
		}

		conn.Create(&user)
		c.JSON(http.StatusCreated, user)
	})

	app.Get("/users", func(c *ginji.Context) {
		db, _ := c.Get("db")
		conn := db.(*gorm.DB)

		var users []User
		conn.Find(&users)
		c.JSON(http.StatusOK, users)
	})

	app.Get("/users/:id", func(c *ginji.Context) {
		db, _ := c.Get("db")
		conn := db.(*gorm.DB)
		id := c.Param("id")

		var user User
		if result := conn.First(&user, id); result.Error != nil {
			c.JSON(http.StatusNotFound, ginji.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	fmt.Println("Server is running on :3000")
	app.Run(":3000")
}
