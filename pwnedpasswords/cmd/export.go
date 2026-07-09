package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/database/sqlite"
)

// newExportCmd builds the "export" sub-command.
func newExportCmd() *cobra.Command {
	var top int
	var format string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write the most-breached passwords as a denylist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExport(cmd.Context(), cmd.OutOrStdout(), newConsoleLogger(verbose),
				databasePath, top, format, progressInterval)
		},
	}
	cmd.Flags().IntVarP(&top, "top", "n", 1000, "number of passwords to write (ignored by csv)")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "output format: text, json, or csv (csv is a full dump for 'merge')")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", defaultProgressInterval,
		"interval between progress reports (csv only)")
	return cmd
}

// denylistEntry is one row of JSON denylist output.
type denylistEntry struct {
	Password string `json:"password"`
	Count    int64  `json:"count"`
}

// runExport writes passwords for the chosen format: text and json emit the top-N
// denylist by breach count, while csv streams the whole table for 'merge'.
func runExport(ctx context.Context, out io.Writer, console *slog.Logger, dbPath string, top int, format string, interval time.Duration) error {
	queries, db, err := database.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	switch format {
	case "csv":
		return exportCSV(ctx, out, console, db, interval)
	case "text", "json":
		rows, err := queries.TopPasswords(ctx, int64(top))
		if err != nil {
			return err
		}
		if format == "text" {
			return exportText(out, rows)
		}
		return exportJSON(out, rows)
	default:
		return fmt.Errorf("unknown format %q: use text, json, or csv", format)
	}
}

// exportText writes one password per line.
func exportText(out io.Writer, rows []sqlite.Password) error {
	for _, row := range rows {
		if _, err := fmt.Fprintln(out, row.Password); err != nil {
			return err
		}
	}
	return nil
}

// exportJSON writes the rows as an indented array of password/count objects.
func exportJSON(out io.Writer, rows []sqlite.Password) error {
	entries := make([]denylistEntry, len(rows))
	for i, row := range rows {
		entries[i] = denylistEntry{Password: row.Password, Count: row.Count}
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

// exportCSV streams the entire passwords table as "password,count" records
// without materializing it, so it scales to the whole table. The csv.Writer
// quotes any password containing a comma, quote, or newline, so 'merge' reads it
// back exactly. The dump is unordered: merge ignores order, so skipping the sort
// keeps it a plain table scan. Progress goes to the console only, leaving CSV the
// sole occupant of out.
func exportCSV(ctx context.Context, out io.Writer, console *slog.Logger, db *sql.DB, interval time.Duration) error {
	rows, err := db.QueryContext(ctx, "SELECT password, count FROM passwords")
	if err != nil {
		return err
	}
	defer rows.Close()

	var counters exportCounters
	rep := startReporter(interval, counters.reportTo(console))
	defer rep.stopAndReport()

	writer := csv.NewWriter(out)
	var password string
	var count int64
	for rows.Next() {
		if err := rows.Scan(&password, &count); err != nil {
			return err
		}
		if err := writer.Write([]string{password, strconv.FormatInt(count, 10)}); err != nil {
			return err
		}
		counters.record(password)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

// exportCounters tallies a CSV dump. Its fields are atomic because the reporter
// goroutine reads them while the export loop updates them.
type exportCounters struct {
	written atomic.Int64           // rows written so far
	sample  atomic.Pointer[string] // most recent password written, for display
}

// record counts one written row.
func (e *exportCounters) record(password string) {
	e.written.Add(1)
	e.sample.Store(&password)
}

// reportTo returns a report function that logs the row count as a friendly
// console line. Read-only export writes no structured file record.
func (e *exportCounters) reportTo(console *slog.Logger) func(kind string) {
	return func(kind string) {
		written := e.written.Load()
		sample := ""
		if s := e.sample.Load(); s != nil {
			sample = *s
		}
		console.Info(humanExportProgress(kind, written, sample))
	}
}

// humanExportProgress renders a friendly one-line progress or summary message.
func humanExportProgress(kind string, written int64, sample string) string {
	line := humanize.Comma(written) + " rows written"
	if sample != "" {
		line += " > current: " + sample
	}
	if kind == "summary" {
		return "Finished: " + line
	}
	return line
}
