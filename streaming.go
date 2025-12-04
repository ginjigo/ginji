package ginji

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Stream sends a streaming response from an io.Reader.
func (c *Context) Stream(contentType string, reader io.Reader) error {
	c.SetHeader("Content-Type", contentType)
	c.SetHeader("Cache-Control", "no-cache")
	c.SetHeader("X-Accel-Buffering", "no") // Disable nginx buffering

	// Set chunked transfer encoding
	c.SetHeader("Transfer-Encoding", "chunked")

	// Copy from reader to response
	_, err := io.Copy(c.Res, reader)
	return err
}

// File sends a file with proper headers.
func (c *Context) File(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Set headers
	c.SetHeader("Content-Type", detectContentType(filepath))
	c.SetHeader("Content-Length", fmt.Sprintf("%d", stat.Size()))
	c.SetHeader("Last-Modified", stat.ModTime().UTC().Format(http.TimeFormat))

	// Send file
	_, err = io.Copy(c.Res, file)
	return err
}

// Attachment sends a file as a downloadable attachment.
func (c *Context) Attachment(filepath, filename string) error {
	if filename == "" {
		filename = filepath
	}

	c.SetHeader("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.File(filepath)
}

// FileStream streams a file without buffering the entire content.
func (c *Context) FileStream(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	contentType := detectContentType(filepath)
	return c.Stream(contentType, file)
}

// detectContentType returns the content type based on file extension.
func detectContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// SaveUploadedFile saves an uploaded file to dst.
func (c *Context) SaveUploadedFile(fileHeader *multipart.FileHeader, dst string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	// Copy content
	_, err = io.Copy(out, src)
	return err
}

// ChunkedJSON sends JSON in chunks (for large responses).
func (c *Context) ChunkedJSON(v any) error {
	c.SetHeader("Content-Type", "application/json")
	c.SetHeader("Transfer-Encoding", "chunked")

	data, err := jsonMarshal(v)
	if err != nil {
		return err
	}

	_, err = c.Res.Write(data)
	if flusher, ok := c.Res.(http.Flusher); ok {
		flusher.Flush()
	}

	return err
}

// StreamJSON streams JSON objects one by one.
func (c *Context) StreamJSON(items <-chan any) error {
	c.SetHeader("Content-Type", "application/json")
	c.SetHeader("Transfer-Encoding", "chunked")

	// Start array
	_, _ = c.Res.Write([]byte("["))
	if flusher, ok := c.Res.(http.Flusher); ok {
		flusher.Flush()
	}

	first := true
	for item := range items {
		if !first {
			_, _ = c.Res.Write([]byte(","))
		}
		first = false

		data, err := jsonMarshal(item)
		if err != nil {
			return err
		}

		_, _ = c.Res.Write(data)
		if flusher, ok := c.Res.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	// End array
	c.Res.Write([]byte("]"))
	if flusher, ok := c.Res.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}
