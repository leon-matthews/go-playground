// Package wordlist imports candidate passwords from word-list files, recording
// any that match the breach corpus.
package wordlist

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"pwnedpasswords/checker"
	"pwnedpasswords/database"
	"pwnedpasswords/filter"
	"pwnedpasswords/logging"
	"pwnedpasswords/progress"
)

// Options carries the resolved command-line settings.
type Options struct {
	DBPath     string
	CachePath  string
	FilterPath string
	Progress   time.Duration
	Files      []string
}

// Run opens the databases, loads the filter if present, and records every
// word-list password found in the breach corpus along with its breach count.
func Run(ctx context.Context, logs logging.Logging, opts Options) (err error) {
	_, writeDB, err := database.Open(ctx, opts.DBPath)
	if err != nil {
		return err
	}
	defer writeDB.Close()

	cacheQueries, cacheDB, err := database.OpenHashes(ctx, opts.CachePath)
	if err != nil {
		return err
	}
	defer cacheDB.Close()
	slog.Info("using databases", "output", opts.DBPath, "pwnedcache", opts.CachePath)

	found, err := filter.Open(opts.FilterPath, opts.CachePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		slog.Warn("no filter found; every candidate will hit the hashes table - build one with 'buildfilter'",
			"path", opts.FilterPath)
		found = nil
	case errors.Is(err, filter.ErrStale):
		return fmt.Errorf("filter %q is stale; rebuild it with 'buildfilter'", opts.FilterPath)
	case err != nil:
		return err
	default:
		defer found.Close()
		slog.Info("using filter", "path", opts.FilterPath, "elements", found.NumEntries, "blocks", found.NumBlocks)
	}

	writer := database.NewBatchWriter(writeDB)
	defer func() {
		if cerr := writer.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	chk := &checker.Checker{Writer: writer, Cache: cacheQueries, Filter: found}
	prog := &progress.Progress{}
	rep := progress.StartReporter(opts.Progress, prog.ReportTo(logs.Console, logs.File, found != nil))
	defer rep.StopAndReport()

	for _, file := range opts.Files {
		if ctx.Err() != nil {
			break
		}
		before := prog.Snapshot()
		if err := importFile(ctx, chk, prog, file); err != nil {
			return fmt.Errorf("importing %q: %w", file, err)
		}
		after := prog.Snapshot()
		slog.Info("imported word list", "file", file,
			"filter_queries", after.FilterQueries-before.FilterQueries,
			"hash_queries", after.HashQueries-before.HashQueries,
			"found", after.Found-before.Found,
			"changed", after.Changed-before.Changed)
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
func importFile(ctx context.Context, chk *checker.Checker, prog *progress.Progress, path string) error {
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

	var t progress.Tally
	defer prog.Add(&t)

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
		if err := chk.Check(ctx, &t, line); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		if sinceFlush++; sinceFlush >= progress.FlushEvery {
			prog.Add(&t)
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
