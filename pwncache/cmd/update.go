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

	"github.com/spf13/cobra"

	"pwncache/database"
	"pwncache/pwned"
)

var (
	concurrency int
	limit       int
	progress    time.Duration
	profile     bool
	retries     int
)

// newUpdateCmd builds the "update" sub-command, which downloads or refreshes
// the local mirror of the password hash database.
func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Download or refresh the local password hash database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			defer logs.logFile.Close()
			return runUpdate(cmd.Context())
		},
	}

	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 64, "number of parallel fetch workers")
	cmd.Flags().IntVar(&limit, "limit", 0, "stop after this many prefixes (0 = no limit)")
	cmd.Flags().DurationVarP(&progress, "progress", "p", 10*time.Second, "interval between progress reports")
	cmd.Flags().BoolVar(&profile, "profile", false, "write CPU (cpu.pprof) and heap (heap.pprof) profiles")
	cmd.Flags().IntVar(&retries, "retries", 10, "retry attempts per failed fetch (0 disables)")
	return cmd
}

// runUpdate downloads hash lists until finished, limited, or interrupted.
func runUpdate(ctx context.Context) error {
	logs.file.Info(
		"starting run",
		"concurrency", concurrency,
		"database", databasePath,
		"limit", limit,
		"progress", progress.String(),
		"profile", profile,
		"retries", retries,
		"verbose", verbose,
		"quiet", quiet,
	)

	queries, db, err := database.Open(ctx, databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	downloader := pwned.NewDownloader(db, queries)
	downloader.Concurrency = concurrency
	downloader.Limit = limit
	downloader.Progress = progress
	downloader.MaxRetries = retries
	downloader.ConsoleLog = logs.console
	downloader.FileLog = logs.file

	if profile {
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
				signal.Stop(sig) // a second Ctrl-C now hard-kills the process
				captureHeap()    // workers still hold their buffers, so RAM is real
				cancel()         // begin the graceful shutdown
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
