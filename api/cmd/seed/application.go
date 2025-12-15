package main

import (
	"log"
	"os"

	"gorm.io/gorm"
)

var app = &application{
	logger:    log.New(os.Stdout, "[seed] ", log.LstdFlags),
	errLogger: log.New(os.Stderr, "[seed] ERROR: ", log.LstdFlags),
}

type application struct {
	db        *gorm.DB
	logger    *log.Logger
	errLogger *log.Logger
}
