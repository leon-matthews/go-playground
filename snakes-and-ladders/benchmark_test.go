package main

import (
	"math/rand/v2"
	"testing"
)

// BenchmarkPlayCount measures playCount throughput in games per second.
func BenchmarkPlayCount(b *testing.B) {
	// Fixed seed keeps runs comparable between benchmark sessions
	rng := rand.NewPCG(1, 2)
	const gamesPerOp = 1_000

	numGames := 0
	for b.Loop() {
		playCount(rng, gamesPerOp)
		numGames += gamesPerOp
	}
	b.ReportMetric(float64(numGames)/b.Elapsed().Seconds(), "games/s")
}
