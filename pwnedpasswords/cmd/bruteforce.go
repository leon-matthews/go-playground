package main

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/database/sqlite"
	"pwnedpasswords/filter"
)

// Tuning for the parallel search.
const (
	chunkSize       = 1 << 20   // candidates handed to a worker at a time
	ctxCheckMask    = 1<<14 - 1 // check for cancellation this often within a chunk
	progressEvery   = 60 * time.Second
	bruteforceUsage = "Generate candidate passwords in order and record breach matches"
)

// Building blocks for the cumulative candidate alphabets.
const (
	charsLower   = "abcdefghijklmnopqrstuvwxyz"
	charsDigits  = "0123456789"
	charsUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsSymbols = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

// bruteforceOptions carries the resolved command-line settings.
type bruteforceOptions struct {
	dbPath     string
	cachePath  string
	filterPath string
	alphabet   []byte
	resume     string
	workers    int
}

// newBruteforceCmd builds the "bruteforce" sub-command.
func newBruteforceCmd() *cobra.Command {
	var level int
	var resume string
	var filterPath string
	var workers int
	cmd := &cobra.Command{
		Use:   "bruteforce",
		Short: bruteforceUsage,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := newLogger(verbose, quiet)
			alphabet, err := alphabetForLevel(level)
			if err != nil {
				return err
			}
			return runBruteforce(cmd.Context(), logger, bruteforceOptions{
				dbPath:     databasePath,
				cachePath:  pwnedcachePath,
				filterPath: filterPath,
				alphabet:   alphabet,
				resume:     resume,
				workers:    workers,
			})
		},
	}
	cmd.Flags().IntVarP(&level, "alphabet", "a", 4,
		"character set: 1=lowercase, 2=+space+digits, 3=+uppercase, 4=+symbols")
	cmd.Flags().StringVar(&resume, "resume", "",
		"resume from this pattern (as logged when interrupted)")
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter",
		"membership filter used to skip database lookups")
	cmd.Flags().IntVarP(&workers, "workers", "w", 0,
		"number of parallel workers (default: number of CPUs)")
	return cmd
}

// alphabetForLevel returns the cumulative candidate character set for a level.
func alphabetForLevel(level int) ([]byte, error) {
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

// runBruteforce opens the databases, loads the filter if present, and runs the
// parallel search; without a filter it falls back to a slow serial scan.
func runBruteforce(ctx context.Context, logger *charmlog.Logger, opts bruteforceOptions) error {
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

	length, indices, err := resumeStart(opts.alphabet, opts.resume)
	if err != nil {
		return err
	}

	found, err := filter.Open(opts.filterPath, opts.cachePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		logger.Warn("no filter found; every candidate will hit the database. Build one with 'buildfilter'.",
			"path", opts.filterPath)
		return bruteforceSerial(ctx, logger, writeQueries, cacheQueries, opts.alphabet, length, indices)
	case errors.Is(err, filter.ErrStale):
		return fmt.Errorf("filter %q is stale; rebuild it with 'buildfilter'", opts.filterPath)
	case err != nil:
		return err
	}
	defer found.Close()

	workers := opts.workers
	if workers < 1 {
		workers = runtime.NumCPU()
	}
	logger.Info("loaded filter", "elements", found.Elements, "blocks", found.NumBlocks, "workers", workers)
	return bruteforceParallel(ctx, logger, workerContext{
		write:    writeQueries,
		cache:    cacheQueries,
		filter:   found,
		alphabet: opts.alphabet,
		workers:  workers,
	}, length, indices)
}

// workerContext holds the read-only state every worker shares.
type workerContext struct {
	write    *sqlite.Queries
	cache    *sqlite.Queries
	filter   *filter.Filter
	alphabet []byte
	workers  int
}

// bruteforceParallel enumerates candidates length by length, sharding each
// length across workers and consulting the filter before any database lookup.
func bruteforceParallel(ctx context.Context, logger *charmlog.Logger, wc workerContext, startLength int, startCur []int) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var tried, matched, currentLength atomic.Int64
	var firstErr error
	var once sync.Once

	stopTicker := make(chan struct{})
	tickerDone := make(chan struct{})
	go func() {
		defer close(tickerDone)
		ticker := time.NewTicker(progressEvery)
		defer ticker.Stop()
		for {
			select {
			case <-stopTicker:
				return
			case <-ticker.C:
				logger.Info("progress", "tried", tried.Load(), "matched", matched.Load(), "length", currentLength.Load())
			}
		}
	}()
	stop := func() {
		close(stopTicker)
		<-tickerDone
	}

	co := &coord{base: len(wc.alphabet), chunk: chunkSize}
	length := startLength
	cur := startCur
	for {
		currentLength.Store(int64(length))
		co.reset(cur)

		var wg sync.WaitGroup
		for range wc.workers {
			wg.Go(func() {
				workerTried, workerMatched, err := bruteWorker(runCtx, co, wc, length)
				tried.Add(workerTried)
				matched.Add(workerMatched)
				if err != nil {
					once.Do(func() { firstErr = err })
					cancel()
				}
			})
		}
		wg.Wait()

		if firstErr != nil {
			stop()
			return firstErr
		}
		if ctx.Err() != nil {
			resume := pattern(co.frontier(wc.workers), wc.alphabet)
			stop()
			logger.Info("interrupted", "tried", tried.Load(), "matched", matched.Load(), "resume", resume)
			return nil
		}

		length++
		cur = make([]int, length)
		logger.Info("length", "n", length)
	}
}

// bruteWorker pulls chunks from the coordinator and processes each candidate:
// generate, hash, filter, and only on a filter hit touch the database.
func bruteWorker(ctx context.Context, co *coord, wc workerContext, length int) (tried, matched int64, err error) {
	base := len(wc.alphabet)
	indices := make([]int, length)
	buf := make([]byte, length)
	for {
		if ctx.Err() != nil {
			return tried, matched, nil
		}
		start, ok := co.next()
		if !ok {
			return tried, matched, nil
		}
		copy(indices, start)

		for step := 0; step < co.chunk; step++ {
			if step&ctxCheckMask == 0 && ctx.Err() != nil {
				return tried, matched, nil
			}
			for i, index := range indices {
				buf[i] = wc.alphabet[index]
			}
			sum := sha1.Sum(buf)
			if wc.filter.Contains(sum[:]) {
				hit, err := recordCandidate(ctx, wc.write, wc.cache, string(buf))
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return tried, matched, nil
					}
					return tried, matched, err
				}
				if hit {
					matched++
				}
			}
			tried++
			if !advance(indices, base) {
				break // reached the end of this length
			}
		}
	}
}

// bruteforceSerial is the fallback used when no filter is available: it checks
// every candidate against the database directly.
func bruteforceSerial(ctx context.Context, logger *charmlog.Logger, write, cache *sqlite.Queries, alphabet []byte, length int, indices []int) error {
	base := len(alphabet)
	var tried, matched int64
	buf := make([]byte, length)
	for {
		if ctx.Err() != nil {
			logger.Info("interrupted", "tried", tried, "matched", matched, "resume", pattern(indices, alphabet))
			return nil
		}
		if len(buf) != length {
			buf = make([]byte, length)
		}
		for i, index := range indices {
			buf[i] = alphabet[index]
		}

		hit, err := recordCandidate(ctx, write, cache, string(buf))
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("interrupted", "tried", tried, "matched", matched, "resume", pattern(indices, alphabet))
				return nil
			}
			return err
		}
		tried++
		if hit {
			matched++
		}
		if !advance(indices, base) {
			length++
			indices = make([]int, length)
			logger.Info("length", "n", length)
		}
	}
}

// coord hands out contiguous chunks of one length's candidate space to workers.
type coord struct {
	mu    sync.Mutex
	cur   []int // odometer position of the next chunk to hand out
	done  bool
	base  int
	chunk int
}

// reset positions the coordinator at the start of a new length.
func (c *coord) reset(start []int) {
	c.mu.Lock()
	c.cur = append([]int(nil), start...)
	c.done = false
	c.mu.Unlock()
}

// next returns the start of the next chunk, or ok=false once the length is done.
func (c *coord) next() (start []int, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done {
		return nil, false
	}
	start = append([]int(nil), c.cur...)
	if !addN(c.cur, c.base, c.chunk) {
		c.done = true
	}
	return start, true
}

// frontier returns a conservatively safe resume point: every candidate before
// it is guaranteed processed. At most workers chunks can still be in flight, so
// rewinding the cursor by that much never skips unprocessed candidates.
func (c *coord) frontier(workers int) []int {
	c.mu.Lock()
	defer c.mu.Unlock()
	fr := append([]int(nil), c.cur...)
	subN(fr, c.base, workers*c.chunk)
	return fr
}

// resumeStart returns the length and odometer indices to begin at. Enumeration
// starts exactly at the resume pattern, or at the first one-character candidate.
func resumeStart(alphabet []byte, resume string) (length int, indices []int, err error) {
	if resume == "" {
		return 1, make([]int, 1), nil
	}
	position := make(map[byte]int, len(alphabet))
	for i, b := range alphabet {
		position[b] = i
	}
	indices = make([]int, len(resume))
	for i := 0; i < len(resume); i++ {
		index, ok := position[resume[i]]
		if !ok {
			return 0, nil, fmt.Errorf("resume byte %q is not in the selected alphabet", resume[i])
		}
		indices[i] = index
	}
	return len(resume), indices, nil
}

// pattern renders odometer indices as their candidate string.
func pattern(indices []int, alphabet []byte) string {
	b := make([]byte, len(indices))
	for i, index := range indices {
		b[i] = alphabet[index]
	}
	return string(b)
}

// advance increments the odometer by one, least-significant digit last. It
// returns false on a complete roll-over.
func advance(indices []int, base int) bool {
	for pos := len(indices) - 1; pos >= 0; pos-- {
		indices[pos]++
		if indices[pos] < base {
			return true
		}
		indices[pos] = 0
	}
	return false
}

// addN adds n to the odometer. It returns false if the value rolls past the last
// candidate of this length, in which case indices is left undefined.
func addN(indices []int, base, n int) bool {
	carry := n
	for pos := len(indices) - 1; pos >= 0 && carry > 0; pos-- {
		total := indices[pos] + carry
		indices[pos] = total % base
		carry = total / base
	}
	return carry == 0
}

// subN subtracts n from the odometer, clamping at all-zeros on underflow.
func subN(indices []int, base, n int) {
	for pos := len(indices) - 1; n > 0 && pos >= 0; pos-- {
		digit := indices[pos] - n%base
		n /= base
		if digit < 0 {
			digit += base
			n++ // borrow
		}
		indices[pos] = digit
	}
	if n > 0 {
		for i := range indices {
			indices[i] = 0
		}
	}
}
