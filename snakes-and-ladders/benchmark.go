package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

// gameCounts records how many games finished at each length, indexed by length.
//
// Counts are int64 so that long runs do not overflow on 32-bit builds.
type gameCounts []int64

// MarshalJSON writes the counts as an object with keys in ascending game-length order.
func (c gameCounts) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for length, count := range c {
		if count == 0 {
			continue
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		fmt.Fprintf(&b, `"%d":%d`, length, count)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

// BenchmarkResult holds the results for a benchmark run.
type BenchmarkResult struct {
	// Counts maps game length against number of games.
	Counts gameCounts `json:"counts"`
	// Elapsed is the seconds spent playing; combined results sum every worker's span.
	Elapsed float64 `json:"elapsed"`
	// NumGames is the total number of games played.
	NumGames int64 `json:"num_games"`
	// Shortest is the full roll and position history of the shortest game played.
	Shortest Game `json:"shortest"`
	// Longest is as per Shortest, but for the longest game.
	Longest Game `json:"longest"`
}

// Add combines two results and creates a new one.
func (r BenchmarkResult) Add(other BenchmarkResult) BenchmarkResult {
	counts := make(gameCounts, max(len(r.Counts), len(other.Counts)))
	copy(counts, r.Counts)
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

// chunkGames is how many games a worker claims at once, bounding deadline overshoot and imbalance.
const chunkGames = 1024

// progressInterval is how often benchmarkParallel reports progress on long runs.
const progressInterval = 10 * time.Second

// playGames plays solo games of snakes and ladders until the work runs out.
//
// Games are claimed from the shared remaining counter, one chunk at a time,
// until the counter is exhausted or the context is cancelled. Returns the
// games actually played, including the shortest and longest seen.
func playGames(ctx context.Context, rng *rand.PCG, remaining *atomic.Int64) BenchmarkResult {
	// Counts are indexed by game length; 512 covers all but the freakiest games
	counts := make(gameCounts, 512)
	var shortest, longest Game
	var played int64

	// Reuse one buffer for every game, keeping copies of record-breaking games only
	moves := make(Game, 0, 512)

	start := time.Now()
	for ctx.Err() == nil {
		// Claim the next chunk; the counter goes negative once the work runs out
		games := min(chunkGames, remaining.Add(-chunkGames)+chunkGames)
		if games <= 0 {
			break
		}
		for range games {
			moves = snakesAndLadders(rng, moves)
			numMoves := len(moves)
			if numMoves >= len(counts) {
				counts = append(counts, make(gameCounts, numMoves+1-len(counts))...)
			}
			counts[numMoves]++
			if shortest == nil || numMoves < len(shortest) {
				shortest = slices.Clone(moves)
			}
			if numMoves > len(longest) {
				longest = slices.Clone(moves)
			}
		}
		played += games
	}
	elapsed := time.Since(start).Seconds()

	// Trim the zero tail so equal results compare equal and marshal compactly
	for len(counts) > 0 && counts[len(counts)-1] == 0 {
		counts = counts[:len(counts)-1]
	}

	return BenchmarkResult{
		Counts:   counts,
		Elapsed:  elapsed,
		NumGames: played,
		Shortest: shortest,
		Longest:  longest,
	}
}

// benchmarkParallel plays totalGames games shared between numJobs goroutines
// and combines their results.
//
// Workers claim work in small chunks from a single pool, so they all finish
// within one chunk of each other, and of the context deadline if one is set.
// While the run is going, the progress callback is given the games claimed
// so far, once every progressInterval.
func benchmarkParallel(ctx context.Context, numJobs int, totalGames int64, progress func(played int64)) BenchmarkResult {
	var remaining atomic.Int64
	remaining.Store(totalGames)

	// Start jobs, each with its own random number generator
	results := make(chan BenchmarkResult, numJobs)
	for range numJobs {
		go func() {
			rng := rand.NewPCG(rand.Uint64(), rand.Uint64())
			results <- playGames(ctx, rng, &remaining)
		}()
	}

	// Merge results as workers finish, reporting progress in between
	ticker := time.NewTicker(progressInterval)
	defer ticker.Stop()
	var combined BenchmarkResult
	for finished := 0; finished < numJobs; {
		select {
		case result := <-results:
			combined = combined.Add(result)
			finished++
		case <-ticker.C:
			// Claimed work trails actual play by at most one chunk per worker
			progress(min(totalGames, totalGames-remaining.Load()))
		}
	}
	return combined
}
