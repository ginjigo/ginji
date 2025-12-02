package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectOptions struct {
	Name        string
	Database    string // "None", "SQLite", "PostgreSQL", "MySQL"
	ORM         string // "None", "GORM", "sqlc", "ent"
	Middlewares []string
	Deployment  string // "None", "Docker"
	Tests       bool
}

func generateProject(opts ProjectOptions) error {
	fmt.Printf("Generating project with opts: %+v\n", opts)
	// Create directories
	dirs := []string{
		opts.Name,
		filepath.Join(opts.Name, "cmd", "server"),
		filepath.Join(opts.Name, "internal", "handlers"),
		filepath.Join(opts.Name, "internal", "models"),
	}

	if opts.Database != "None" {
		dirs = append(dirs, filepath.Join(opts.Name, "internal", "database"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create go.mod
	goModContent := fmt.Sprintf("module %s\n\ngo 1.23.0\n", opts.Name)
	if err := createFile(filepath.Join(opts.Name, "go.mod"), goModContent); err != nil {
		return err
	}

	// Create handlers
	if err := createHandlers(opts); err != nil {
		return err
	}

	// Create models
	if err := createModels(opts); err != nil {
		return err
	}

	// Create main.go
	if err := createMainGo(opts); err != nil {
		return err
	}

	// Create database.go if needed
	if opts.Database != "None" {
		if err := createDatabaseGo(opts); err != nil {
			return err
		}
	}

	// Create Dockerfile
	if opts.Deployment == "Docker" {
		if err := createDockerfile(opts); err != nil {
			return err
		}
	}

	// Create Tests
	if opts.Tests {
		if err := createTests(opts); err != nil {
			return err
		}
	}

	return nil
}

func createHandlers(opts ProjectOptions) error {
	content := fmt.Sprintf(`package handlers

import (
	"net/http"
	"github.com/ginjigo/ginji"
)

func HelloHandler(c *ginji.Context) {
	c.JSON(http.StatusOK, ginji.H{"message": "Welcome to %s!"})
}

func HealthCheck(c *ginji.Context) {
	c.JSON(http.StatusOK, ginji.H{"status": "ok"})
}
`, opts.Name)
	return createFile(filepath.Join(opts.Name, "internal", "handlers", "hello.go"), content)
}

func createModels(opts ProjectOptions) error {
	content := `package models

type User struct {
	ID    int    ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}
`
	return createFile(filepath.Join(opts.Name, "internal", "models", "user.go"), content)
}

func createMainGo(opts ProjectOptions) error {
	imports := []string{
		"fmt",
		"github.com/ginjigo/ginji",
		fmt.Sprintf("%s/internal/handlers", opts.Name),
	}

	if opts.Database != "None" {
		imports = append(imports, fmt.Sprintf("%s/internal/database", opts.Name))
	}

	mainBody := ""

	// Database Init
	if opts.Database != "None" {
		mainBody += "\t// Initialize Database\n"
		mainBody += "\tdatabase.Connect()\n\n"
	}

	// App Init
	mainBody += "\tapp := ginji.New()\n\n"

	// Middlewares
	if len(opts.Middlewares) > 0 {
		mainBody += "\t// Middlewares\n"
		for _, m := range opts.Middlewares {
			if m == "Logger" {
				mainBody += "\tapp.Use(ginji.Logger())\n"
			} else if m == "Recovery" {
				mainBody += "\tapp.Use(ginji.Recovery())\n"
			} else if m == "CORS" {
				mainBody += "\tapp.Use(ginji.CORS(ginji.DefaultCORS()))\n"
			}
		}
		mainBody += "\n"
	}

	// Routes
	mainBody += "\t// Routes\n"
	mainBody += "\tapp.Get(\"/\", handlers.HelloHandler)\n"
	mainBody += "\tapp.Get(\"/health\", handlers.HealthCheck)\n\n"

	// Run
	mainBody += "\tfmt.Println(\"Server running on :8080\")\n"
	mainBody += "\tapp.Run(\":8080\")\n"

	content := fmt.Sprintf("package main\n\nimport (\n\t\"%s\"\n)\n\nfunc main() {\n%s}\n", strings.Join(imports, "\"\n\t\""), mainBody)
	return createFile(filepath.Join(opts.Name, "cmd", "server", "main.go"), content)
}

func createDatabaseGo(opts ProjectOptions) error {
	content := "package database\n\n"

	if opts.ORM == "GORM" {
		content += "import (\n"
		content += "\t\"gorm.io/gorm\"\n"
		if opts.Database == "SQLite" {
			content += "\t\"gorm.io/driver/sqlite\"\n"
		} else if opts.Database == "PostgreSQL" {
			content += "\t\"gorm.io/driver/postgres\"\n"
		} else if opts.Database == "MySQL" {
			content += "\t\"gorm.io/driver/mysql\"\n"
		}
		content += ")\n\n"
		content += "var DB *gorm.DB\n\n"
		content += "func Connect() {\n"
		content += "\tvar err error\n"
		if opts.Database == "SQLite" {
			content += "\tDB, err = gorm.Open(sqlite.Open(\"app.db\"), &gorm.Config{})\n"
		} else {
			content += "\t// TODO: Configure connection string\n"
			content += "\t// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})\n"
		}
		content += "\tif err != nil {\n"
		content += "\t\tpanic(\"failed to connect database\")\n"
		content += "\t}\n"
		content += "}\n"
	} else if opts.ORM == "sqlc" {
		content += "import (\n"
		content += "\t\"database/sql\"\n"
		content += "\t_ \"github.com/lib/pq\" // Example for Postgres\n"
		content += ")\n\n"
		content += "var DB *sql.DB\n\n"
		content += "func Connect() {\n"
		content += "\t// TODO: Initialize sqlc connection\n"
		content += "\t// DB, err := sql.Open(\"postgres\", \"user=foo dbname=bar sslmode=disable\")\n"
		content += "}\n"
	} else if opts.ORM == "ent" {
		content += "import (\n"
		content += "\t\"context\"\n"
		content += "\t\"log\"\n"
		content += "\t// \"your_project/ent\" // Import your generated ent package\n"
		content += ")\n\n"
		content += "func Connect() {\n"
		content += "\t// client, err := ent.Open(\"sqlite3\", \"file:ent?mode=memory&cache=shared&_fk=1\")\n"
		content += "\t// if err != nil {\n"
		content += "\t// \tlog.Fatalf(\"failed opening connection to sqlite: %v\", err)\n"
		content += "\t// }\n"
		content += "\t// defer client.Close()\n"
		content += "\t// if err := client.Schema.Create(context.Background()); err != nil {\n"
		content += "\t// \tlog.Fatalf(\"failed creating schema resources: %v\", err)\n"
		content += "\t// }\n"
		content += "}\n"
	} else {
		content += "func Connect() {\n"
		content += "\t// TODO: Implement database connection\n"
		content += "}\n"
	}

	return createFile(filepath.Join(opts.Name, "internal", "database", "database.go"), content)
}

func createDockerfile(opts ProjectOptions) error {
	content := `FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /server cmd/server/main.go

EXPOSE 8080

CMD [ "/server" ]
`
	return createFile(filepath.Join(opts.Name, "Dockerfile"), content)
}

func createTests(opts ProjectOptions) error {
	content := fmt.Sprintf(`package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/ginjigo/ginji"
	"%s/internal/handlers"
)

func TestRootRoute(t *testing.T) {
	app := ginji.New()
	app.Get("/", handlers.HelloHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %%d, got %%d", http.StatusOK, w.Code)
	}

	expected := "{\"message\":\"Welcome to %s!\"}\n"
	if w.Body.String() != expected {
		t.Errorf("Expected body %%s, got %%s", expected, w.Body.String())
	}
}
`, opts.Name, opts.Name)
	return createFile(filepath.Join(opts.Name, "cmd", "server", "main_test.go"), content)
}

func createFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
