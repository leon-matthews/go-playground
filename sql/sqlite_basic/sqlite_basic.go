// Basic SQL example
package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func main() {
	// Connect to DB
	ctx := context.Background()
	db, err := openDB(ctx)
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
	for _, prefix := range prefixes {
		if err := insertPrefix(ctx, db, prefix); err != nil {
			log.Fatal(err)
		}
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
func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "db.sqlite3")
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
