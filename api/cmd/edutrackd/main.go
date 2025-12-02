package main

import (
	"log"

	"gorm.io/gorm"
)

var app *application

type application struct {
	db        *gorm.DB
	logger    *log.Logger
	errLogger *log.Logger
}

func main() {
	defer func() {
		db, _ := app.db.DB()
		_ = db.Close()
	}()
}
