package main

import (
	"fmt"

	"github.com/ginjigo/ginji"

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
	if err := db.AutoMigrate(&User{}); err != nil {
		fmt.Printf("Failed to migrate database: %v\n", err)
	}

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
			_ = c.JSON(ginji.StatusBadRequest, ginji.H{"error": err.Error()})
			return
		}

		conn.Create(&user)
		_ = c.JSON(ginji.StatusCreated, user)
	})

	app.Get("/users", func(c *ginji.Context) {
		db, _ := c.Get("db")
		conn := db.(*gorm.DB)

		var users []User
		conn.Find(&users)
		_ = c.JSON(ginji.StatusOK, users)
	})

	app.Get("/users/:id", func(c *ginji.Context) {
		db, _ := c.Get("db")
		conn := db.(*gorm.DB)
		id := c.Param("id")

		var user User
		if result := conn.First(&user, id); result.Error != nil {
			_ = c.JSON(ginji.StatusNotFound, ginji.H{"error": "User not found"})
			return
		}
		_ = c.JSON(ginji.StatusOK, user)
	})

	fmt.Println("Server is running on :3000")
	if err := app.Run(":3000"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
