// Command pwncache downloads the Have I Been Pwned password database to SQLite.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/spf13/pflag"

	"pwncache/database"
	"pwncache/pwned"
)

var (
	concurrency  = pflag.IntP("concurrency", "c", 64, "number of parallel fetch workers")
	databasePath = pflag.StringP("database", "d", "pwned.db", "path to the SQLite database")
	limit        = pflag.Int("limit", 0, "stop after this many prefixes (0 = no limit)")
	progress     = pflag.DurationP("progress", "p", 10*time.Second, "interval between progress reports")
	profile      = pflag.Bool("profile", false, "Write a CPU profile to default.pgo for PGO builds")
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

	if *profile {
		stopProfile, err := startProfile("default.pgo")
		if err != nil {
			return fmt.Errorf("creating profile: %w", err)
		}
		defer stopProfile()
	}

	err = downloader.Run(ctx)
	if errors.Is(err, context.Canceled) {
		slog.Warn("interrupted, progress saved")
		return nil
	}
	return err
}

// startProfile begins writing a CPU profile to the given path.
//
// Returns the stop function that ends profiling, closes the file, and reports
// where the profile went. A plain create suffices here, with none of the care
// writeResults takes, as a spoiled profile is simply overwritten next run.
func startProfile(path string) (func(), error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return nil, err
	}
	return func() {
		pprof.StopCPUProfile()
		file.Close()
		fmt.Fprintf(os.Stderr, "CPU profile written to %s\n", path)
	}, nil
}
