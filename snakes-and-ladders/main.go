// Silly benchmark which plays many, many solo games of snakes and ladders.
//
// A Go port of the Python original, snakes_and_ladders.py, found in the
// parent directory.
//
// Copyright 2011-2026 Leon Matthews. Released under the Apache 2.0 licence.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// options holds the parsed command-line arguments.
type options struct {
	jobs     int
	json     bool
	numGames int
	seconds  int
}

// parse builds the parsed options from the given command-line arguments.
//
// Invalid arguments and requests for help exit the program directly, as the
// Python argparse module does.
func parse(args []string) (options, error) {
	flags := pflag.NewFlagSet("go_ladders", pflag.ExitOnError)
	flags.SetOutput(os.Stderr)

	// Multicore? Plain '-j' means all cores, but pflag needs '-j4' or '-j=4' for a count.
	numCores := runtime.NumCPU()
	jobs := flags.IntP("jobs", "j", 1, fmt.Sprintf("Run on multiple cores (%d found)", numCores))
	flags.Lookup("jobs").NoOptDefVal = strconv.Itoa(numCores)

	// JSON output?
	jsonOut := flags.Bool("json", false, "Dump detailed results to stdout as JSON")

	// Iterations or seconds?
	numGames := flags.Float64P("games", "n", 0, "Number of games to play, eg. 100 or 1e6")
	seconds := flags.IntP("seconds", "s", 10, "Approximate seconds to play for.")

	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.Changed("games") && flags.Changed("seconds") {
		return options{}, errors.New("only one of -n and -s may be given")
	}
	if flags.NArg() > 0 {
		return options{}, fmt.Errorf(
			"unrecognised arguments: %s (note that a core count needs '-j=4', not '-j 4')",
			strings.Join(flags.Args(), " "),
		)
	}
	if *jobs < 1 {
		return options{}, fmt.Errorf("number of jobs must be at least one, given: %d", *jobs)
	}

	// Guard the int conversion below; out-of-range conversions are implementation-dependent
	if math.IsNaN(*numGames) || *numGames < 0 || *numGames >= float64(math.MaxInt) {
		return options{}, fmt.Errorf("number of games out of range (0 to %d): %v", math.MaxInt, *numGames)
	}

	return options{
		jobs:     *jobs,
		json:     *jsonOut,
		numGames: int(*numGames),
		seconds:  *seconds,
	}, nil
}

// run plays the requested benchmark and prints its results.
//
// Summaries are printed to stderr, detailed JSON results to stdout.
func run(opts options) int {
	// Choose function
	var function func(*rand.PCG, int) BenchmarkResult
	var argument int
	if opts.numGames != 0 {
		fmt.Fprintf(os.Stderr, "Playing %s games of Snakes & Ladders ", comma(opts.numGames))
		function = playCount
		argument = opts.numGames
	} else {
		fmt.Fprintf(os.Stderr, "Playing Snakes & Ladders for at least %d seconds ", opts.seconds)
		function = playTime
		argument = opts.seconds
	}

	if opts.jobs == 1 {
		fmt.Fprintln(os.Stderr, "with a single goroutine.")
	} else {
		fmt.Fprintf(os.Stderr, "using %d goroutines.\n", opts.jobs)
	}

	// Run benchmark
	result := benchmarkParallel(opts.jobs, function, argument)
	elapsed := result.Elapsed / float64(opts.jobs)
	rate := float64(result.NumGames) / elapsed
	fmt.Fprintf(
		os.Stderr,
		"%s games finished in %.2f seconds (%.2fs CPU) = %s games per second\n",
		comma(result.NumGames), elapsed, result.Elapsed, comma(int(math.Round(rate))),
	)

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

	// JSON?
	if opts.json {
		encoded, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println(string(encoded))
	}

	return 0
}

// comma formats an integer with thousands separators, eg. 1234567 becomes "1,234,567".
func comma(n int) string {
	if n < 0 {
		return "-" + comma(-n)
	}
	if n < 1000 {
		return strconv.Itoa(n)
	}
	return comma(n/1000) + "," + fmt.Sprintf("%03d", n%1000)
}

func main() {
	opts, err := parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	os.Exit(run(opts))
}
