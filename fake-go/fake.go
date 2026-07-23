// Package fake generates realistic random data for tests and fixtures.
//
// It is a Go port of the Python "fake" package, covering numbers, text, people,
// addresses, and relative dates. All generators are methods on a Faker, which
// holds its own random source so output can be made reproducible by seeding.
package fake

import "math/rand/v2"

// Faker produces random data from a single, optionally seeded, source.
//
// A Faker is not safe for concurrent use; create one per goroutine.
type Faker struct {
	rng *rand.Rand
}

// New returns a Faker seeded with the given value.
//
// Two Fakers created with the same seed produce identical sequences, which is
// useful for reproducible fixtures.
func New(seed uint64) *Faker {
	return &Faker{rng: rand.New(rand.NewPCG(seed, seed))}
}

// NewRandom returns a Faker seeded from the operating system's entropy.
//
// Its output is not reproducible; use New when you need a repeatable sequence.
func NewRandom() *Faker {
	return &Faker{rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}

// count returns a random integer in the inclusive range [low, high].
//
// Unlike Int it permits low == high, matching the Python get_count helper used
// for "a fixed count, or a range" arguments.
func (f *Faker) count(low, high int) int {
	if low > high {
		panic("fake: low is greater than high")
	}
	return low + f.rng.IntN(high-low+1)
}
