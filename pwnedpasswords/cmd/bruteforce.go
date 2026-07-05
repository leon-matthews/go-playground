package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/bits"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/filter"
)

// Tuning for the parallel search.
const (
	// maxChunk caps the candidates handed to a worker at once. Shorter lengths
	// use a smaller chunk (see chunkForSpace) so every worker still gets work.
	maxChunk = 1 << 20
	// minChunk floors the chunk size, keeping coordinator lock overhead negligible.
	minChunk = 1 << 10
	// chunksPerWorker is the number of chunks aimed at each worker, so uneven
	// per-candidate cost still balances out across them.
	chunksPerWorker = 8

	ctxCheckMask    = 1<<14 - 1 // check for cancellation this often within a chunk
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
	progress   time.Duration
}

// newBruteforceCmd builds the "bruteforce" sub-command.
func newBruteforceCmd() *cobra.Command {
	var level int
	var resume string
	var filterPath string
	var workers int
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "bruteforce",
		Short: bruteforceUsage,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logs, err := setupLogging(verbose)
			if err != nil {
				return err
			}
			defer logs.logFile.Close()
			alphabet, err := alphabetForLevel(level)
			if err != nil {
				return err
			}
			return runBruteforce(cmd.Context(), logs, bruteforceOptions{
				dbPath:     databasePath,
				cachePath:  pwnedcachePath,
				filterPath: filterPath,
				alphabet:   alphabet,
				resume:     resume,
				workers:    workers,
				progress:   progressInterval,
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
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
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
func runBruteforce(ctx context.Context, logs logging, opts bruteforceOptions) (err error) {
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

	length, indices, err := resumeStart(opts.alphabet, opts.resume)
	if err != nil {
		return err
	}

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

	if found == nil {
		return bruteforceSerial(ctx, chk, prog, opts.alphabet, length, indices)
	}

	workers := opts.workers
	if workers < 1 {
		workers = runtime.NumCPU()
	}
	slog.Info("parallel search", "workers", workers)
	return bruteforceParallel(ctx, chk, prog, opts.alphabet, workers, length, indices)
}

// bruteforceParallel enumerates candidates length by length, sharding each
// length across workers and consulting the filter before any database lookup.
func bruteforceParallel(ctx context.Context, chk *checker, prog *progress, alphabet []byte, workers, startLength int, startCur []int) error {
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
				if err := bruteWorker(runCtx, chk, co, prog, alphabet, length); err != nil {
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

// bruteWorker pulls chunks from the coordinator and runs each candidate through
// the checker, folding its counts into the shared progress every flushEvery
// candidates so a long chunk still reports steady progress.
func bruteWorker(ctx context.Context, chk *checker, co *coord, prog *progress, alphabet []byte, length int) error {
	base := len(alphabet)
	indices := make([]int, length)
	buf := make([]byte, length)
	var t tally
	defer prog.add(&t)
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
			if err := chk.check(ctx, &t, buf); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
			if sinceFlush++; sinceFlush >= flushEvery {
				prog.add(&t)
				sinceFlush = 0
			}
			if !advance(indices, base) {
				break // reached the end of this length
			}
		}
		prog.add(&t)
	}
}

// bruteforceSerial is the fallback used when no filter is available: it runs
// every candidate through the checker, which then hits the database directly.
func bruteforceSerial(ctx context.Context, chk *checker, prog *progress, alphabet []byte, length int, indices []int) error {
	base := len(alphabet)
	buf := make([]byte, length)
	var t tally
	defer prog.add(&t)
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

		if err := chk.check(ctx, &t, buf); err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("interrupted", "resume", pattern(indices, alphabet))
				return nil
			}
			return err
		}
		if sinceFlush++; sinceFlush >= flushEvery {
			prog.add(&t)
			sinceFlush = 0
		}
		if !advance(indices, base) {
			length++
			indices = make([]int, length)
			slog.Info("length", "n", length)
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

// chunkForSpace picks the chunk size for a candidate space of the given size so
// that each worker gets about chunksPerWorker chunks, clamped to [minChunk,
// maxChunk]. A small space therefore still splits across workers instead of
// landing on one, while a huge space keeps chunks bounded.
func chunkForSpace(space uint64, workers int) int {
	target := space / (uint64(workers) * chunksPerWorker)
	switch {
	case target < minChunk:
		return minChunk
	case target > maxChunk:
		return maxChunk
	default:
		return int(target)
	}
}

// powSat returns base**exp, saturating at the maximum uint64 rather than
// wrapping, so callers can compare an astronomically large candidate space
// against ordinary bounds without overflow.
func powSat(base, exp uint64) uint64 {
	result := uint64(1)
	for range exp {
		hi, lo := bits.Mul64(result, base)
		if hi != 0 {
			return ^uint64(0)
		}
		result = lo
	}
	return result
}
