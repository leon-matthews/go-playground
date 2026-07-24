// Package database opens the cine SQLite store and applies its schema.
package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"

	// Registers the "sqlite" driver with database/sql
	_ "modernc.org/sqlite"

	"local.dev/cine/database/sqlite"
)

//go:embed schema.sql
var schema string

// bulkLoadPragmas tune SQLite for the single-writer bulk rebuild. The importer
// builds a throwaway file that is renamed into place only once it succeeds, so
// durability of the build buys nothing - a crash just discards the file.
var bulkLoadPragmas = []string{
	"journal_mode(OFF)",       // No journal: the whole database is one file to rename
	"synchronous(OFF)",        // No fsync: a failed build is discarded, not recovered
	"locking_mode(EXCLUSIVE)", // Sole writer; keeps the wal-index in heap, no -shm file
	"cache_size(-65536)",      // 64MiB page cache
	"temp_store(MEMORY)",      // Keep temp tables and indexes in RAM
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
//
// The schema is applied on open and the connection is tuned for the bulk
// rebuild, so point Open at a throwaway file that is renamed into place once the
// import succeeds.
func Open(ctx context.Context, path string) (*sqlite.Queries, *sql.DB, error) {
	db, err := sql.Open("sqlite", path+pragmaQuery(bulkLoadPragmas...))
	if err != nil {
		return nil, nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	// One connection is both the lone writer and mandatory under EXCLUSIVE locking
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
