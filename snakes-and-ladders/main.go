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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// options holds the parsed command-line arguments.
type options struct {
	jobs     int
	json     bool
	numGames int64
	seconds  int
}

// parse builds the parsed options from the given command-line arguments.
//
// Invalid arguments and requests for help exit the program directly, as the
// Python argparse module does.
func parse(args []string) (options, error) {
	flags := pflag.NewFlagSet("go_ladders", pflag.ExitOnError)
	flags.SetOutput(os.Stderr)

	// Multicore? Plain '-j' means all cores; normalizeJobs supports make-style counts.
	numCores := runtime.NumCPU()
	jobs := flags.IntP("jobs", "j", 1, fmt.Sprintf("Run on multiple cores (%d found)", numCores))
	flags.Lookup("jobs").NoOptDefVal = strconv.Itoa(numCores)

	// JSON output?
	jsonOut := flags.Bool("json", false, "Dump detailed results to stdout as JSON")

	// Iterations or seconds?
	numGames := flags.Int64P("games", "n", 0, "Total number of games to play, eg. 100 or 1_000_000")
	seconds := flags.IntP("seconds", "s", 10, "Seconds to play for")

	if err := flags.Parse(normalizeJobs(args)); err != nil {
		return options{}, err
	}
	if flags.Changed("games") && flags.Changed("seconds") {
		return options{}, errors.New("only one of -n and -s may be given")
	}
	if flags.NArg() > 0 {
		return options{}, fmt.Errorf("unrecognised arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *jobs < 1 {
		return options{}, fmt.Errorf("number of jobs must be at least one, given: %d", *jobs)
	}
	// The upper bound keeps the timeout within a time.Duration's count of nanoseconds
	const maxSeconds = math.MaxInt64 / int64(time.Second)
	if *seconds < 1 || int64(*seconds) > maxSeconds {
		return options{}, fmt.Errorf("number of seconds out of range (1 to %d), given: %d", maxSeconds, *seconds)
	}

	if flags.Changed("games") && *numGames < 1 {
		return options{}, fmt.Errorf("number of games must be at least one, given: %d", *numGames)
	}

	return options{
		jobs:     *jobs,
		json:     *jsonOut,
		numGames: *numGames,
		seconds:  *seconds,
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
// Summaries are printed to stderr, detailed JSON results to stdout.
func run(opts options) int {
	// A game-count target plays from a finite pool; a time limit plays an
	// effectively unbounded pool until the context deadline stops the workers.
	// An interrupt cancels either mode early, reporting the games played so far.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	totalGames := opts.numGames
	if opts.numGames > 0 {
		fmt.Fprintf(os.Stderr, "Playing %s games of Snakes & Ladders ", comma(opts.numGames))
	} else {
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

	// Run benchmark
	start := time.Now()
	progress := func(played int64) {
		elapsed := time.Since(start).Seconds()
		rate := comma(int64(math.Round(float64(played) / elapsed)))
		if opts.numGames > 0 {
			percent := 100 * float64(played) / float64(opts.numGames)
			fmt.Fprintf(os.Stderr, "%s of %s games (%.1f%%) at %s games per second\n",
				comma(played), comma(opts.numGames), percent, rate)
		} else {
			fmt.Fprintf(os.Stderr, "%s games after %.0f of %d seconds, at %s games per second\n",
				comma(played), elapsed, opts.seconds, rate)
		}
	}
	result := benchmarkParallel(ctx, opts.jobs, totalGames, progress)
	wall := time.Since(start).Seconds()

	// Note interruption before calling stop, as stop itself cancels the context
	interrupted := ctx.Err() == context.Canceled

	// Restore default signal handling, so a second interrupt kills immediately
	stop()
	if interrupted {
		if opts.numGames > 0 {
			fmt.Fprintf(os.Stderr, "Interrupted after %s of %s games.\n",
				comma(result.NumGames), comma(opts.numGames))
		} else {
			fmt.Fprintf(os.Stderr, "Interrupted after %.2f of %d seconds.\n", wall, opts.seconds)
		}
	}

	rate := float64(result.NumGames) / wall
	fmt.Fprintf(
		os.Stderr,
		"%s games finished in %.2f seconds (%.2fs worker time) = %s games per second\n",
		comma(result.NumGames), wall, result.Elapsed, comma(int64(math.Round(rate))),
	)

	// An interrupt can arrive before any games at all; skip the empty statistics
	if result.NumGames > 0 {
		median, err := multisetMedian(result.Counts, medianHigh)
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

	// JSON?
	if opts.json {
		// Run-level facts live here, not in the mergeable per-worker results
		detailed := struct {
			BenchmarkResult
			Wall        float64 `json:"wall"`
			Interrupted bool    `json:"interrupted"`
		}{result, wall, interrupted}
		encoded, err := json.MarshalIndent(detailed, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println(string(encoded))
	}

	// Exit 130 mimics the shell's 128 plus signal number convention for SIGINT
	if interrupted {
		return 130
	}
	return 0
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
