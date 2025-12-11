package main

import (
	"log"
	"os"

	"gorm.io/gorm"
	edutrack "lahuerta.tecmm.edu.mx/edutrack"
	"lahuerta.tecmm.edu.mx/edutrack/http"
)

var app = &application{
	logger:    log.New(os.Stdout, "[edutrackd] ", log.LstdFlags),
	errLogger: log.New(os.Stderr, "[edutrackd] ERROR: ", log.LstdFlags),
}

type application struct {
	db        *gorm.DB
	logger    *log.Logger
	errLogger *log.Logger
	server    *http.Server
	edutrack  *edutrack.App
}
