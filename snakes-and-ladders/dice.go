package main

import (
	"math/bits"
	"math/rand/v2"
)

// batchSize is 23 rolls per word: 24 would fit, but wastes 23% of words to rejection, not 1.5%.
const batchSize = 23

// sixPow23 is 6^23, the number of distinct batchSize-roll batches.
const sixPow23 = 789_730_223_053_602_816

// rejectBelow is 2^64 mod 6^23, the number of words left over after dealing 23 to each batch.
const rejectBelow = 282_948_943_476_686_848

// D6 deals exactly uniform rolls, drawing 23 per 64-bit word.
type D6 struct {
	rng       *rand.Rand
	batch     uint64
	rollsLeft int
}

// roll returns the next D6 roll, 1 to 6.
func (d *D6) roll() int {
	if d.rollsLeft == 0 {
		d.refill()
	}
	// The high word of x*6 is the roll, the low word the next x; all 23 digits are exact.
	hi, lo := bits.Mul64(d.batch, 6)
	d.batch = lo
	d.rollsLeft--
	return int(hi) + 1
}

// refill draws a word whose top 23 base-6 digits are exactly uniform.
func (d *D6) refill() {
	// A word x yields the digits of floor(x*6^23 / 2^64), but as 6^23 does not divide
	// 2^64 some values get 24 words while the rest get 23. Rejecting words whose product
	// fraction falls below 2^64 mod 6^23 (Lemire's method) trims every value to exactly
	// 23 words, so all 6^23 batches are equally likely. About 1.5% of words are rejected.
	x := d.rng.Uint64()
	for _, lo := bits.Mul64(x, sixPow23); lo < rejectBelow; _, lo = bits.Mul64(x, sixPow23) {
		x = d.rng.Uint64()
	}
	d.batch = x
	d.rollsLeft = batchSize
}

// Creating other dice
//
// The batching scheme generalises to any N-sided die. A 64-bit word is read as a
// fixed-point fraction, and each roll peels off one base-N digit: the high word of
// batch*N is the roll, the low word is the next batch. Only three constants change:
//
//   - batchSize: digits drawn per word.
//   - nPowK: N^batchSize, the number of distinct batches.
//   - rejectBelow: 2^64 mod nPowK, the words rejected so every batch is equally likely.
//
// The rejection rate is rejectBelow/2^64, so compare candidate sizes by useful rolls
// per word, batchSize*(1 - rate), rather than assuming the largest that fits is best:
// the D6 gets 22.7 rolls per word from 23 digits but only 18.5 from 24. Power-of-two
// dice need no rejection at all - a D8 roll is just three bits masked off the word.
//
// For a D20, 14 digits (2.30% rejection, 13.7 rolls per word) beats 13 (0.08%, 13.0):
//
//	const d20BatchSize = 14
//	const twentyPow14 = 1_638_400_000_000_000_000
//	const d20RejectBelow = 424_344_073_709_551_616
//
// roll and refill are then copies of the D6 versions with 6 and sixPow23 swapped for
// 20 and twentyPow14.
