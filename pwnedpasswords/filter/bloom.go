package filter

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

const (
	bytesPerBlock = 64 // 512 bits, one x86-64 cache line
	wordsPerBlock = 8
)

// SHA1Hash is a 20-byte SHA-1 digest, the sole element type of the filter.
type SHA1Hash [20]byte

// SplitBlockBloom is an in-memory split-block Bloom filter.
//
// Each element maps to a single 512-bit block (one cache line) and sets a tunable
// number of probe bits spread round-robin across the block's eight 64-bit words. A
// query therefore touches exactly one cache line. Probe positions are generated
// from the SHA-1 by double hashing; the digest is already uniformly random, so
// no separate hash function is needed.
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
	if probes < 1 {
		return SplitBlockBloom{}, fmt.Errorf("probes must be at least 1, got %d", probes)
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

// locate picks both which block to use, and which bits to use inside that block.
//
// It does so deterministically from the input hash's bytes, as we already know
// they are randomly distributed - they are already from a good hash function.
//
// It's hard to read because of all the bit twiddling, but a Bloom filter is
// fundamentally a bit-based data structure after all!
func locate(hash SHA1Hash, mask uint64, probes int) (base uint64, masks [wordsPerBlock]uint64) {
	// Extract core Bloom filter values directly from input, as it's already a hash
	blockHash := binary.LittleEndian.Uint64(hash[0:8])
	position := binary.LittleEndian.Uint64(hash[8:16])

	// <<1 clears the low bit (making it even), then | 1 sets it forcing it to be odd
	step := uint64(binary.LittleEndian.Uint32(hash[16:20]))<<1 | 1

	// blockHash & mask means blockHash % NumBlocks because mask = NumBlocks - 1
	base = (blockHash & mask) * wordsPerBlock

	// Build bit masks used later to set or check the bits inside block
	for j := range probes {
		masks[j&(wordsPerBlock-1)] |= uint64(1) << (position & 63)
		position += step
	}
	return base, masks
}
