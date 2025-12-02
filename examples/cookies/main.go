package main

import (
	"fmt"
	"ginji/ginji"
	"net/http"
	"time"
)

func main() {
	app := ginji.New()

	app.Get("/set", func(c *ginji.Context) {
		cookie := &http.Cookie{
			Name:     "user",
			Value:    "ginji-user",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		c.SetCookie(cookie)
		c.Text(http.StatusOK, "Cookie set!")
	})

	app.Get("/get", func(c *ginji.Context) {
		cookie, err := c.Cookie("user")
		if err != nil {
			c.Text(http.StatusBadRequest, "Cookie not found")
			return
		}
		c.Text(http.StatusOK, "User: "+cookie.Value)
	})

	fmt.Println("Server running on :8083")
	app.Listen(":8083")
}
