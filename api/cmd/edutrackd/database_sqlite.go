//go:build sqlite

package main

import (
	"log"
	"os"

	"lahuerta.tecmm.edu.mx/edutrack/database/sqlite"
)

func init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "edutrack.db"
	}

	db, err := sqlite.Open(dsn)
	if err != nil {
		log.Fatalf("edutrackd: failed to open database: %s", err)
	}

	app.db = db
}
