package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"local.dev/pwnedcache/internal/database"
	"local.dev/pwnedcache/internal/pwned"
)

// updateCmd holds the flags of the "update" command.
type updateCmd struct {
	Concurrency int           `short:"c" default:"100" help:"Number of parallel fetch workers."`
	Limit       int           `default:"0" help:"Stop after this many prefixes (0 = no limit)."`
	Progress    time.Duration `short:"p" default:"10s" help:"Interval between progress reports."`
	Profile     bool          `help:"Write CPU (cpu.pprof) and heap (heap.pprof) profiles."`
	Retries     int           `default:"10" help:"Retry attempts per failed fetch (0 disables)."`
}

// Run downloads hash lists until finished, limited, or interrupted.
func (c *updateCmd) Run(ctx context.Context, global *cli, logs logging) error {
	logs.file.Info(
		"starting run",
		"concurrency", c.Concurrency,
		"database", global.Database,
		"limit", c.Limit,
		"progress", c.Progress.String(),
		"profile", c.Profile,
		"retries", c.Retries,
		"verbose", global.Verbose,
	)

	queries, db, err := database.Open(ctx, global.Database)
	if err != nil {
		return err
	}

	// Leaving WAL mode is housekeeping, so a failure warns rather than fails
	defer func() {
		if err := database.Close(ctx, db); err != nil {
			slog.Warn("closing database", "error", err)
		}
	}()

	downloader := pwned.NewDownloader(db, queries)
	downloader.Concurrency = c.Concurrency
	downloader.Limit = c.Limit
	downloader.Progress = c.Progress
	downloader.MaxRetries = c.Retries
	downloader.ConsoleLog = logs.console
	downloader.FileLog = logs.file

	if c.Profile {
		stopProfile, err := startProfile("cpu.pprof")
		if err != nil {
			return fmt.Errorf("creating profile: %w", err)
		}

		// once ensures exactly one heap snapshot: the earliest trigger wins
		var once sync.Once
		captureHeap := func() {
			once.Do(func() {
				if err := writeHeapProfile("heap.pprof"); err != nil {
					slog.Error("writing heap profile", "error", err)
				}
			})
		}
		// Stop CPU profiling, then snapshot the heap at exit if not already done
		defer func() {
			stopProfile()
			captureHeap()
		}()

		// The download ignores the parent's signal cancellation so the handler
		// below can snapshot the still-live heap before the workers wind down.
		runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
		defer cancel()
		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt)
			select {
			case <-sig:
				signal.Stop(sig)
				captureHeap() // workers still hold their buffers, so RAM is real
				cancel()      // begin the graceful shutdown
			case <-runCtx.Done():
				signal.Stop(sig)
			}
		}()
		ctx = runCtx
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

// writeHeapProfile writes a snapshot of the heap to the given path.
//
// A heap profile is a point-in-time sample, not a recording, so it is taken
// once at exit. A GC first settles the in-use figures; the cumulative
// alloc_space view does not depend on when the snapshot is taken.
func writeHeapProfile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	runtime.GC() // settle live heap so in-use space is accurate
	if err := pprof.WriteHeapProfile(file); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "heap profile written to %s\n", path)
	return nil
}
