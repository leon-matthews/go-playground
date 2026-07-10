// Package database opens the SQLite stores this tool reads and writes.
//
// We read from a database of SHA-1 hashes created by our sibling
// tool `pwnedcache`. Ther are in the table `hashes` in a separate SQLite file
// that we mount read-only.
//
// We write results to the `passwords` table that we maintain with this tool
// in its own SQLite file.
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

// readPragmas tune the read-only cache for its scan-and-point-lookup workload.
var readPragmas = []string{
	"cache_size(-65536)",      // 64MiB page cache
	"locking_mode(exclusive)", // static single-process file: skip per-query lock syscalls
}

// writePragmas tune the writable database for its single-connection workload.
var writePragmas = []string{
	"cache_size(-65536)",      // 64MiB disk cache
	"journal_mode(WAL)",       // WAL keeps the file crash-consistent
	"locking_mode(EXCLUSIVE)", // Keep the wal-index in heap (no -shm file)
	"synchronous(NORMAL)",     // Drop the per-commit fsync
	"temp_store(MEMORY)",      // Keep temp tables etc. in RAM
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

// OpenRO opens the database file optimised for read-only access, capping the
// connection pool at maxConns.
func OpenRO(ctx context.Context, path string, maxConns int) (*sqlite.Queries, *sql.DB, error) {
	// The file: scheme is required for modernc to honour the mode=ro URI param
	db, err := sql.Open("sqlite", "file:"+path+pragmaQuery(readPragmas...)+"&mode=ro")
	if err != nil {
		return nil, nil, fmt.Errorf("opening database read-only %q: %w", path, err)
	}

	// Cap the pool at the caller's concurrency; readers share the file via SHARED locks
	db.SetMaxOpenConns(maxConns)

	// sql.Open is lazy, so ping now to fail early on a missing or bad file
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("opening database read-only %q: %w", path, err)
	}
	return sqlite.New(db), db, nil
}
