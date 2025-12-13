package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Open initializes an SQLite database.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return db, err
	}

	return db, nil
}
