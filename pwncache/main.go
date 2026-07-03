// Command pwncache downloads the Have I Been Pwned password database to SQLite.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/pflag"

	"pwncache/database"
	"pwncache/pwned"
)

var (
	concurrency  = pflag.IntP("concurrency", "c", 32, "number of parallel fetch workers")
	databasePath = pflag.StringP("database", "d", "pwned.db", "path to the SQLite database")
	limit        = pflag.Int("limit", 0, "stop after this many prefixes (0 = no limit)")
	progress     = pflag.Duration("progress", 10*time.Second, "interval between progress reports")
	retries      = pflag.Int("retries", 10, "retry attempts per failed fetch (0 disables)")
	verbose      = pflag.BoolP("verbose", "v", false, "debug-level logging")
	quiet        = pflag.BoolP("quiet", "q", false, "warnings and errors only")
)

func main() {
	pflag.Parse()

	logs, err := setupLogging(*verbose, *quiet)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logging setup:", err)
		os.Exit(1)
	}
	defer logs.logFile.Close()

	// Cancel the run cleanly on Ctrl-C; state is safe in the database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, logs.console, logs.file); err != nil {
		slog.Error("download failed", "error", err)
		os.Exit(1)
	}
}

// run downloads hash lists until finished, limited, or interrupted.
func run(ctx context.Context, console, file *slog.Logger) error {
	queries, db, err := database.Open(ctx, *databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	downloader := pwned.NewDownloader(db, queries)
	downloader.Concurrency = *concurrency
	downloader.Limit = *limit
	downloader.Progress = *progress
	downloader.MaxRetries = *retries
	downloader.ConsoleLog = console
	downloader.FileLog = file

	err = downloader.Run(ctx)
	if errors.Is(err, context.Canceled) {
		slog.Warn("interrupted, progress saved")
		return nil
	}
	return err
}
