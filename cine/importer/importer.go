package importer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"

	"local.dev/cine/database"
	"local.dev/cine/reader"
)

// Import builds a cine database at path from the IMDb dataset files in dir,
// logging progress to logger.
//
// The database is built in a temporary file and renamed into place only once the
// whole import succeeds, so an interrupted or failed run leaves any database
// already at path untouched.
func Import(ctx context.Context, path, dir string, logger *log.Logger) error {
	logger.Info("building database", "path", path)
	start := time.Now()

	temp := path + ".tmp"
	if err := os.Remove(temp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing %s: %w", temp, err)
	}

	if err := buildInto(ctx, temp, dir, logger); err != nil {
		os.Remove(temp) // don't leave a half-built file behind
		return err
	}

	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("installing database: %w", err)
	}
	logger.Info("database built", "path", path, "took", time.Since(start).Round(time.Millisecond))
	return nil
}

// layer imports one dataset file within its own transaction and reports the
// number of rows written to its primary table.
type layer struct {
	name string
	run  func(*sql.Tx) (int64, error)
}

// buildInto opens a fresh database at temp and imports every layer into it, each
// file within its own transaction, logging per-file progress.
func buildInto(ctx context.Context, temp, dir string, logger *log.Logger) error {
	_, db, err := database.Open(ctx, temp)
	if err != nil {
		return err
	}
	defer db.Close()

	ratingsStart := time.Now()
	ratings, err := readRatings(dir)
	if err != nil {
		return err
	}
	logger.Info("ratings loaded", "count", len(ratings), "took", time.Since(ratingsStart).Round(time.Millisecond))

	// Each layer imports one file within its own transaction, in dependency order.
	layers := []layer{
		{"titles", func(tx *sql.Tx) (int64, error) { return importTitlesLayer(ctx, tx, dir, ratings) }},
		{"names", func(tx *sql.Tx) (int64, error) { return importNamesLayer(ctx, tx, dir) }},
		{"episodes", func(tx *sql.Tx) (int64, error) { return importEpisodesLayer(ctx, tx, dir) }},
		{"crew", func(tx *sql.Tx) (int64, error) { return importCrewLayer(ctx, tx, dir) }},
		{"principals", func(tx *sql.Tx) (int64, error) { return importPrincipalsLayer(ctx, tx, dir) }},
		{"akas", func(tx *sql.Tx) (int64, error) { return importAkasLayer(ctx, tx, dir) }},
	}
	for _, l := range layers {
		logger.Info("importing", "file", l.name)
		start := time.Now()
		count, err := inTx(ctx, db, l.run)
		if err != nil {
			return fmt.Errorf("%s: %w", l.name, err)
		}
		logger.Info("imported", "file", l.name, "rows", count,
			"took", time.Since(start).Round(time.Millisecond))
	}
	return nil
}

// inTx runs fn inside a transaction, committing on success and discarding the
// transaction (the whole build is thrown away) on failure. It passes through
// fn's row count.
func inTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) (int64, error)) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	count, err := fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
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
func importTitlesLayer(ctx context.Context, tx *sql.Tx, dir string, ratings map[int64]rating) (int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleBasics))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count, lookups, err := importTitles(ctx, tx, file, ratings)
	if err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "titles_types", lookups.titleType); err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "genres", lookups.genre); err != nil {
		return 0, err
	}
	return count, nil
}

// importNamesLayer runs the names pass within tx and writes its profession lookup.
func importNamesLayer(ctx context.Context, tx *sql.Tx, dir string) (int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileNameBasics))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count, profession, err := importNames(ctx, tx, file)
	if err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "professions", profession); err != nil {
		return 0, err
	}
	return count, nil
}

// importEpisodesLayer runs the episodes pass within tx.
func importEpisodesLayer(ctx context.Context, tx *sql.Tx, dir string) (int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleEpisode))
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return importEpisodes(ctx, tx, file)
}

// importCrewLayer runs the crew pass within tx.
func importCrewLayer(ctx context.Context, tx *sql.Tx, dir string) (int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleCrew))
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return importCrew(ctx, tx, file)
}

// importPrincipalsLayer runs the principals pass within tx and writes its
// category and job lookups.
func importPrincipalsLayer(ctx context.Context, tx *sql.Tx, dir string) (int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitlePrincipals))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count, lookups, err := importPrincipals(ctx, tx, file)
	if err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "principals_categories", lookups.category); err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "principals_jobs", lookups.job); err != nil {
		return 0, err
	}
	return count, nil
}

// importAkasLayer runs the akas pass within tx and writes its region, language,
// aka_type and attribute lookups.
func importAkasLayer(ctx context.Context, tx *sql.Tx, dir string) (int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleAkas))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count, lookups, err := importAkas(ctx, tx, file)
	if err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "regions", lookups.region); err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "languages", lookups.language); err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "akas_types", lookups.akaType); err != nil {
		return 0, err
	}
	if err := flushLookup(ctx, tx, "attributes", lookups.attribute); err != nil {
		return 0, err
	}
	return count, nil
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
