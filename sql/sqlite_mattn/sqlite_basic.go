// Basic SQLite example using mattn's driver
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

func main() {
	// Connect to DB
	ctx := context.Background()
	db, err := openDB(ctx, "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	err = createTables(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	// Insert rows
	prefixes := []string{
		"00000", "00001", "00002", "00003", "00004", "00005", "00006", "00007",
		"00008", "00009", "0000a", "0000b", "0000c", "0000d", "0000e", "0000f",
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, prefix := range prefixes {
		if err := insertPrefix(ctx, db, prefix); err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	// Select rows
	found, err := selectPrefixes(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range found {
		fmt.Printf("%s ", p.prefix)
	}
	fmt.Println()

	// Backup database
	backup, err := openDB(ctx, "db.backup.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer backup.Close()

	// TODO This is broken. No error occurs but we get two empty databases!
	err = backupDB(ctx, db, backup)
	if err != nil {
		log.Fatal(err)
	}
}

type Prefix struct {
	id      uint
	prefix  string
	updated float64
}

// createTables creates tables, but only if necessary
func createTables(ctx context.Context, db *sql.DB) error {
	s := `
CREATE TABLE IF NOT EXISTS prefixes (
    id      INTEGER PRIMARY KEY,
    prefix  TEXT NOT NULL UNIQUE,
    updated REAL NOT NULL       -- Unix epoch
) STRICT;
	`
	_, err := db.ExecContext(ctx, s)
	if err != nil {
		err = fmt.Errorf("create tables: %w", err)
	}
	return err
}

// insertPrefix adds an entry to the prefixes table
func insertPrefix(ctx context.Context, db *sql.DB, prefix string) error {
	if len(prefix) != 5 {
		return fmt.Errorf("insert prefixes: prefix wrong length")
	}
	updated := float64(time.Now().UnixNano()) / 1e9
	query := "INSERT OR REPLACE INTO prefixes (prefix, updated) VALUES ($1, $2)"
	_, err := db.ExecContext(ctx, query, prefix, updated)
	if err != nil {
		err = fmt.Errorf("insert prefixes: %w", err)
	}
	return nil
}

// openDB opens SQLite3 file, creating it as necessary
// source can be a filename or the SQLite-specific :memory:
func openDB(ctx context.Context, source string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return db, nil
}

// selectPrefixes returns all prefixes records from the db
func selectPrefixes(ctx context.Context, db *sql.DB) ([]Prefix, error) {
	rows, err := db.Query("SELECT * FROM prefixes")
	if err != nil {
		return nil, fmt.Errorf("select prefixes: %w", err)
	}
	defer rows.Close()

	var prefixes []Prefix
	for rows.Next() {
		var p Prefix
		if err := rows.Scan(&p.id, &p.prefix, &p.updated); err != nil {
			return nil, fmt.Errorf("select prefixes: %w", err)
		}
		prefixes = append(prefixes, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select prefixes: %w", err)
	}
	return prefixes, nil
}

// backupDB using the SQLite backup API to safely create a copy of fromDB.
// The destination database (toDB) must be completely empty. Either can be in-memory.
// Order of arguments matches io.Copy, following 'to = from' assignment mnemonic
func backupDB(ctx context.Context, destination, source *sql.DB) error {
	// Start direct connections
	toConn, err := destination.Conn(ctx)
	if err != nil {
		return err
	}

	fromConn, err := source.Conn(ctx)
	if err != nil {
		return err
	}

	// Run backup calls inside nested Raw() calls
	err = toConn.Raw(func(toConn any) error {
		return fromConn.Raw(func(fromConn any) error {
			// Convert for SQLite-specific functionality
			toConnSQLite, ok := toConn.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("convert destination connection to SQLiteConn")
			}

			fromConnSQLite, ok := fromConn.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("convert source connection to SQLiteConn")
			}

			// Actually run the backup command
			b, err := toConnSQLite.Backup("main", fromConnSQLite, "main")
			if err != nil {
				return fmt.Errorf("starting SQLite backup: %w", err)
			}

			// Copy the whole DB in one step
			done, err := b.Step(-1)
			if !done {
				return fmt.Errorf("backup step not done")
			}
			if err != nil {
				return fmt.Errorf("stepping backup: %w", err)
			}

			err = b.Finish()
			if err != nil {
				return fmt.Errorf("finishing backup: %w", err)
			}
			return err
		})
	})

	return err
}
