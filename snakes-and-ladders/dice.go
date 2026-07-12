package main

import (
	"math/bits"
	"math/rand/v2"
)

// D6 deals six-sided D6 rolls, drawing eight per call from a PCG generator.
type D6 struct {
	rng       *rand.PCG
	batch     uint64
	rollsLeft int
}

// roll returns the next D6 roll, 1 to 6.
func (d *D6) roll() int {
	if d.rollsLeft == 0 {
		d.batch, d.rollsLeft = d.rng.Uint64(), 8
	}
	// The high word of x*6 is the roll, the low word the next x; bias one part in 2^64/6^8
	hi, lo := bits.Mul64(d.batch, 6)
	d.batch = lo
	d.rollsLeft--
	return int(hi) + 1
}
