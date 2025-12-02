//go:build sqlite

package main

import (
	"log"
	"os"

	"lahuerta.tecmm.edu.mx/edutrack/database/sqlite"
)

func init() {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sqlite.Open(dsn)
	if err != nil {
		log.Fatalf("edutrackd: %s", err)
	}

	app.db = db
}
