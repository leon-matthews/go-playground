package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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
func runImport(ctx context.Context, logs logging, opts importOptions) (err error) {
	_, writeDB, err := database.Open(ctx, opts.dbPath)
	if err != nil {
		return err
	}
	defer writeDB.Close()

	cacheQueries, cacheDB, err := database.OpenHashes(ctx, opts.cachePath)
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
		slog.Info("using filter", "path", opts.filterPath, "elements", found.NumEntries, "blocks", found.NumBlocks)
	}

	writer := database.NewBatchWriter(writeDB)
	defer func() {
		if cerr := writer.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	chk := &checker{writer: writer, cache: cacheQueries, filter: found}
	prog := &progress{}
	rep := startReporter(opts.progress, prog.reportTo(logs.console, logs.file, found != nil))
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

// Cap each line at 1 KiB; a word list holds one short word per line, so anything larger is malformed.
const maxLineBytes = 1024

// Quote this many leading bytes of an over-long line to help identify the malformed input.
const snippetBytes = 64

// importFile streams one word list, running every non-empty line through the
// checker and folding its counts into the shared progress.
// Over-long lines are skipped with a warning rather than aborting the import.
func importFile(ctx context.Context, chk *checker, prog *progress, path string) error {
	// A path of "-" reads the word list from stdin, which we must not close.
	file := os.Stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		file = f
	}

	var t tally
	defer prog.add(&t)

	// The buffer size doubles as the over-long threshold: a line that fills it
	// without a newline comes back with isPrefix set.
	reader := bufio.NewReaderSize(file, maxLineBytes)
	lineNum := 0
	sinceFlush := 0
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		lineNum++
		if ctx.Err() != nil {
			return nil
		}

		if isPrefix {
			// line holds only the first chunk; snapshot a snippet, then drain the
			// rest without buffering it. string() copies, so it survives drainLine.
			snippet := string(line[:min(len(line), snippetBytes)])
			if err := drainLine(reader); err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			slog.Warn("skipping over-long line",
				"line", lineNum, "limit", maxLineBytes, "starts_with", snippet)
			continue
		}

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
	return nil
}

// drainLine consumes the remainder of an over-long line, up to and including its
// newline, without retaining any of it.
// Call only after ReadLine has reported isPrefix.
func drainLine(reader *bufio.Reader) error {
	for {
		_, isPrefix, err := reader.ReadLine()
		if err != nil {
			return err
		}
		if !isPrefix {
			return nil
		}
	}
}
