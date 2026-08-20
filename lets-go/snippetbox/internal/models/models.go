// Package models provides the database storage layer
package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

// OpenDB wraps sql.Open() and returns an sql.DB connection pool.
// Be sure to call its Close method
func OpenDB(ctx context.Context, dsn string) (*sql.DB, error) {
	// Create pool
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Actually make connection
	err = PingRetry(ctx, db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// PingRetry keeps attempting to ping db until the context's deadline is reached.
// It's an error to provide a context without a deadline
func PingRetry(ctx context.Context, db *sql.DB) error {
	// Ensure ctx has a deadline
	if _, ok := ctx.Deadline(); !ok {
		return fmt.Errorf("context deadline not set")
	}

	// Disable noisy mysql stderr-logging
	mysql.SetLogger(log.New(io.Discard, "", 0))

	// Use zero value for an immediate check
	var naptime time.Duration
	var dbError error
	for {
		select {
		case <-ctx.Done():
			err := errors.Join(ctx.Err(), dbError)
			return fmt.Errorf("ping db: %w", err)

		case <-time.After(naptime):
			dbError = db.PingContext(ctx)
			if dbError == nil {
				return nil
			}
			naptime = 100 * time.Millisecond
		}
	}
}
