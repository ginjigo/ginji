package ginji

import (
	"bytes"
	"html/template"
)

// SwaggerUIConfig represents Swagger UI configuration.
type SwaggerUIConfig struct {
	Title       string
	Description string
	Version     string
	SpecURL     string // URL to OpenAPI spec JSON (default: /openapi.json)
	BasePath    string // Base path for Swagger UI (default: /docs)
}

// SwaggerUI serves Swagger UI at the specified path.
func (engine *Engine) SwaggerUI(basePath string, config OpenAPIConfig) {
	if config.Title == "" {
		config.Title = "API Documentation"
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}

	specPath := basePath + "/openapi.json"

	// Serve OpenAPI spec JSON
	engine.Get(specPath, func(c *Context) {
		spec := engine.GenerateOpenAPI(config)
		_ = c.JSON(200, spec)
	})

	// Serve Swagger UI HTML
	engine.Get(basePath, func(c *Context) {
		html := generateSwaggerHTML(config.Title, specPath)
		_ = c.HTML(200, html)
	})
}

// generateSwaggerHTML generates the Swagger UI HTML page.
func generateSwaggerHTML(title, specURL string) string {
	tmpl := template.Must(template.New("swagger").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
        .swagger-ui .topbar {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '{{.SpecURL}}',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true,
                tryItOutEnabled: true
            });
            window.ui = ui;
        };
    </script>
</body>
</html>
`))

	type templateData struct {
		Title   string
		SpecURL string
	}

	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, templateData{
		Title:   title,
		SpecURL: specURL,
	})

	return buf.String()
}
