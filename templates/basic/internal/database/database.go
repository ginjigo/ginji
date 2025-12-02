package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error
	DB, err = gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}
