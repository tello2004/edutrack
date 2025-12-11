//go:build !sqlite

package main

import (
	"lahuerta.tecmm.edu.mx/edutrack/database/mock"
)

func init() {
	db, err := mock.Open("mock")
	if err != nil {
		app.errLogger.Fatalf("edutrackd: %s", err)
	}

	app.db = db
}
