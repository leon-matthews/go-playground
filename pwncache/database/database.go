// Package database opens the SQLite store and applies its schema.
package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"

	// Registers the "sqlite" driver with database/sql
	_ "modernc.org/sqlite"

	"pwncache/database/sqlite"
)

//go:embed schema.sql
var schema string

// bulkLoadPragmas tunes SQLite for a single-writer bulk load. WAL keeps the
// file crash-consistent, synchronous=NORMAL drops the per-commit fsync,
// EXCLUSIVE locking keeps the wal-index in heap (no -shm file), and a 256 MB
// page cache holds the write working set. The driver sorts pragmas before
// running them, which lands EXCLUSIVE after WAL but before the first write.
var bulkLoadPragmas = []string{
	"journal_mode(WAL)",
	"synchronous(NORMAL)",
	"locking_mode(EXCLUSIVE)",
	"cache_size(-262144)",
	"temp_store(MEMORY)",
}

// pragmaQuery builds the modernc-sqlite DSN query that runs each pragma once on
// connection open.
func pragmaQuery(pragmas ...string) string {
	values := url.Values{}
	for _, pragma := range pragmas {
		values.Add("_pragma", pragma)
	}
	return "?" + values.Encode()
}

// Open connects to the SQLite database at path, creating it if needed.
// The schema is applied on every open, so a fresh file is ready to use.
func Open(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	db, err := sql.Open("sqlite", path+pragmaQuery(bulkLoadPragmas...))
	if err != nil {
		return nil, nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	// A single connection is both the lone writer and mandatory under
	// EXCLUSIVE locking, which lets only one connection touch the file
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("applying schema: %w", err)
	}

	// Prepare every query once; closing db later also closes the statements
	queries, err := sqlite.Prepare(ctx, db)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("preparing queries: %w", err)
	}
	return queries, db, nil
}
