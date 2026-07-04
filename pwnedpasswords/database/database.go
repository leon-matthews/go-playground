// Package database opens the SQLite stores this tool reads and writes.
package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"

	// Registers the "sqlite" driver with database/sql.
	_ "modernc.org/sqlite"

	"pwnedpasswords/database/sqlite"
)

//go:embed schema.sql
var schema string

// writePragmas tune the writable database for its single-connection workload.
var writePragmas = []string{
	"journal_mode(WAL)",
	"synchronous(NORMAL)",
	"busy_timeout(5000)",
}

// pragmaQuery builds the modernc DSN query that runs each pragma once on open.
func pragmaQuery(pragmas ...string) string {
	values := url.Values{}
	for _, pragma := range pragmas {
		values.Add("_pragma", pragma)
	}
	return "?" + values.Encode()
}

// Open opens the writable pwnedpasswords database, creating it if needed.
// The schema is applied on every open, so a fresh file is ready to use.
func Open(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	db, err := sql.Open("sqlite", path+pragmaQuery(writePragmas...))
	if err != nil {
		return nil, nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	// One connection is the lone writer, so no lock contention arises
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("applying schema: %w", err)
	}
	return sqlite.New(db), db, nil
}

// OpenCache opens the pwnedcache database read-only.
// The file is treated as a static snapshot and is never modified.
func OpenCache(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	// The file: scheme is required for modernc to honour the mode=ro URI param
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, nil, fmt.Errorf("opening cache %q: %w", path, err)
	}

	// sql.Open is lazy, so ping now to fail early on a missing or bad file
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("opening cache %q: %w", path, err)
	}
	return sqlite.New(db), db, nil
}
