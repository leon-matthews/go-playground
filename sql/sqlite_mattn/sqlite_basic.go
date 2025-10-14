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

/*
https://phiresky.github.io/blog/2020/sqlite-performance-tuning/

Good article, key points are:

pragma journal_mode = WAL;
pragma synchronous = normal;
pragma temp_store = memory;
pragma mmap_size = 30000000000; // tO tHE mOON!
*/

/*
https://www.reddit.com/r/golang/comments/16xswxd/comment/k34ppfo/

Here are my general tips regarding mattn's driver, which my team has used to
build SQLite-backed microservices:

1. Set the journal mode to WAL and synchronous to Normal.
2. Use two connections, one read-only with max open connections set to some
   large number, and one read-write set to a maximum of 1 open connection.
3. Set the transaction locking mode to IMMEDIATE and use transactions for
   any multi-query methods.
4. Set the busy timeout to some large value, like 5000. I'm not sure why
   this is necessary, since I figured the pool size of 1 would obviate the
   need for this, but it seems necessary (otherwise you can get database is
   locked errors).

With these few settings, we get good performance for our use case
(>2K mid-size writes/sec, 30K reads per second on 2 vCPU and an SSD). I'd
also recommend using Litestream to perform WAL shipping to S3.
*/

/*
https://github.com/mattn/go-sqlite3/issues/1022#issuecomment-1067353980

Note that you need to prefix your connection string with file: for the
various options to be interpreted properly. Also, you should not call
sql.Open once per worker, as sql.DB itself represents a pool, not an
individual connection.

My general recommendation is to make two pools (as in two sql.DBs), one
with mode=ro and one with mode=rw. Use wal mode (_journal_mode=wal),
which will allow reads to happen concurrently with writes. Do not use
shared cache mode. Throttle the read/write pool to a single connection
using SetMaxOpenConns, as SQLite doesn't support multiple concurrent
writers anyway. The read-only pool should be throttled as per your
application requirements and system constraints. The read/write pool
should also use BEGIN IMMEDITATE when starting transactions
(_txlock=immediate), to avoid certain issues that can result in "database
is locked" errors.
*/

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
	// Strict tables (from v3.37, Nov 2021)
	// https://www.sqlite.org/stricttables.html
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
