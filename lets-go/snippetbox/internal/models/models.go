// Package models provides the database storage layer
package models

import (
	"database/sql"
)

// OpenDB wraps sql.Open() and returns an sql.DB connection poll
// Be sure to call its Close method
func OpenDB(dsn string) (*sql.DB, error) {
	// Create pool
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Actually make connection
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
