// Silly benchmark which plays many, many solo games of snakes and ladders.
//
// A Go port of the Python original, snakes_and_ladders.py, found in the
// parent directory.
//
// Copyright 2011-2026 Leon Matthews. Released under the Apache 2.0 licence.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/pflag"
	"local.dev/ladders"
)

// options holds the parsed command-line arguments.
type options struct {
	jobs      int
	jsonPaths []string
	numGames  int64
	profile   bool
	progress  time.Duration
	seconds   int
}

// parse builds the parsed options from the given command-line arguments.
//
// Invalid arguments and requests for help exit the program directly, as the
// Python argparse module does.
func parse(args []string) (options, error) {
	flags := pflag.NewFlagSet("go_ladders", pflag.ExitOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: go_ladders [flags] [results.json ... target.json]")
		fmt.Fprintln(os.Stderr, "Benchmark results accumulate into a single named JSON file. Naming several")
		fmt.Fprintln(os.Stderr, "files skips the benchmark, merging them into the last file instead.")
		flags.PrintDefaults()
	}

	// Multicore? Plain '-j' means all cores; normalizeJobs supports make-style counts.
	numCores := runtime.NumCPU()
	jobs := flags.IntP("jobs", "j", 1, fmt.Sprintf("Run on multiple cores (%d found)", numCores))
	flags.Lookup("jobs").NoOptDefVal = strconv.Itoa(numCores)

	// Iterations or seconds?
	numGames := flags.Int64P("games", "n", 0, "Total number of games to play, eg. 100 or 1_000_000")
	seconds := flags.IntP("seconds", "s", 10, "Seconds to play for, 0 to run until interrupted")

	// How often to report progress and update the results file during a long run?
	progress := flags.DurationP("progress", "p", 60*time.Second, "Interval between progress reports, eg. 30s or 2m")

	// Capture a CPU profile? Its fixed name is the one 'go build' auto-detects.
	profile := flags.Bool("profile", false, "Write a CPU profile to default.pgo for PGO builds")

	if err := flags.Parse(normalizeJobs(args)); err != nil {
		return options{}, err
	}
	if flags.Changed("games") && flags.Changed("seconds") {
		return options{}, errors.New("only one of -n and -s may be given")
	}
	if flags.NArg() > 1 && (flags.Changed("jobs") || flags.Changed("games") ||
		flags.Changed("seconds") || flags.Changed("progress") || flags.Changed("profile")) {
		return options{}, errors.New("merging several files plays no games, so -j, -n, -s, -p, and --profile may not be given")
	}
	if *jobs < 1 {
		return options{}, fmt.Errorf("number of jobs must be at least one, given: %d", *jobs)
	}
	// The upper bound keeps the timeout within a time.Duration's count of nanoseconds
	const maxSeconds = math.MaxInt64 / int64(time.Second)
	if *seconds < 0 || int64(*seconds) > maxSeconds {
		return options{}, fmt.Errorf("number of seconds out of range (0 to %d), given: %d", maxSeconds, *seconds)
	}
	if *progress < time.Second {
		return options{}, fmt.Errorf("progress interval must be at least one second, given: %v", *progress)
	}

	if flags.Changed("games") && *numGames < 1 {
		return options{}, fmt.Errorf("number of games must be at least one, given: %d", *numGames)
	}

	// Reject the same file named twice; Abs catches spellings like ./A.json
	seen := make(map[string]string, flags.NArg())
	for _, path := range flags.Args() {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return options{}, err
		}
		if earlier, found := seen[absolute]; found {
			return options{}, fmt.Errorf("file named twice: %s and %s", earlier, path)
		}
		seen[absolute] = path
	}

	return options{
		jobs:      *jobs,
		jsonPaths: flags.Args(),
		numGames:  *numGames,
		profile:   *profile,
		progress:  *progress,
		seconds:   *seconds,
	}, nil
}

// readResults loads and combines results from the given JSON files.
//
// Every file must exist, parse, and pass the consistency check, except the
// last, which is the output target and so is allowed to be missing.
func readResults(paths []string) (ladders.BenchmarkResult, error) {
	var combined ladders.BenchmarkResult
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			if index == len(paths)-1 && errors.Is(err, os.ErrNotExist) {
				continue
			}
			return ladders.BenchmarkResult{}, err
		}
		var result ladders.BenchmarkResult
		if err := json.Unmarshal(data, &result); err != nil {
			return ladders.BenchmarkResult{}, fmt.Errorf("%s: %w", path, err)
		}
		if err := result.Validate(); err != nil {
			return ladders.BenchmarkResult{}, fmt.Errorf("%s: %w", path, err)
		}
		combined = combined.Add(result)
	}
	return combined, nil
}

// lockResults claims exclusive use of the given results file.
//
// A sibling lock file created with O_EXCL makes a second concurrent run fail
// loudly, rather than silently losing one run's games. Returns the unlock
// function that releases the claim.
func lockResults(path string) (func(), error) {
	lock := path + ".lock"
	file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf(
				"%s exists: is another run writing to %s? (delete the lock file if not)", lock, path,
			)
		}
		return nil, err
	}
	file.Close()
	return func() { os.Remove(lock) }, nil
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

// normalizeJobs rewrites make-style job counts into the -j=4 form pflag needs.
//
// Mirrors GNU make, which accepts an attached count like -j4, a separate
// all-digits argument like -j 4, or a bare -j meaning every core.
func normalizeJobs(args []string) []string {
	normalized := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--":
			// Everything after the terminator is positional; pass it through untouched
			return append(normalized, args[index:]...)
		case strings.HasPrefix(arg, "-j") && isDigits(arg[2:]):
			// Attached count, eg. -j4
			normalized = append(normalized, "-j="+arg[2:])
		case (arg == "-j" || arg == "--jobs") && index+1 < len(args) && isDigits(args[index+1]):
			// Separate count, eg. -j 4; consume the digits, as GNU make does
			index++
			normalized = append(normalized, "-j="+args[index])
		default:
			normalized = append(normalized, arg)
		}
	}
	return normalized
}

// isDigits reports whether s is entirely ASCII digits, with at least one.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for index := 0; index < len(s); index++ {
		if s[index] < '0' || s[index] > '9' {
			return false
		}
	}
	return true
}

// run plays the requested benchmark and prints its results.
//
// Summaries are printed to stderr. Detailed results accumulate into a named
// JSON file, freshly updated as each interval elapses; naming several files
// skips the benchmark and merges instead.
func run(opts options) int {
	// Profile everything, JSON handling included; the game loop still dominates
	if opts.profile {
		stopProfile, err := startProfile("default.pgo")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		defer stopProfile()
	}

	// Claim the output file for the whole run, so two runs cannot share it
	if len(opts.jsonPaths) > 0 {
		unlock, err := lockResults(opts.jsonPaths[len(opts.jsonPaths)-1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		defer unlock()
	}

	// Read earlier results up front, so a bad path fails before a long run
	prior, err := readResults(opts.jsonPaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// Several files? Merge them into the last, playing no games at all.
	if len(opts.jsonPaths) > 1 {
		return merge(prior, opts.jsonPaths)
	}

	// A game-count target plays from a finite pool; a time limit plays an
	// effectively unbounded pool until the context deadline stops the workers.
	// Zero seconds sets no deadline at all, playing until interrupted.
	// An interrupt cancels any mode early, reporting the games played so far.
	// Beyond Ctrl+C's SIGINT, system shutdown sends SIGTERM and a closing terminal SIGHUP
	ctx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer stop()
	totalGames := opts.numGames
	switch {
	case opts.numGames > 0:
		fmt.Fprintf(os.Stderr, "Playing %s games of Snakes & Ladders ", comma(opts.numGames))
	case opts.seconds == 0:
		fmt.Fprint(os.Stderr, "Playing Snakes & Ladders until interrupted ")
		totalGames = math.MaxInt64
	default:
		fmt.Fprintf(os.Stderr, "Playing Snakes & Ladders for %d seconds ", opts.seconds)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.seconds)*time.Second)
		defer cancel()
		totalGames = math.MaxInt64
	}

	if opts.jobs == 1 {
		fmt.Fprintln(os.Stderr, "with a single goroutine.")
	} else {
		fmt.Fprintf(os.Stderr, "using %d goroutines.\n", opts.jobs)
	}

	// Play in progress-interval-long segments, reporting and updating the file after each
	logger := newProgressLogger()
	start := time.Now()
	var result ladders.BenchmarkResult
	for {
		// In seconds mode the parent deadline caps the final segment
		segmentStart := time.Now()
		segmentCtx, cancel := context.WithTimeout(ctx, opts.progress)
		segment := ladders.Run(segmentCtx, opts.jobs, totalGames-result.NumGames)
		cancel()
		segmentWall := time.Since(segmentStart)
		result = result.Add(segment)
		result.Wall = time.Since(start).Seconds()

		// Each update rewrites the whole snapshot, depending on no earlier one
		if len(opts.jsonPaths) == 1 {
			if err := writeResults(prior.Add(result), opts.jsonPaths[0]); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				return 1
			}
		}
		if ctx.Err() != nil || result.NumGames >= totalGames {
			break
		}

		// Still going: report a tick; the final segment is left to the summary
		reportProgress(logger, result.NumGames, segment.NumGames, segmentWall, time.Since(start))
	}

	// Note interruption before calling stop, as stop itself cancels the context
	interrupted := ctx.Err() == context.Canceled

	// Restore default signal handling, so a second interrupt kills immediately
	stop()
	if interrupted {
		switch {
		case opts.numGames > 0:
			fmt.Fprintf(os.Stderr, "Interrupted after %s of %s games.\n",
				comma(result.NumGames), comma(opts.numGames))
		case opts.seconds == 0:
			fmt.Fprintf(os.Stderr, "Interrupted after %.2f seconds.\n", result.Wall)
		default:
			fmt.Fprintf(os.Stderr, "Interrupted after %.2f of %d seconds.\n", result.Wall, opts.seconds)
		}
	}

	if code := printSummary(result); code != 0 {
		return code
	}

	// Exit 130 mimics the shell's 128 plus signal number convention for SIGINT
	if interrupted {
		return 130
	}
	return 0
}

// merge reports the combined results from several files, writing them to the last.
//
// The combining itself happens as the files are read; here the totals are
// presented just as a benchmark run's would be.
func merge(combined ladders.BenchmarkResult, paths []string) int {
	target := paths[len(paths)-1]
	sources := strings.Join(paths[:len(paths)-1], ", ")
	fmt.Fprintf(os.Stderr, "Merging results from %s into %s\n", sources, target)
	if code := printSummary(combined); code != 0 {
		return code
	}
	if err := writeResults(combined, target); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

// newProgressLogger builds the charm logger used for progress ticks on stderr.
func newProgressLogger() *charmlog.Logger {
	return charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
}

// reportProgress logs one progress tick to the console.
//
// The line shows the games played so far, the rate over the segment just
// finished, and the elapsed wall time.
func reportProgress(logger *charmlog.Logger, played, segmentGames int64, segmentWall, elapsed time.Duration) {
	var rate int64
	if segmentWall > 0 {
		rate = int64(math.Round(float64(segmentGames) / segmentWall.Seconds()))
	}
	logger.Infof("%s games (%s/s), %s elapsed", comma(played), comma(rate), elapsed.Round(time.Second))
}

// printSummary prints a result's game count, timings, and move records to stderr.
func printSummary(result ladders.BenchmarkResult) int {
	// Guard the rate against a zero wall, as comes from merging empty results
	var rate float64
	if result.Wall > 0 {
		rate = float64(result.NumGames) / result.Wall
	}
	fmt.Fprintf(
		os.Stderr,
		"%s games finished in %.2f seconds (%.2fs worker time) = %s games per second\n",
		comma(result.NumGames), result.Wall, result.Elapsed, comma(int64(math.Round(rate))),
	)

	// An interrupt can arrive before any games at all; skip the empty statistics
	if result.NumGames > 0 {
		median, err := result.Median()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Fprintf(
			os.Stderr,
			"The shortest game took %d moves, the longest %d, while the median was %d.\n",
			len(result.Shortest), len(result.Longest), int(median),
		)
	}
	return 0
}

// writeResults writes a result to the given path as indented JSON.
//
// The data goes to a temporary file beside the target, renamed into place
// only once safely on disk, so a failed write cannot corrupt earlier results.
// Results that fail the consistency check are refused outright.
func writeResults(result ladders.BenchmarkResult, path string) error {
	// Catch double-count style bugs here, before they spread into the file
	if err := result.Validate(); err != nil {
		return fmt.Errorf("refusing to write inconsistent results: %w", err)
	}
	encoded, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return err
	}

	// The same directory as the target keeps the rename on one filesystem
	temp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	// Clean up after any failure; a no-op once the rename claims the file
	defer os.Remove(temp.Name())
	defer temp.Close()

	if _, err := temp.Write(append(encoded, '\n')); err != nil {
		return err
	}
	// CreateTemp makes a private file; match the mode WriteFile used to apply
	if err := temp.Chmod(0o644); err != nil {
		return err
	}
	// Force data to disk first, so a crash cannot publish an empty file
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

// comma formats an integer with thousands separators, eg. 1234567 becomes "1,234,567".
func comma(n int64) string {
	if n < 0 {
		// Negate in uint64 space, as the magnitude of math.MinInt64 overflows int64
		return "-" + commaUint64(uint64(-n))
	}
	return commaUint64(uint64(n))
}

// commaUint64 does the digit grouping for comma.
func commaUint64(n uint64) string {
	if n < 1000 {
		return strconv.FormatUint(n, 10)
	}
	return commaUint64(n/1000) + "," + fmt.Sprintf("%03d", n%1000)
}

func main() {
	opts, err := parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	os.Exit(run(opts))
}
