package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open initializes a PostgreSQL database connection.
// The dsn should be a valid PostgreSQL connection string, for example:
// "host=localhost user=edutrack password=secret dbname=edutrack port=5432 sslmode=disable"
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return db, err
	}

	return db, nil
}
