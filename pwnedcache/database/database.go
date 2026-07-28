// Package database opens the SQLite store and applies its schema.
package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"

	// Registers the "sqlite" driver with database/sql
	_ "modernc.org/sqlite"

	"pwnedcache/database/sqlite"
)

//go:embed schema.sql
var schema string

// ErrNotFound reports that no database exists at the requested path.
var ErrNotFound = errors.New("database not found")

// bulkLoadPragmas tunes SQLite for a single-writer bulk load.
var bulkLoadPragmas = []string{
	"journal_mode(WAL)",       // WAL keeps the file crash-consistent
	"synchronous(NORMAL)",     // Drop the per-commit fsync
	"locking_mode(EXCLUSIVE)", // Keep the wal-index in heap (no -shm file)
	"cache_size(-65536)",      // 64MiB disk cache
	"temp_store(MEMORY)",      // Keep temp tables etc. in RAM
}

// pragmaOptions builds the modernc-sqlite DSN options that run each pragma once
// on connection open.
func pragmaOptions(pragmas ...string) url.Values {
	options := url.Values{}
	for _, pragma := range pragmas {
		options.Add("_pragma", pragma)
	}
	return options
}

// buildDSN builds the modernc-sqlite DSN addressing path with the given options.
// The file: form is what lets an option such as mode=ro take effect, and it in
// turn means escaping any character a URI would otherwise read as syntax.
func buildDSN(path string, options url.Values) string {
	return "file:" + (&url.URL{Path: path}).EscapedPath() + "?" + options.Encode()
}

// Open connects to the SQLite database at path, creating it if needed.
// The schema is applied on every open, so a fresh file is ready to use.
func Open(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	db, err := sql.Open("sqlite", buildDSN(path, pragmaOptions(bulkLoadPragmas...)))
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

// OpenReadOnly connects to an existing SQLite database at path.
// Unlike [Open] it neither creates the file nor applies the schema, so a
// missing database reports ErrNotFound instead of querying as empty.
func OpenReadOnly(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	// Checked up front so a missing file gets a clearer error than SQLite gives
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("%w: %q", ErrNotFound, path)
		}
		return nil, nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	db, err := sql.Open("sqlite", buildDSN(path, url.Values{"mode": {"ro"}}))
	if err != nil {
		return nil, nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	// Queries run one at a time, so further connections would earn nothing
	db.SetMaxOpenConns(1)

	queries, err := sqlite.Prepare(ctx, db)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("preparing queries: %w", err)
	}
	return queries, db, nil
}
