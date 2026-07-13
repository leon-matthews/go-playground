// Package ladders plays and benchmarks solo games of snakes and ladders.
package ladders

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"local.dev/ladders/dice"
)

// GameCounts records how many games finished at each length, indexed by length.
//
// Counts are int64 so that long runs do not overflow on 32-bit builds.
type GameCounts []int64

// MarshalJSON writes the counts as an object with keys in ascending game-length order.
func (c GameCounts) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON reads the length-keyed object form back into a counts slice.
func (c *GameCounts) UnmarshalJSON(data []byte) error {
	var byLength map[string]int64
	if err := json.Unmarshal(data, &byLength); err != nil {
		return err
	}
	var counts GameCounts
	for key, count := range byLength {
		length, err := strconv.Atoi(key)
		if err != nil || length < 0 {
			return fmt.Errorf("bad game length %q", key)
		}
		if count < 0 {
			return fmt.Errorf("negative count for game length %d: %d", length, count)
		}
		if length >= len(counts) {
			counts = append(counts, make(GameCounts, length+1-len(counts))...)
		}
		counts[length] = count
	}
	*c = counts
	return nil
}

// BenchmarkResult holds the results for a benchmark run.
type BenchmarkResult struct {
	// Counts maps game length against number of games.
	Counts GameCounts `json:"counts"`
	// Elapsed is the seconds spent playing; combined results sum every worker's span.
	Elapsed float64 `json:"elapsed"`
	// Wall is the wall-clock seconds taken; combined results sum every run's wall.
	Wall float64 `json:"wall"`
	// NumGames is the total number of games played.
	NumGames int64 `json:"num_games"`
	// Shortest is the full roll and position history of the shortest game played.
	Shortest Game `json:"shortest"`
	// Longest is as per Shortest, but for the longest game.
	Longest Game `json:"longest"`
}

// Add combines two results and creates a new one.
func (r BenchmarkResult) Add(other BenchmarkResult) BenchmarkResult {
	counts := make(GameCounts, max(len(r.Counts), len(other.Counts)))
	copy(counts, r.Counts)
	for length, count := range other.Counts {
		counts[length] += count
	}
	return BenchmarkResult{
		Counts:   counts,
		Elapsed:  r.Elapsed + other.Elapsed,
		Wall:     r.Wall + other.Wall,
		NumGames: r.NumGames + other.NumGames,
		Shortest: shorterGame(r.Shortest, other.Shortest),
		Longest:  longerGame(r.Longest, other.Longest),
	}
}

// Validate checks the cross-field consistency of a result.
//
// The counts must sum to the recorded game total, and the shortest and
// longest games must match the lowest and highest counted lengths.
func (r BenchmarkResult) Validate() error {
	var total int64
	first, last := 0, 0
	for length, count := range r.Counts {
		if count == 0 {
			continue
		}
		if first == 0 {
			first = length
		}
		last = length
		total += count
	}
	if total != r.NumGames {
		return fmt.Errorf("counts sum to %d games, but %d were recorded", total, r.NumGames)
	}
	if len(r.Shortest) != first {
		return fmt.Errorf("shortest game has %d moves, but the lowest count is for length %d",
			len(r.Shortest), first)
	}
	if len(r.Longest) != last {
		return fmt.Errorf("longest game has %d moves, but the highest count is for length %d",
			len(r.Longest), last)
	}
	return nil
}

// Median returns the high median game length across all games played.
//
// An error is returned if no games were played, as the median is then undefined.
func (r BenchmarkResult) Median() (float64, error) {
	return multisetMedian(r.Counts, medianHigh)
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

// playGames plays solo games of snakes and ladders until the work runs out.
//
// Games are claimed from the shared remaining counter, one chunk at a time,
// until the counter is exhausted or the context is cancelled. Returns the
// games actually played, including the shortest and longest seen.
func playGames(ctx context.Context, rng *rand.Rand, remaining *atomic.Int64) BenchmarkResult {
	// Counts are indexed by game length; 512 covers all but the freakiest games
	counts := make(GameCounts, 512)
	var shortest, longest Game
	var played int64

	// Reuse one buffer for every game, keeping copies of record-breaking games only
	moves := make(Game, 0, 512)

	// One die for the whole run, so rolls left in its batch carry over between games
	d6 := dice.NewD6(rng)

	start := time.Now()
	for ctx.Err() == nil {
		// Claim the next chunk; the counter goes negative once the work runs out
		games := min(chunkGames, remaining.Add(-chunkGames)+chunkGames)
		if games <= 0 {
			break
		}
		for range games {
			moves = snakesAndLadders(d6, moves)
			numMoves := len(moves)
			if numMoves >= len(counts) {
				counts = append(counts, make(GameCounts, numMoves+1-len(counts))...)
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

// Run plays totalGames games shared between numJobs goroutines and combines
// their results.
//
// Workers claim work in small chunks from a single pool, so they all finish
// within one chunk of each other, and of the context deadline if one is set.
func Run(ctx context.Context, numJobs int, totalGames int64) BenchmarkResult {
	var remaining atomic.Int64
	remaining.Store(totalGames)

	// Start jobs, each with its own random number generator
	results := make(chan BenchmarkResult, numJobs)
	for range numJobs {
		go func() {
			pcg := rand.NewPCG(rand.Uint64(), rand.Uint64())
			r := rand.New(pcg)
			results <- playGames(ctx, r, &remaining)
		}()
	}

	// Wait for, and combine results
	var combined BenchmarkResult
	for range numJobs {
		combined = combined.Add(<-results)
	}
	return combined
}
