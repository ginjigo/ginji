package main

import (
	"fmt"
	"ginji/ginji"
	"io"
	"net/http"
	"os"
)

func main() {
	app := ginji.New()

	app.Get("/", func(c *ginji.Context) {
		html := `
		<!DOCTYPE html>
		<html>
		<body>
			<form action="/upload" method="post" enctype="multipart/form-data">
				Select file to upload:
				<input type="file" name="file" id="file">
				<input type="submit" value="Upload Image" name="submit">
			</form>
		</body>
		</html>
		`
		c.HTML(http.StatusOK, html)
	})

	app.Post("/upload", func(c *ginji.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.Text(http.StatusBadRequest, "Error retrieving file")
			return
		}

		src, err := file.Open()
		if err != nil {
			c.Text(http.StatusInternalServerError, err.Error())
			return
		}
		defer src.Close()

		// Save to disk
		dst, err := os.Create(file.Filename)
		if err != nil {
			c.Text(http.StatusInternalServerError, err.Error())
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			c.Text(http.StatusInternalServerError, err.Error())
			return
		}

		c.Text(http.StatusOK, fmt.Sprintf("File %s uploaded successfully", file.Filename))
	})

	fmt.Println("Server running on :8084")
	app.Listen(":8084")
}
