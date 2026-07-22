package export

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"

	"pwnedpasswords/database"
	"pwnedpasswords/database/sqlite"
	"pwnedpasswords/logging"
	"pwnedpasswords/progress"
)

// Commit the merge in batches this size, checkpointing the WAL after each, so a
// large restore neither holds one giant transaction nor lets the WAL grow unbounded.
const mergeBatchRows = 10_000

// mergeCounters tallies a merge run. Its fields are atomic because the reporter
// goroutine reads them while the merge loop updates them.
type mergeCounters struct {
	read      atomic.Int64           // valid rows parsed (merged + skipped)
	merged    atomic.Int64           // rows inserted
	skipped   atomic.Int64           // rows whose password was already present
	malformed atomic.Int64           // rows dropped with a warning
	sample    atomic.Pointer[string] // most recent valid row, for display
}

// recordMerged counts one freshly inserted password.
func (m *mergeCounters) recordMerged(password string) {
	m.read.Add(1)
	m.merged.Add(1)
	m.sample.Store(&password)
}

// recordSkipped counts one password that was already present.
func (m *mergeCounters) recordSkipped(password string) {
	m.read.Add(1)
	m.skipped.Add(1)
	m.sample.Store(&password)
}

// recordMalformed counts one dropped row.
func (m *mergeCounters) recordMalformed() {
	m.malformed.Add(1)
}

// mergeSnapshot is a point-in-time read of the counters, used for per-file deltas.
type mergeSnapshot struct {
	read, merged, skipped, malformed int64
}

// snapshot reads the four tallies together.
func (m *mergeCounters) snapshot() mergeSnapshot {
	return mergeSnapshot{
		read:      m.read.Load(),
		merged:    m.merged.Load(),
		skipped:   m.skipped.Load(),
		malformed: m.malformed.Load(),
	}
}

// reportTo returns a report function that logs the tallies as a friendly console
// line and the matching structured record to the file.
func (m *mergeCounters) reportTo(console, file *slog.Logger) func(kind string) {
	return func(kind string) {
		read, merged := m.read.Load(), m.merged.Load()
		skipped, malformed := m.skipped.Load(), m.malformed.Load()
		sample := ""
		if s := m.sample.Load(); s != nil {
			sample = *s
		}
		file.Info(
			kind,
			"read", read,
			"merged", merged,
			"skipped_existing", skipped,
			"skipped_malformed", malformed,
		)
		console.Info(humanMergeProgress(kind, read, merged, skipped, malformed, sample))
	}
}

// humanMergeProgress renders a friendly one-line progress or summary message.
func humanMergeProgress(kind string, read, merged, skipped, malformed int64, sample string) string {
	line := fmt.Sprintf(
		"%s read > %s merged > %s existing > %s malformed",
		humanize.Comma(read),
		humanize.Comma(merged),
		humanize.Comma(skipped),
		humanize.Comma(malformed),
	)
	if sample != "" {
		line += " > current: " + sample
	}
	if kind == "summary" {
		return "Finished: " + line
	}
	return line
}

// Merge streams one or more CSVs of "password,count" rows into the output
// database, inserting each password that is not already present and leaving
// existing counts untouched. Every input is opened up front so a bad path fails
// before any heavy work begins. Malformed rows are skipped with a warning rather
// than aborting the run, so one bad line cannot spoil a long restore.
func Merge(ctx context.Context, logs logging.Logging, dbPath string, csvPaths []string, interval time.Duration) (err error) {
	_, db, err := database.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	inputs, closeInputs, err := openMergeInputs(csvPaths)
	if err != nil {
		return err
	}
	defer closeInputs()

	var counters mergeCounters
	rep := progress.StartReporter(interval, counters.reportTo(logs.Console, logs.File))
	defer rep.StopAndReport()

	// Writes use a cancellation-free context so an interrupt commits the rows read
	// so far cleanly instead of tearing down the open transaction; the loops below
	// honour ctx themselves, and merge is additive so a resumed run skips them.
	dbCtx := context.WithoutCancel(ctx)

	for _, in := range inputs {
		if ctx.Err() != nil {
			break
		}
		slog.Info("merging CSV into database", "csv", in.name, "database", dbPath)
		before := counters.snapshot()
		if err := mergeFile(ctx, dbCtx, db, &counters, in.reader); err != nil {
			return fmt.Errorf("merging %q: %w", in.name, err)
		}
		after := counters.snapshot()
		slog.Info("merged CSV", "csv", in.name,
			"read", after.read-before.read,
			"merged", after.merged-before.merged,
			"skipped_existing", after.skipped-before.skipped,
			"skipped_malformed", after.malformed-before.malformed)
	}

	_, err = db.ExecContext(dbCtx, "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// mergeInput pairs an opened reader with a display name for logging and errors.
type mergeInput struct {
	name   string
	reader io.Reader
}

// openMergeInputs opens every path before the merge begins so a missing or
// unreadable file fails in seconds rather than hours into a large run.
// A path of "-" reads from stdin. The returned closer closes every real file,
// leaving stdin alone; it also runs if a later open fails.
func openMergeInputs(paths []string) (inputs []mergeInput, closeAll func(), err error) {
	var files []*os.File
	closeAll = func() {
		for _, f := range files {
			_ = f.Close()
		}
	}
	for _, path := range paths {
		if path == "-" {
			inputs = append(inputs, mergeInput{name: "stdin", reader: os.Stdin})
			continue
		}
		f, oerr := os.Open(path)
		if oerr != nil {
			closeAll()
			return nil, nil, oerr
		}
		files = append(files, f)
		inputs = append(inputs, mergeInput{name: path, reader: f})
	}
	return inputs, closeAll, nil
}

// mergeFile streams one CSV reader into the database, inserting each password not
// already present and folding its tallies into the shared counters.
// Malformed rows are skipped with a warning rather than aborting.
func mergeFile(ctx, dbCtx context.Context, db *sql.DB, counters *mergeCounters, input io.Reader) error {
	reader := csv.NewReader(input)
	reader.FieldsPerRecord = -1 // check the field count ourselves for a clearer message
	reader.ReuseRecord = true

	tx, queries, err := beginMergeBatch(dbCtx, db)
	if err != nil {
		return err
	}
	// Roll back an unfinished batch on any early return; the happy path clears tx.
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	rowNum, batchRows := 0, 0
	for {
		record, rerr := reader.Read()
		if errors.Is(rerr, io.EOF) {
			break
		}
		rowNum++
		if ctx.Err() != nil {
			break
		}
		if rerr != nil {
			slog.Warn("skipping malformed CSV row", "row", rowNum, "error", rerr)
			counters.recordMalformed()
			continue
		}

		password, count, ok := parseMergeRow(record)
		if !ok {
			slog.Warn("skipping malformed CSV row", "row", rowNum, "record", record)
			counters.recordMalformed()
			continue
		}

		changed, ierr := queries.InsertPassword(dbCtx, sqlite.InsertPasswordParams{Password: password, Count: count})
		if ierr != nil {
			return ierr
		}
		if changed > 0 {
			counters.recordMerged(password)
		} else {
			counters.recordSkipped(password)
		}

		if batchRows++; batchRows >= mergeBatchRows {
			if tx, queries, err = commitMergeBatch(dbCtx, db, tx); err != nil {
				return err
			}
			batchRows = 0
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

// parseMergeRow validates one CSV record and returns its password and count.
// It reports ok=false for a wrong field count, a password that is not valid
// UTF-8, or an unparseable, negative count.
func parseMergeRow(record []string) (password string, count int64, ok bool) {
	if len(record) != 2 {
		return "", 0, false
	}
	if !utf8.ValidString(record[0]) {
		return "", 0, false
	}
	count, err := strconv.ParseInt(record[1], 10, 64)
	if err != nil || count < 0 {
		return "", 0, false
	}
	return record[0], count, true
}

// beginMergeBatch opens a transaction and its bound queries for one batch.
func beginMergeBatch(ctx context.Context, db *sql.DB) (*sql.Tx, *sqlite.Queries, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	return tx, sqlite.New(tx), nil
}

// commitMergeBatch commits the current batch, checkpoints the WAL so it stays
// bounded, and opens the next batch.
func commitMergeBatch(ctx context.Context, db *sql.DB, tx *sql.Tx) (*sql.Tx, *sqlite.Queries, error) {
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	if _, err := db.ExecContext(ctx, "PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
		return nil, nil, err
	}
	return beginMergeBatch(ctx, db)
}
