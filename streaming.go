package ginji

import (
	"errors"
	"fmt"
	"io"
	"log" // Added log import
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	// Validate file path to prevent directory traversal
	if err := validateFilePath(filepath); err != nil {
		return err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}()

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

	// Sanitize filename to prevent header injection
	filename = sanitizeFilename(filename)
	c.SetHeader("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.File(filepath)
}

// FileStream streams a file without buffering the entire content.
func (c *Context) FileStream(filepath string) error {
	// Validate file path to prevent directory traversal
	if err := validateFilePath(filepath); err != nil {
		return err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}()

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
	// Validate destination path to prevent directory traversal
	if err := validateFilePath(dst); err != nil {
		return err
	}

	// Check file size (default 32MB limit)
	const maxUploadSize = 32 << 20 // 32 MB
	if fileHeader.Size > maxUploadSize {
		return fmt.Errorf("file too large: %d bytes (max %d bytes)", fileHeader.Size, maxUploadSize)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Printf("Failed to close source file: %v", err)
		}
	}()

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Printf("Failed to close output file: %v", err)
		}
	}()

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
	_, _ = c.Res.Write([]byte("]"))
	if flusher, ok := c.Res.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// validateFilePath checks if a file path is safe and doesn't contain directory traversal attempts.
func validateFilePath(path string) error {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check for directory traversal patterns
	if strings.Contains(cleanPath, "..") {
		return errors.New("invalid file path: directory traversal detected")
	}

	// Ensure it's not an absolute path (unless explicitly allowed)
	// This prevents accessing arbitrary system files
	if filepath.IsAbs(cleanPath) {
		return errors.New("invalid file path: absolute paths not allowed")
	}

	return nil
}

// sanitizeFilename removes potentially dangerous characters from filenames
// to prevent header injection and other attacks.
func sanitizeFilename(filename string) string {
	// Get just the base filename (no path components)
	filename = filepath.Base(filename)

	// Remove or replace dangerous characters
	// Replace quotes, newlines, carriage returns that could break headers
	replacer := strings.NewReplacer(
		"\"", "",
		"\r", "",
		"\n", "",
		"\\", "",
	)

	return replacer.Replace(filename)
}
