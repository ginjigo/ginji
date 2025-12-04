package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ginjigo/ginji"
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
		_ = c.HTML(ginji.StatusOK, html)
	})

	app.Post("/upload", func(c *ginji.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			_ = c.Text(ginji.StatusBadRequest, "Error retrieving file")
			return
		}

		src, err := file.Open()
		if err != nil {
			_ = c.Text(ginji.StatusInternalServerError, err.Error())
			return
		}
		defer func() { _ = src.Close() }()

		// Save to disk
		dst, err := os.Create(file.Filename)
		if err != nil {
			_ = c.Text(ginji.StatusInternalServerError, err.Error())
			return
		}
		defer func() { _ = dst.Close() }()

		if _, err = io.Copy(dst, src); err != nil {
			_ = c.Text(ginji.StatusInternalServerError, err.Error())
			return
		}

		_ = c.Text(ginji.StatusOK, fmt.Sprintf("File %s uploaded successfully", file.Filename))
	})

	fmt.Println("Server running on :8084")
	if err := app.Listen(":8084"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
