// Package bruteforce generates candidate passwords by brute force, in odometer
// order shortest first, and records any that match the breach corpus.
package bruteforce

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"pwnedpasswords/checker"
	"pwnedpasswords/database"
	"pwnedpasswords/filter"
	"pwnedpasswords/logging"
	"pwnedpasswords/progress"
)

// Tuning for the parallel bruteforce.
const (
	// maxChunk caps the candidates handed to a worker at once. Shorter lengths
	// use a smaller chunk (see chunkForSpace) so every worker still gets work.
	maxChunk = 1 << 20
	// minChunk floors the chunk size, keeping coordinator lock overhead negligible.
	minChunk = 1 << 10
	// chunksPerWorker is the number of chunks aimed at each worker, so uneven
	// per-candidate cost still balances out across them.
	chunksPerWorker = 8

	ctxCheckMask = 1<<14 - 1 // check for cancellation this often within a chunk
)

// Building blocks for the cumulative candidate alphabets.
const (
	charsLower   = "abcdefghijklmnopqrstuvwxyz"
	charsDigits  = "0123456789"
	charsUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsSymbols = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

// Options carries the resolved command-line settings.
type Options struct {
	DBPath     string
	CachePath  string
	FilterPath string
	Alphabet   []byte
	Resume     string
	Workers    int
	Progress   time.Duration
}

// AlphabetForLevel returns the cumulative candidate character set for a level.
func AlphabetForLevel(level int) ([]byte, error) {
	switch level {
	case 1:
		return []byte(charsLower), nil
	case 2:
		return []byte(charsLower + " " + charsDigits), nil
	case 3:
		return []byte(charsLower + " " + charsDigits + charsUpper), nil
	case 4:
		return []byte(charsLower + " " + charsDigits + charsUpper + charsSymbols), nil
	default:
		return nil, fmt.Errorf("unknown alphabet %d: choose 1-4", level)
	}
}

// Run opens the databases, loads the filter if present, and runs the parallel
// bruteforce; without a filter it falls back to a slow serial scan.
func Run(ctx context.Context, logs logging.Logging, opts Options) (err error) {
	workers := opts.Workers
	if workers < 1 {
		workers = runtime.NumCPU()
	}

	_, writeDB, err := database.Open(ctx, opts.DBPath)
	if err != nil {
		return err
	}
	defer writeDB.Close()

	cacheQueries, cacheDB, err := database.OpenRO(ctx, opts.CachePath, workers)
	if err != nil {
		return err
	}
	defer cacheDB.Close()
	slog.Info("using databases", "output", opts.DBPath, "pwnedcache", opts.CachePath)

	length, indices, err := resumeStart(opts.Alphabet, opts.Resume)
	if err != nil {
		return err
	}

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

	if found == nil {
		return searchSerial(ctx, chk, prog, opts.Alphabet, length, indices)
	}

	slog.Info("parallel bruteforce", "workers", workers)
	return searchParallel(ctx, chk, prog, opts.Alphabet, workers, length, indices)
}

// searchParallel enumerates candidates length by length, sharding each length
// across workers and consulting the filter before any database lookup.
func searchParallel(ctx context.Context, chk *checker.Checker, prog *progress.Progress, alphabet []byte, workers, startLength int, startCur []int) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var firstErr error
	var once sync.Once

	co := &coord{base: len(alphabet)}
	length := startLength
	cur := startCur
	for {
		co.chunk = chunkForSpace(powSat(uint64(len(alphabet)), uint64(length)), workers)
		co.reset(cur)

		var wg sync.WaitGroup
		for range workers {
			wg.Go(func() {
				if err := searchWorker(runCtx, chk, co, prog, alphabet, length); err != nil {
					once.Do(func() { firstErr = err })
					cancel()
				}
			})
		}
		wg.Wait()

		if firstErr != nil {
			return firstErr
		}
		if ctx.Err() != nil {
			slog.Info("interrupted", "resume", pattern(co.frontier(workers), alphabet))
			return nil
		}

		length++
		cur = make([]int, length)
		slog.Info("length", "n", length)
	}
}

// searchWorker pulls chunks from the coordinator and runs each candidate through
// the checker, folding its counts into the shared progress every FlushEvery
// candidates so a long chunk still reports steady progress.
func searchWorker(ctx context.Context, chk *checker.Checker, co *coord, prog *progress.Progress, alphabet []byte, length int) error {
	base := len(alphabet)
	indices := make([]int, length)
	buf := make([]byte, length)
	var t progress.Tally
	defer prog.Add(&t)
	sinceFlush := 0
	for {
		if ctx.Err() != nil {
			return nil
		}
		start, ok := co.next()
		if !ok {
			return nil
		}
		copy(indices, start)

		for step := 0; step < co.chunk; step++ {
			if step&ctxCheckMask == 0 && ctx.Err() != nil {
				return nil
			}
			for i, index := range indices {
				buf[i] = alphabet[index]
			}
			if err := chk.Check(ctx, &t, buf); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
			if sinceFlush++; sinceFlush >= progress.FlushEvery {
				prog.Add(&t)
				sinceFlush = 0
			}
			if !advance(indices, base) {
				break // reached the end of this length
			}
		}
		prog.Add(&t)
	}
}

// searchSerial is the fallback used when no filter is available: it runs every
// candidate through the checker, which then hits the database directly.
func searchSerial(ctx context.Context, chk *checker.Checker, prog *progress.Progress, alphabet []byte, length int, indices []int) error {
	base := len(alphabet)
	buf := make([]byte, length)
	var t progress.Tally
	defer prog.Add(&t)
	sinceFlush := 0
	for {
		if ctx.Err() != nil {
			slog.Info("interrupted", "resume", pattern(indices, alphabet))
			return nil
		}
		if len(buf) != length {
			buf = make([]byte, length)
		}
		for i, index := range indices {
			buf[i] = alphabet[index]
		}

		if err := chk.Check(ctx, &t, buf); err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("interrupted", "resume", pattern(indices, alphabet))
				return nil
			}
			return err
		}
		if sinceFlush++; sinceFlush >= progress.FlushEvery {
			prog.Add(&t)
			sinceFlush = 0
		}
		if !advance(indices, base) {
			length++
			indices = make([]int, length)
			slog.Info("length", "n", length)
		}
	}
}
