package dice

import (
	"math/rand/v2"
	"testing"
)

// BenchmarkD6Roll measures the batched D6.Roll, drawing 23 rolls per word.
func BenchmarkD6Roll(b *testing.B) {
	// Fixed seed keeps runs comparable between benchmark sessions
	d6 := NewD6(rand.New(rand.NewPCG(1, 2)))
	for b.Loop() {
		d6.Roll()
	}
}

// BenchmarkRandIntN measures a plain rand.IntN call, one word per roll, for comparison.
func BenchmarkRandIntN(b *testing.B) {
	// Same PCG source as the D6, so only the batching scheme differs
	rng := rand.New(rand.NewPCG(1, 2))
	for b.Loop() {
		rng.IntN(6)
	}
}
