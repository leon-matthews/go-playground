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

	"github.com/dustin/go-humanize"
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
func importFile(ctx context.Context, chk *checker, prog *progress, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var t tally
	defer prog.add(&t)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, maxLineBytes), maxLineBytes)
	lineNum := 0
	sinceFlush := 0
	for scanner.Scan() {
		lineNum++
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
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			detail := ""
			if snippet, serr := lineSnippet(path, lineNum+1, snippetBytes); serr == nil {
				detail = fmt.Sprintf(" (starts with %q)", snippet)
			}
			return fmt.Errorf("line %s exceeds the %s characters; please check for encoding errors%s",
				humanize.Comma(int64(lineNum+1)), humanize.Comma(maxLineBytes), detail)
		}
		return err
	}
	return nil
}

// lineSnippet reopens path and returns up to maxBytes from the start of the given
// 1-indexed line, for quoting after the scanner has discarded an over-long token.
// It re-reads from the top, but runs only on the error path just before aborting.
func lineSnippet(path string, line, maxBytes int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for range line - 1 {
		if _, err := reader.ReadBytes('\n'); err != nil {
			return nil, err
		}
	}
	buf := make([]byte, maxBytes)
	n, err := io.ReadFull(reader, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return buf[:n], nil
}
