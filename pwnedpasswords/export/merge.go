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

// Merge streams a CSV of "password,count" rows into the output database,
// inserting each password that is not already present and leaving existing
// counts untouched. Malformed rows are skipped with a warning rather than
// aborting the run, so one bad line cannot spoil a long restore.
func Merge(ctx context.Context, logs logging.Logging, dbPath, csvPath string, interval time.Duration) (err error) {
	_, db, err := database.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// A path of "-" reads the CSV from stdin, which we must not close.
	input := os.Stdin
	source := "stdin"
	if csvPath != "-" {
		f, err := os.Open(csvPath)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
		source = csvPath
	}
	slog.Info("merging CSV into database", "csv", source, "database", dbPath)

	reader := csv.NewReader(input)
	reader.FieldsPerRecord = -1 // check the field count ourselves for a clearer message
	reader.ReuseRecord = true

	var counters mergeCounters
	rep := progress.StartReporter(interval, counters.reportTo(logs.Console, logs.File))
	defer rep.StopAndReport()

	// Writes use a cancellation-free context so an interrupt commits the rows read
	// so far cleanly instead of tearing down the open transaction; the loop below
	// honours ctx itself, and merge is additive so a resumed run skips them.
	dbCtx := context.WithoutCancel(ctx)

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
	_, err = db.ExecContext(dbCtx, "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// parseMergeRow validates one CSV record and returns its password and count.
// It reports ok=false for a wrong field count or an unparseable, negative count.
func parseMergeRow(record []string) (password string, count int64, ok bool) {
	if len(record) != 2 {
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
