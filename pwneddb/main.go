// Command pwneddb downloads the Have I Been Pwned password database to SQLite.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"time"

	charminglog "github.com/charmbracelet/log"
	"github.com/spf13/pflag"

	"pwneddb/database"
	"pwneddb/pwned"
)

var (
	databasePath = pflag.StringP("database", "d", "pwned.db", "path to the SQLite database")
	limit        = pflag.Int("limit", 0, "stop after this many prefixes (0 = no limit)")
	progress     = pflag.Duration("progress", 30*time.Second, "interval between progress reports")
	verbose      = pflag.BoolP("verbose", "v", false, "debug-level logging")
	quiet        = pflag.BoolP("quiet", "q", false, "warnings and errors only")
)

// setupLogging installs a colourised charmbracelet logger as the slog default.
func setupLogging() {
	// Level is chosen from the -v and -q flags; -v wins if both are set.
	level := charminglog.InfoLevel
	switch {
	case *verbose:
		level = charminglog.DebugLevel
	case *quiet:
		level = charminglog.WarnLevel
	}

	handler := charminglog.NewWithOptions(os.Stderr, charminglog.Options{
		Level:           level,
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	pflag.Parse()
	setupLogging()

	// Cancel the run cleanly on Ctrl-C; state is safe in the database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("download failed", "error", err)
		os.Exit(1)
	}
}

// run downloads hash lists until finished, limited, or interrupted.
func run(ctx context.Context) error {
	queries, db, err := database.Open(ctx, *databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	downloader := pwned.NewDownloader(db, queries)
	downloader.Limit = *limit
	downloader.Progress = *progress

	err = downloader.Run(ctx)
	if errors.Is(err, context.Canceled) {
		slog.Warn("interrupted, progress saved")
		return nil
	}
	return err
}
