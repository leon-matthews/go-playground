// Package database opens the SQLite store and applies its schema.
package database

import (
	"context"
	_ "embed"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"pwneddb/database/sqlite"
)

//go:embed schema.sql
var schema string

// Open connects to the SQLite database at path, creating it if needed.
// The schema is applied on every open, so a fresh file is ready to use.
func Open(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	// Sequential access uses a single writer, avoiding SQLITE_BUSY
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("applying schema: %w", err)
	}

	return sqlite.New(db), db, nil
}
