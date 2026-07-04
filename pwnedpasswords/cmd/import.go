package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/filter"
)

// importOptions carries the resolved command-line settings.
type importOptions struct {
	dbPath     string
	cachePath  string
	filterPath string
	progress   time.Duration
	files      []string
}

// newImportCmd builds the "import" sub-command.
func newImportCmd() *cobra.Command {
	var filterPath string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "import <wordlist>...",
		Short: "Import word lists, recording passwords found in the breach corpus",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logs, err := setupLogging(verbose)
			if err != nil {
				return err
			}
			defer logs.logFile.Close()
			return runImport(cmd.Context(), logs, importOptions{
				dbPath:     databasePath,
				cachePath:  pwnedcachePath,
				filterPath: filterPath,
				progress:   progressInterval,
				files:      args,
			})
		},
	}
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter",
		"membership filter used to skip database lookups")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
	return cmd
}

// runImport opens the databases, loads the filter if present, and records every
// word-list password found in the breach corpus along with its breach count.
func runImport(ctx context.Context, logs logging, opts importOptions) error {
	writeQueries, writeDB, err := database.Open(ctx, opts.dbPath)
	if err != nil {
		return err
	}
	defer writeDB.Close()

	cacheQueries, cacheDB, err := database.OpenCache(ctx, opts.cachePath)
	if err != nil {
		return err
	}
	defer cacheDB.Close()
	slog.Info("using databases", "output", opts.dbPath, "pwnedcache", opts.cachePath)

	found, err := filter.Open(opts.filterPath, opts.cachePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		slog.Warn("no filter found; every candidate will hit the hashes table - build one with 'buildfilter'",
			"path", opts.filterPath)
		found = nil
	case errors.Is(err, filter.ErrStale):
		return fmt.Errorf("filter %q is stale; rebuild it with 'buildfilter'", opts.filterPath)
	case err != nil:
		return err
	default:
		defer found.Close()
		slog.Info("using filter", "path", opts.filterPath, "elements", found.Elements, "blocks", found.NumBlocks)
	}

	chk := &checker{write: writeQueries, cache: cacheQueries, filter: found}
	prog := &progress{}
	rep := startReporter(prog, logs.console, logs.file, opts.progress, found != nil)
	defer rep.stopAndReport()

	for _, file := range opts.files {
		if ctx.Err() != nil {
			break
		}
		before := prog.snapshot()
		if err := importFile(ctx, chk, prog, file); err != nil {
			return fmt.Errorf("importing %q: %w", file, err)
		}
		after := prog.snapshot()
		slog.Info("imported word list", "file", file,
			"filter_queries", after.filterQueries-before.filterQueries,
			"hash_queries", after.hashQueries-before.hashQueries,
			"found", after.found-before.found,
			"changed", after.changed-before.changed)
	}
	return nil
}

// importFile streams one word list, running every non-empty line through the
// checker and folding its counts into the shared progress.
func importFile(ctx context.Context, chk *checker, prog *progress, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var t tally
	defer prog.add(&t)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	sinceFlush := 0
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := chk.check(ctx, &t, line); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		if sinceFlush++; sinceFlush >= flushEvery {
			prog.add(&t)
			sinceFlush = 0
		}
	}
	return scanner.Err()
}
