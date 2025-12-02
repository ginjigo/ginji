module {{.Name}}

go 1.23.0

require (
	github.com/ginjigo/ginji v0.1.9
{{if eq .ORM "GORM"}}
	gorm.io/gorm v1.25.12
{{if eq .Database "SQLite"}}	gorm.io/driver/sqlite v1.5.6{{end}}
{{if eq .Database "PostgreSQL"}}	gorm.io/driver/postgres v1.5.9{{end}}
{{if eq .Database "MySQL"}}	gorm.io/driver/mysql v1.5.7{{end}}
{{end}}
{{if eq .ORM "sqlc"}}
	github.com/lib/pq v1.10.9
{{end}}
{{if eq .ORM "ent"}}
	entgo.io/ent v0.12.5
	github.com/mattn/go-sqlite3 v1.14.19
{{end}}
)
