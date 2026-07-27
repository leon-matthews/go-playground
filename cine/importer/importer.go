package importer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"

	"local.dev/cine/common"
	"local.dev/cine/database"
	"local.dev/cine/reader"
)

// Version is the cine release recorded in each build's build_info row.
const Version = "0.1.0"

// layerFull is the base layer recorded in build_info.layer. It is the only one
// produced so far, because buildInto imports every dataset file.
const layerFull = 2

// Import builds a cine database at path from the IMDb dataset files in dir,
// logging progress to logger.
//
// The database is built in a temporary file and renamed into place only once the
// whole import succeeds, so an interrupted or failed run leaves any database
// already at path untouched.
func Import(ctx context.Context, path, dir string, rules FilterRules, logger *log.Logger) error {
	logger.Info("building database", "path", path, "filter", rules)
	start := time.Now()

	temp := path + ".tmp"
	if err := os.Remove(temp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing %s: %w", temp, err)
	}

	if err := buildInto(ctx, temp, dir, rules, logger); err != nil {
		os.Remove(temp) // don't leave a half-built file behind
		return err
	}

	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("installing database: %w", err)
	}

	fields := []any{"path", path, "took", time.Since(start).Round(time.Millisecond)}
	if info, err := os.Stat(path); err == nil {
		fields = append(fields, "size", common.Bytes(info.Size()))
	}
	logger.Info("database built", fields...)
	return nil
}

// counts reports what one import pass consumed and produced. The two differ
// wherever a pass fans out: each title.crew row becomes one credit per person.
type counts struct {
	read    int64 // source rows consumed from the file
	written int64 // rows written to the pass's primary table
}

// layer imports one dataset file within its own transaction.
type layer struct {
	file string // canonical dataset file name, for logging and errors
	run  func(*sql.Tx) (counts, error)
}

// sourceRow is one consumed dataset file, destined for build_sources.
type sourceRow struct {
	file         string
	lastModified string
	bytes        int64
	rowsRead     int64
}

// buildInto opens a fresh database at temp and imports every layer into it, each
// file within its own transaction, logging per-file progress.
func buildInto(ctx context.Context, temp, dir string, rules FilterRules, logger *log.Logger) error {
	started := time.Now()
	_, db, err := database.Open(ctx, temp)
	if err != nil {
		return err
	}
	defer db.Close()

	ratingsStart := time.Now()
	ratings, ratingsRead, err := readRatings(dir)
	if err != nil {
		return err
	}
	logger.Info("ratings loaded",
		"file", reader.FileTitleRatings,
		"count", common.Commas(int64(len(ratings))),
		"took", time.Since(ratingsStart).Round(time.Millisecond))

	// title.ratings has no table of its own, so record it before the layers.
	sources, err := appendSource(nil, dir, reader.FileTitleRatings, ratingsRead)
	if err != nil {
		return err
	}

	// The allow-list is settled in full before any layer writes, so every pass sees
	// the same one and none has to accumulate it for the next.
	filterStart := time.Now()
	filter, err := buildFilter(dir, rules, ratings)
	if err != nil {
		return err
	}
	if kept, filtering := filter.size(); filtering {
		logger.Info("filter built",
			"rules", rules,
			"titles", common.Commas(int64(kept)),
			"took", time.Since(filterStart).Round(time.Millisecond))
	}

	// Each layer imports one file within its own transaction, in dependency order.
	// Every layer takes the filter; names needs it only for its known-for junction.
	layers := []layer{
		{reader.FileTitleBasics, func(tx *sql.Tx) (counts, error) {
			return importTitlesLayer(ctx, tx, dir, ratings, filter)
		}},
		{reader.FileNameBasics, func(tx *sql.Tx) (counts, error) { return importNamesLayer(ctx, tx, dir, filter) }},
		{reader.FileTitleEpisode, func(tx *sql.Tx) (counts, error) { return importEpisodesLayer(ctx, tx, dir, filter) }},
		{reader.FileTitleCrew, func(tx *sql.Tx) (counts, error) { return importCrewLayer(ctx, tx, dir, filter) }},
		{reader.FileTitlePrincipals, func(tx *sql.Tx) (counts, error) {
			return importPrincipalsLayer(ctx, tx, dir, filter)
		}},
		{reader.FileTitleAkas, func(tx *sql.Tx) (counts, error) { return importAkasLayer(ctx, tx, dir, filter) }},
	}
	for _, l := range layers {
		start := time.Now()
		var count counts
		err := inTx(ctx, db, func(tx *sql.Tx) error {
			var err error
			count, err = l.run(tx)
			return err
		})
		if err != nil {
			return fmt.Errorf("%s: %w", l.file, err)
		}
		if sources, err = appendSource(sources, dir, l.file, count.read); err != nil {
			return err
		}
		elapsed := time.Since(start)
		// Both counts are logged: they part company wherever a pass filters or fans out.
		logger.Info("imported",
			"file", l.file,
			"read", common.Commas(count.read),
			"rows", common.Commas(count.written),
			"took", elapsed.Round(time.Millisecond),
			"rate", ratePerSecond(count.written, elapsed))
	}

	return inTx(ctx, db, func(tx *sql.Tx) error {
		return writeBuildMetadata(ctx, tx, sources, rules, started, time.Now())
	})
}

// appendSource stats one consumed dataset file and appends its provenance.
//
// A file's own timestamp is the only provenance a build can keep: wget copies it
// from the server, and the seven files carry timestamps hours apart, so there is
// no single date for a download.
func appendSource(sources []sourceRow, dir, file string, rowsRead int64) ([]sourceRow, error) {
	info, err := os.Stat(filepath.Join(dir, file))
	if err != nil {
		return nil, fmt.Errorf("recording source %s: %w", file, err)
	}
	return append(sources, sourceRow{
		file:         file,
		lastModified: timestamp(info.ModTime()),
		bytes:        info.Size(),
		rowsRead:     rowsRead,
	}), nil
}

// buildSourceColumns are the build_sources columns in the order bindSourceRow
// writes them.
var buildSourceColumns = []string{"file", "last_modified", "bytes", "rows_read"}

// bindSourceRow appends a build_sources row's values in buildSourceColumns order.
func bindSourceRow(args []any, s sourceRow) []any {
	return append(args, s.file, s.lastModified, s.bytes, s.rowsRead)
}

// writeBuildMetadata records what the build consumed and when it ran. It runs
// last, because a build that fails is discarded whole and has nothing to say.
func writeBuildMetadata(ctx context.Context, tx *sql.Tx, sources []sourceRow, rules FilterRules, started, finished time.Time) error {
	inserter, err := newBatchInserter(ctx, tx, "build_sources", buildSourceColumns, bindSourceRow)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if err := inserter.Add(ctx, source); err != nil {
			return err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	const insert = `INSERT INTO build_info
		(id, layer, cine_version, started_at, finished_at, filter_rated, filter_not_adult)
		VALUES (1, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, insert,
		layerFull, Version, timestamp(started), timestamp(finished),
		boolToInt(rules.Rated), boolToInt(rules.NotAdult)); err != nil {
		return fmt.Errorf("writing build_info: %w", err)
	}
	return nil
}

// timestamp formats an instant the way the metadata tables store one.
func timestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
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

// ratePerSecond formats an import throughput as comma-grouped rows per second.
func ratePerSecond(count int64, d time.Duration) string {
	secs := d.Seconds()
	if secs <= 0 {
		return common.Commas(count) + "/s"
	}
	return common.Commas(int64(float64(count)/secs)) + "/s"
}

// readRatings loads title.ratings from dir into a lookup map, reporting how many
// source rows it read.
func readRatings(dir string) (map[int64]rating, int64, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleRatings))
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	return loadRatings(file)
}

// importTitlesLayer runs the titles pass within tx and writes its lookup tables.
func importTitlesLayer(ctx context.Context, tx *sql.Tx, dir string, ratings map[int64]rating, filter titleFilter) (counts, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleBasics))
	if err != nil {
		return counts{}, err
	}
	defer file.Close()

	count, lookups, err := importTitles(ctx, tx, file, ratings, filter)
	if err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "titles_types", lookups.titleType); err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "genres", lookups.genre); err != nil {
		return counts{}, err
	}
	return count, nil
}

// importNamesLayer runs the names pass within tx and writes its profession lookup.
func importNamesLayer(ctx context.Context, tx *sql.Tx, dir string, filter titleFilter) (counts, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileNameBasics))
	if err != nil {
		return counts{}, err
	}
	defer file.Close()

	count, profession, err := importNames(ctx, tx, file, filter)
	if err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "professions", profession); err != nil {
		return counts{}, err
	}
	return count, nil
}

// importEpisodesLayer runs the episodes pass within tx.
func importEpisodesLayer(ctx context.Context, tx *sql.Tx, dir string, filter titleFilter) (counts, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleEpisode))
	if err != nil {
		return counts{}, err
	}
	defer file.Close()
	return importEpisodes(ctx, tx, file, filter)
}

// importCrewLayer runs the crew pass within tx.
func importCrewLayer(ctx context.Context, tx *sql.Tx, dir string, filter titleFilter) (counts, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleCrew))
	if err != nil {
		return counts{}, err
	}
	defer file.Close()
	return importCrew(ctx, tx, file, filter)
}

// importPrincipalsLayer runs the principals pass within tx and writes its
// category and job lookups.
func importPrincipalsLayer(ctx context.Context, tx *sql.Tx, dir string, filter titleFilter) (counts, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitlePrincipals))
	if err != nil {
		return counts{}, err
	}
	defer file.Close()

	count, lookups, err := importPrincipals(ctx, tx, file, filter)
	if err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "principals_categories", lookups.category); err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "principals_jobs", lookups.job); err != nil {
		return counts{}, err
	}
	return count, nil
}

// importAkasLayer runs the akas pass within tx and writes its region, language,
// aka_type and attribute lookups.
func importAkasLayer(ctx context.Context, tx *sql.Tx, dir string, filter titleFilter) (counts, error) {
	file, err := reader.OpenGzip(filepath.Join(dir, reader.FileTitleAkas))
	if err != nil {
		return counts{}, err
	}
	defer file.Close()

	count, lookups, err := importAkas(ctx, tx, file, filter)
	if err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "regions", lookups.region); err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "languages", lookups.language); err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "akas_types", lookups.akaType); err != nil {
		return counts{}, err
	}
	if err := flushLookup(ctx, tx, "attributes", lookups.attribute); err != nil {
		return counts{}, err
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
