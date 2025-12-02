package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ProjectOptions struct {
	Name        string
	Database    string // "None", "SQLite", "PostgreSQL", "MySQL"
	ORM         string // "None", "GORM", "sqlc", "ent"
	Middlewares []string
	Deployment  string // "None", "Docker"
	Tests       bool
}

const (
	repoURL      = "https://github.com/ginjigo/ginji/archive/refs/heads/main.zip"
	templatePath = "ginji-main/templates/basic"
)

func generateProject(opts ProjectOptions) error {
	fmt.Printf("Generating project '%s'...\n", opts.Name)

	// 1. Download and extract template
	if err := downloadAndExtractTemplate(opts.Name); err != nil {
		return fmt.Errorf("failed to download template: %w", err)
	}

	// 2. Process templates (replace placeholders)
	if err := processTemplates(opts.Name, opts); err != nil {
		return fmt.Errorf("failed to process templates: %w", err)
	}

	// 3. Clean up
	if opts.Deployment != "Docker" {
		if err := os.Remove(filepath.Join(opts.Name, "Dockerfile")); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove Dockerfile: %w", err)
		}
	}

	return nil
}

func downloadAndExtractTemplate(destDir string) error {
	fmt.Println("Downloading template from GitHub...")
	resp, err := http.Get(repoURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	fmt.Println("Extracting template...")
	for _, file := range zipReader.File {
		// Check if file is inside the template directory
		if strings.HasPrefix(file.Name, templatePath) {
			// Determine destination path
			relPath, err := filepath.Rel(templatePath, file.Name)
			if err != nil {
				return err
			}

			// Skip the root folder itself
			if relPath == "." {
				continue
			}

			targetPath := filepath.Join(destDir, relPath)

			if file.FileInfo().IsDir() {
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					return err
				}
				continue
			}

			// Ensure parent dir exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			// Create file
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}

			rc, err := file.Open()
			if err != nil {
				outFile.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processTemplates(rootDir string, opts ProjectOptions) error {
	fmt.Println("Configuring project...")
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse as template
		tmpl, err := template.New(path).Parse(string(content))
		if err != nil {
			// If it fails to parse (e.g. binary files), just skip
			return nil
		}

		// Execute template
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, opts); err != nil {
			return err
		}

		// Write back
		return os.WriteFile(path, buf.Bytes(), info.Mode())
	})
}
