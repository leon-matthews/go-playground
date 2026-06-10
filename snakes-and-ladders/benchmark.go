package main

import (
	"fmt"
	"iter"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"time"
)

// gameCounts maps game length against the number of games of that length.
type gameCounts map[int]int

// MarshalJSON writes the counts with keys in ascending game-length order.
func (c gameCounts) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, length := range slices.Sorted(maps.Keys(c)) {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `"%d":%d`, length, c[length])
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

// BenchmarkResult holds the results for a benchmark run.
type BenchmarkResult struct {
	// Counts maps game length against number of games.
	Counts gameCounts `json:"counts"`
	// Elapsed is the time taken in seconds.
	Elapsed float64 `json:"elapsed"`
	// NumGames is the total number of games played.
	NumGames int `json:"num_games"`
	// Shortest is the full roll and position history of the shortest game played.
	Shortest Game `json:"shortest"`
	// Longest is as per Shortest, but for the longest game.
	Longest Game `json:"longest"`
}

// Add combines two results and creates a new one.
func (r BenchmarkResult) Add(other BenchmarkResult) BenchmarkResult {
	counts := make(gameCounts, max(len(r.Counts), len(other.Counts)))
	for length, count := range r.Counts {
		counts[length] += count
	}
	for length, count := range other.Counts {
		counts[length] += count
	}
	return BenchmarkResult{
		Counts:   counts,
		Elapsed:  r.Elapsed + other.Elapsed,
		NumGames: r.NumGames + other.NumGames,
		Shortest: shorterGame(r.Shortest, other.Shortest),
		Longest:  longerGame(r.Longest, other.Longest),
	}
}

// shorterGame returns the shorter of two games, ignoring empty ones.
func shorterGame(a, b Game) Game {
	if len(a) == 0 {
		return b
	}
	if len(b) != 0 && len(b) < len(a) {
		return b
	}
	return a
}

// longerGame returns the longer of two games.
func longerGame(a, b Game) Game {
	if len(b) > len(a) {
		return b
	}
	return a
}

// playCount plays the given number of solo games of snakes and ladders.
//
// Returns a BenchmarkResult containing the shortest and longest games.
func playCount(rng *rand.PCG, numGames int) BenchmarkResult {
	counts := make(gameCounts)
	var shortest, longest Game

	// Reuse one buffer for every game, keeping copies of record-breaking games only
	moves := make(Game, 0, 512)

	start := time.Now()
	for range numGames {
		moves = snakesAndLadders(rng, moves)
		numMoves := len(moves)
		counts[numMoves]++
		if shortest == nil || numMoves < len(shortest) {
			shortest = slices.Clone(moves)
		}
		if numMoves > len(longest) {
			longest = slices.Clone(moves)
		}
	}
	elapsed := time.Since(start).Seconds()

	return BenchmarkResult{
		Counts:   counts,
		Elapsed:  elapsed,
		NumGames: numGames,
		Shortest: shortest,
		Longest:  longest,
	}
}

// playTime keeps playing solo snakes and ladders for at least the given time.
//
// The goal is to play a round number of games while minimising the time
// keeping overhead.
func playTime(rng *rand.PCG, seconds int) BenchmarkResult {
	const minimum = 100
	var result BenchmarkResult
	for totalGames := range currencySeries(minimum) {
		count := totalGames - result.NumGames
		result = result.Add(playCount(rng, count))
		if result.Elapsed > float64(seconds) {
			break
		}
	}
	return result
}

// currencySeries produces a readable series of numbers that is roughly exponential.
//
//	1, 2, 5, 10, 20, 50, 100, 200, etc.
//
// Grows a little faster than a power of two series, reaching one million
// after 19 iterations, rather than 20. The series starts at the first value
// greater than or equal to start.
func currencySeries(start int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for multiplier := 1; ; multiplier *= 10 {
			for _, s := range [...]int{1, 2, 5} {
				if value := s * multiplier; value >= start && !yield(value) {
					return
				}
			}
		}
	}
}

// benchmarkParallel runs the given benchmark function on numJobs goroutines
// and combines the results.
//
// Every goroutine runs the full function with the same argument, just as the
// Python original runs its benchmark function once per process, so the
// combined result holds numJobs times the requested work.
func benchmarkParallel(numJobs int, function func(*rand.PCG, int) BenchmarkResult, argument int) BenchmarkResult {
	// Start jobs, each with its own random number generator
	results := make(chan BenchmarkResult, numJobs)
	for range numJobs {
		go func() {
			rng := rand.NewPCG(rand.Uint64(), rand.Uint64())
			results <- function(rng, argument)
		}()
	}

	// Wait for, and combine results
	var combined BenchmarkResult
	for range numJobs {
		combined = combined.Add(<-results)
	}
	return combined
}
