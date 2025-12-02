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
)
