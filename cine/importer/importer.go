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

// buildInto opens a fresh database at temp and imports every layer into it, each
// file within its own transaction.
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

	// Each layer imports one file within its own transaction, in dependency order.
	layers := []func(*sql.Tx) error{
		func(tx *sql.Tx) error { return importTitlesLayer(ctx, tx, dir, ratings) },
		func(tx *sql.Tx) error { return importNamesLayer(ctx, tx, dir) },
		func(tx *sql.Tx) error { return importEpisodesLayer(ctx, tx, dir) },
		func(tx *sql.Tx) error { return importCrewLayer(ctx, tx, dir) },
		func(tx *sql.Tx) error { return importPrincipalsLayer(ctx, tx, dir) },
		func(tx *sql.Tx) error { return importAkasLayer(ctx, tx, dir) },
	}
	for _, layer := range layers {
		if err := inTx(ctx, db, layer); err != nil {
			return err
		}
	}
	return nil
}

// inTx runs fn inside a transaction, committing on success and discarding the
// transaction (the whole build is thrown away) on failure.
func inTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
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
	if err := flushLookup(ctx, tx, "titles_types", lookups.titleType); err != nil {
		return err
	}
	return flushLookup(ctx, tx, "genres", lookups.genre)
}

// importNamesLayer runs the names pass within tx and writes its profession lookup.
func importNamesLayer(ctx context.Context, tx *sql.Tx, dir string) error {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileNameBasics))
	if err != nil {
		return err
	}
	defer file.Close()

	profession, err := importNames(ctx, tx, file)
	if err != nil {
		return err
	}
	return flushLookup(ctx, tx, "professions", profession)
}

// importEpisodesLayer runs the episodes pass within tx.
func importEpisodesLayer(ctx context.Context, tx *sql.Tx, dir string) error {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleEpisode))
	if err != nil {
		return err
	}
	defer file.Close()
	return importEpisodes(ctx, tx, file)
}

// importCrewLayer runs the crew pass within tx.
func importCrewLayer(ctx context.Context, tx *sql.Tx, dir string) error {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleCrew))
	if err != nil {
		return err
	}
	defer file.Close()
	return importCrew(ctx, tx, file)
}

// importPrincipalsLayer runs the principals pass within tx and writes its
// category and job lookups.
func importPrincipalsLayer(ctx context.Context, tx *sql.Tx, dir string) error {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitlePrincipals))
	if err != nil {
		return err
	}
	defer file.Close()

	lookups, err := importPrincipals(ctx, tx, file)
	if err != nil {
		return err
	}
	if err := flushLookup(ctx, tx, "principals_categories", lookups.category); err != nil {
		return err
	}
	return flushLookup(ctx, tx, "principals_jobs", lookups.job)
}

// importAkasLayer runs the akas pass within tx and writes its region, language,
// aka_type and attribute lookups.
func importAkasLayer(ctx context.Context, tx *sql.Tx, dir string) error {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleAkas))
	if err != nil {
		return err
	}
	defer file.Close()

	lookups, err := importAkas(ctx, tx, file)
	if err != nil {
		return err
	}
	if err := flushLookup(ctx, tx, "regions", lookups.region); err != nil {
		return err
	}
	if err := flushLookup(ctx, tx, "languages", lookups.language); err != nil {
		return err
	}
	if err := flushLookup(ctx, tx, "akas_types", lookups.akaType); err != nil {
		return err
	}
	return flushLookup(ctx, tx, "attributes", lookups.attribute)
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
