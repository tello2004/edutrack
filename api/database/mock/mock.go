package mock

import (
	"gorm.io/gorm"
	"lahuerta.tecmm.edu.mx/edutrack/database/sqlite"
)

// Open initializes a DB session based on an in-memory SQLite database.
func Open(dsn string) (*gorm.DB, error) {
	db, err := sqlite.Open("file:mock?mode=memory&cache=shared")
	if err != nil {
		return db, err
	}

	return db, nil
}
