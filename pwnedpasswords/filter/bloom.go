package filter

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

const (
	bytesPerBlock = 64 // 512 bits, one x86-64 cache line
	wordsPerBlock = 8
	wordIndexBits = 6  // bits needed to address one bit within a 64-bit word
	maxProbes     = 64 // upper bound on probes, sizing the salt table
)

// SHA1Hash is a 20-byte SHA-1 digest, the sole element type of the filter.
type SHA1Hash [20]byte

// SplitBlockBloom is an in-memory split-block Bloom filter.
//
// Each element maps to a single 512-bit block (one cache line) and sets a tunable
// number of probe bits spread round-robin across the block's eight 64-bit words. A
// query therefore touches exactly one cache line. Each probe's bit is chosen by
// multiply-shift hashing the digest against that probe's fixed salt, so every
// probe draws on the whole digest rather than a handful of low bits.
//
// A built filter is immutable: [SplitBlockBloom.Contains] only reads, so any
// number of goroutines may query it concurrently without synchronisation.
type SplitBlockBloom struct {
	blocks []uint64 // NumBlocks * wordsPerBlock words
	mask   uint64   // NumBlocks - 1, for the block index
	probes int      // bits set per element, spread across the eight words

	// NumEntries is the number of hashes added to the filter.
	NumEntries uint64

	// NumBlocks is the block count, always a power of two.
	NumBlocks uint64
}

// newSplitBlockBloom allocates an empty filter with numBlocks blocks, which
// must be a power of two, setting probes bits per element.
func newSplitBlockBloom(numBlocks uint64, probes int) (SplitBlockBloom, error) {
	if numBlocks == 0 || numBlocks&(numBlocks-1) != 0 {
		return SplitBlockBloom{}, fmt.Errorf("numBlocks must be a power of two, got %d", numBlocks)
	}
	if probes < 1 || probes > maxProbes {
		return SplitBlockBloom{}, fmt.Errorf("probes must be between 1 and %d, got %d", maxProbes, probes)
	}
	return SplitBlockBloom{
		blocks:    make([]uint64, numBlocks*wordsPerBlock),
		mask:      numBlocks - 1,
		probes:    probes,
		NumBlocks: numBlocks,
	}, nil
}

// BlocksForBytes returns the largest power-of-two block count that fits within
// size bytes.
func BlocksForBytes(size uint64) uint64 {
	blocks := size / bytesPerBlock
	if blocks < 1 {
		return 1
	}
	return uint64(1) << (bits.Len64(blocks) - 1)
}

// Add inserts a hash, counting it in NumEntries.
//
// Find the right block and set b.probes of its bits.
// Warning: not safe for concurrent use.
func (b *SplitBlockBloom) Add(hash SHA1Hash) {
	base, masks := locate(hash, b.mask, b.probes)
	for i := range wordsPerBlock {
		b.blocks[base+uint64(i)] |= masks[i]
	}
	b.NumEntries++
}

// Contains reports whether a hash may be present.
//
// May be safely called concurrently.
// False positives are possible; false negatives are not.
func (b *SplitBlockBloom) Contains(hash SHA1Hash) bool {
	base, masks := locate(hash, b.mask, b.probes)
	for i := range wordsPerBlock {
		// If any bits from mask are not set, the entry CANNOT exist
		if b.blocks[base+uint64(i)]&masks[i] != masks[i] {
			return false
		}
	}

	// If all bits from mask are set, it's entry PROBABLY exists
	return true
}

// probeSalts holds one fixed odd multiplier per probe. locate mixes the element
// seed against a distinct salt for each probe, so the probes fall on independent
// bits instead of the low-entropy arithmetic progression an in-block double hash
// would give.
var probeSalts = makeProbeSalts()

// makeProbeSalts derives the salt table with splitmix64, a deterministic mixer,
// so the multipliers are well spread yet fixed and reproducible.
func makeProbeSalts() [maxProbes]uint64 {
	var salts [maxProbes]uint64
	state := uint64(0x9E3779B97F4A7C15)
	for i := range salts {
		state += 0x9E3779B97F4A7C15
		z := state
		z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
		z = (z ^ (z >> 27)) * 0x94D049BB133111EB
		z ^= z >> 31
		salts[i] = z | 1 // force odd, as multiply-shift requires
	}
	return salts
}

// locate picks both which block to use, and which bits to use inside that block.
//
// The block index comes from the first eight digest bytes. Each probe then sets
// one bit in a round-robin word, taking the bit from the top bits of a second
// eight-byte seed multiplied by that probe's salt. Mixing the whole 64-bit seed
// on every probe is what keeps the false-positive rate at its designed level.
func locate(hash SHA1Hash, mask uint64, probes int) (base uint64, masks [wordsPerBlock]uint64) {
	blockHash := binary.LittleEndian.Uint64(hash[0:8])
	seed := binary.LittleEndian.Uint64(hash[8:16])

	// blockHash & mask means blockHash % NumBlocks because mask = NumBlocks - 1
	base = (blockHash & mask) * wordsPerBlock

	for j := range probes {
		bit := (seed * probeSalts[j]) >> (64 - wordIndexBits)
		masks[j&(wordsPerBlock-1)] |= uint64(1) << bit
	}
	return base, masks
}
