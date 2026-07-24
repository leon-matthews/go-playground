package importer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"local.dev/cine/database"
	"local.dev/cine/reader"
)

// Import builds a cine database at path from the IMDb dataset files in dir.
//
// The database is built in a temporary file and renamed into place only once the
// whole import succeeds, so an interrupted or failed run leaves any database
// already at path untouched.
func Import(ctx context.Context, path, dir string) error {
	temp := path + ".tmp"
	if err := os.Remove(temp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing %s: %w", temp, err)
	}

	if err := buildInto(ctx, temp, dir); err != nil {
		os.Remove(temp) // don't leave a half-built file behind
		return err
	}

	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("installing database: %w", err)
	}
	return nil
}

// buildInto opens a fresh database at temp and imports every layer into it.
func buildInto(ctx context.Context, temp, dir string) error {
	_, db, err := database.Open(ctx, temp)
	if err != nil {
		return err
	}
	defer db.Close()

	ratings, err := readRatings(dir)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	if err := importTitlesLayer(ctx, tx, dir, ratings); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing titles: %w", err)
	}
	return nil
}

// readRatings loads title.ratings from dir into a lookup map.
func readRatings(dir string) (map[int64]rating, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleRatings))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return loadRatings(file)
}

// importTitlesLayer runs the titles pass within tx and writes its lookup tables.
func importTitlesLayer(ctx context.Context, tx *sql.Tx, dir string, ratings map[int64]rating) error {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleBasics))
	if err != nil {
		return err
	}
	defer file.Close()

	lookups, err := importTitles(ctx, tx, file, ratings)
	if err != nil {
		return err
	}
	if err := flushLookup(ctx, tx, "title_type", lookups.titleType); err != nil {
		return err
	}
	return flushLookup(ctx, tx, "genre", lookups.genre)
}

// flushLookup writes an interner's entries into its two-column lookup table.
func flushLookup(ctx context.Context, tx *sql.Tx, table string, in *interner) error {
	inserter, err := newBatchInserter(ctx, tx, table, []string{"id", "name"},
		func(args []any, e lookupEntry) []any { return append(args, e.id, e.name) })
	if err != nil {
		return err
	}
	for _, entry := range in.entries() {
		if err := inserter.Add(ctx, entry); err != nil {
			return err
		}
	}
	return inserter.Flush(ctx)
}
