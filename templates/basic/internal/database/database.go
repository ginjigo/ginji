{{if eq .ORM "GORM"}}
import (
	"gorm.io/gorm"
{{if eq .Database "SQLite"}}	"gorm.io/driver/sqlite"{{end}}
{{if eq .Database "PostgreSQL"}}	"gorm.io/driver/postgres"{{end}}
{{if eq .Database "MySQL"}}	"gorm.io/driver/mysql"{{end}}
)

var DB *gorm.DB

func Connect() {
	var err error
{{if eq .Database "SQLite"}}	DB, err = gorm.Open(sqlite.Open("app.db"), &gorm.Config{}){{end}}
{{if eq .Database "PostgreSQL"}}	// dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
	// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}){{end}}
{{if eq .Database "MySQL"}}	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}){{end}}

	if err != nil {
		panic("failed to connect database")
	}
}
{{else if eq .ORM "sqlc"}}
import (
	"database/sql"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	// TODO: Initialize sqlc connection
	// DB, err := sql.Open("postgres", "user=foo dbname=bar sslmode=disable")
}
{{else if eq .ORM "ent"}}
import (
	"context"
	"log"
	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	// "your_project/ent" // Import your generated ent package
)

func Connect() {
	// client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	// if err != nil {
	// 	log.Fatalf("failed opening connection to sqlite: %v", err)
	// }
	// defer client.Close()
	// if err := client.Schema.Create(context.Background()); err != nil {
	// 	log.Fatalf("failed creating schema resources: %v", err)
	// }
}
{{else}}
func Connect() {
	// TODO: Implement database connection
}
{{end}}
