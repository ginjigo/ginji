package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// Templates for code generation
const handlerTemplate = `package handlers

import "github.com/ginjigo/ginji"

// {{.Name}} handles {{.Description}}
func {{.Name}}(c *ginji.Context) {
	// TODO: Implement handler logic
	{{if .UsesJSON}}
	var request struct {
		// TODO: Define request structure
	}
	
	if err := c.BindValidate(&request); err != nil {
		_ = c.JSON(400, ginji.H{"error": err.Error()})
		return
	}
	{{end}}
	
	_ = c.JSON(200, ginji.H{
		"message": "{{.Name}} endpoint",
	})
}
`

const middlewareTemplate = `package middleware

import "github.com/ginjigo/ginji"

// {{.Name}} is a middleware that {{.Description}}
func {{.Name}}() ginji.Middleware {
	return func(c *ginji.Context) {
		// TODO: Implement middleware logic before handler
		
		// Call the next handler
		c.Next()
		
		// TODO: Implement middleware logic after handler
	}
}
`

const crudTemplate = `package handlers

import (
	"github.com/ginjigo/ginji"
)

// {{.Name}}Handler provides CRUD operations for {{.Resource}}
type {{.Name}}Handler struct {
	// TODO: Add dependencies (database, services, etc.)
}

// New{{.Name}}Handler creates a new {{.Name}}Handler
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{}
}

// List{{.Plural}} handles GET /{{.RoutePath}}
func (h *{{.Name}}Handler) List{{.Plural}}(c *ginji.Context) {
	// TODO: Implement list logic
	_ = c.JSON(200, ginji.H{
		"{{.Resource}}s": []interface{}{},
	})
}

// Get{{.Name}} handles GET /{{.RoutePath}}/:id
func (h *{{.Name}}Handler) Get{{.Name}}(c *ginji.Context) {
	id := c.Param("id")
	
	// TODO: Implement get logic
	_ = c.JSON(200, ginji.H{
		"id": id,
	})
}

// Create{{.Name}} handles POST /{{.RoutePath}}
func (h *{{.Name}}Handler) Create{{.Name}}(c *ginji.Context) {
	var request struct {
		// TODO: Define create request structure
	}
	
	if err := c.BindValidate(&request); err != nil {
		_ = c.JSON(400, ginji.H{"error": err.Error()})
		return
	}
	
	// TODO: Implement create logic
	_ = c.JSON(201, ginji.H{
		"message": "{{.Name}} created",
	})
}

// Update{{.Name}} handles PUT /{{.RoutePath}}/:id
func (h *{{.Name}}Handler) Update{{.Name}}(c *ginji.Context) {
	id := c.Param("id")
	
	var request struct {
		// TODO: Define update request structure
	}
	
	if err := c.BindValidate(&request); err != nil {
		_ = c.JSON(400, ginji.H{"error": err.Error()})
		return
	}
	
	// TODO: Implement update logic
	_ = c.JSON(200, ginji.H{
		"id":      id,
		"message": "{{.Name}} updated",
	})
}

// Delete{{.Name}} handles DELETE /{{.RoutePath}}/:id
func (h *{{.Name}}Handler) Delete{{.Name}}(c *ginji.Context) {
	id := c.Param("id")
	
	// TODO: Implement delete logic
	_ = c.JSON(200, ginji.H{
		"id":      id,
		"message": "{{.Name}} deleted",
	})
}

// Register{{.Name}}Routes registers all {{.Resource}} routes
func Register{{.Name}}Routes(app *ginji.Engine) {
	handler := New{{.Name}}Handler()
	
	app.Get("/{{.RoutePath}}", handler.List{{.Plural}}).
		Summary("List all {{.Resource}}s").
		Tags("{{.Resource}}")
	
	app.Get("/{{.RoutePath}}/:id", handler.Get{{.Name}}).
		Summary("Get a {{.Resource}} by ID").
		Tags("{{.Resource}}")
	
	app.Post("/{{.RoutePath}}", handler.Create{{.Name}}).
		Summary("Create a new {{.Resource}}").
		Tags("{{.Resource}}")
	
	app.Put("/{{.RoutePath}}/:id", handler.Update{{.Name}}).
		Summary("Update a {{.Resource}}").
		Tags("{{.Resource}}")
	
	app.Delete("/{{.RoutePath}}/:id", handler.Delete{{.Name}}).
		Summary("Delete a {{.Resource}}").
		Tags("{{.Resource}}")
}
`

type GenerateData struct {
	Name        string
	Description string
	UsesJSON    bool
	Resource    string
	Plural      string
	RoutePath   string
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// toPlural makes a simple pluralization
func toPlural(s string) string {
	if strings.HasSuffix(s, "s") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	return s + "s"
}

// generateHandler generates a new handler file
func generateHandler(name string, useJSON bool) error {
	pascalName := toPascalCase(name)
	kebabName := toKebabCase(pascalName)

	data := GenerateData{
		Name:        pascalName,
		Description: fmt.Sprintf("%s operations", kebabName),
		UsesJSON:    useJSON,
	}

	tmpl, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		return err
	}

	os.MkdirAll("handlers", 0755)
	filename := filepath.Join("handlers", fmt.Sprintf("%s.go", kebabName))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return err
	}

	fmt.Printf("✓ Generated handler: %s\n", filename)
	return nil
}

// generateMiddleware generates a new middleware file
func generateMiddleware(name string) error {
	pascalName := toPascalCase(name)
	kebabName := toKebabCase(pascalName)

	data := GenerateData{
		Name:        pascalName,
		Description: fmt.Sprintf("processes %s", kebabName),
	}

	tmpl, err := template.New("middleware").Parse(middlewareTemplate)
	if err != nil {
		return err
	}

	os.MkdirAll("middleware", 0755)
	filename := filepath.Join("middleware", fmt.Sprintf("%s.go", kebabName))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return err
	}

	fmt.Printf("✓ Generated middleware: %s\n", filename)
	return nil
}

// generateCRUD generates a full CRUD handler
func generateCRUD(resource string) error {
	pascalName := toPascalCase(resource)
	kebabName := toKebabCase(pascalName)

	data := GenerateData{
		Name:      pascalName,
		Resource:  strings.ToLower(resource),
		Plural:    toPlural(pascalName),
		RoutePath: toKebabCase(toPlural(pascalName)),
	}

	tmpl, err := template.New("crud").Parse(crudTemplate)
	if err != nil {
		return err
	}

	os.MkdirAll("handlers", 0755)
	filename := filepath.Join("handlers", fmt.Sprintf("%s.go", kebabName))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return err
	}

	fmt.Printf("✓ Generated CRUD handler: %s\n", filename)
	fmt.Printf("  Routes:\n")
	fmt.Printf("    GET    /%s\n", data.RoutePath)
	fmt.Printf("    GET    /%s/:id\n", data.RoutePath)
	fmt.Printf("    POST   /%s\n", data.RoutePath)
	fmt.Printf("    PUT    /%s/:id\n", data.RoutePath)
	fmt.Printf("    DELETE /%s/:id\n", data.RoutePath)

	return nil
}
