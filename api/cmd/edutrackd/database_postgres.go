//go:build postgres

package main

import (
	"log"
	"os"

	"lahuerta.tecmm.edu.mx/edutrack/database/postgres"
)

func init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=edutrack password=edutrack dbname=edutrack port=5432 sslmode=disable"
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("edutrackd: failed to open database: %s", err)
	}

	app.db = db
}
