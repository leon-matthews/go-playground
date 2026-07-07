// AVX-512 twins of Add, Contains and locate, using the experimental simd
// package. This file only compiles under GOEXPERIMENT=simd; a stock build
// gets the scalar paths in filter.go alone.

//go:build goexperiment.simd && amd64

package filter

import (
	"encoding/binary"
	"simd/archsimd"
)

// laneIndex numbers the eight vector lanes, so one round covers eight probes.
var laneIndex = [wordsPerBlock]uint64{0, 1, 2, 3, 4, 5, 6, 7}

// addAVX inserts a hash exactly as [Filter.Add] does, writing the whole block at once.
// Warning: not safe for concurrent use, and faults on CPUs without AVX-512.
func (f *Filter) addAVX(hash SHA1Hash) {
	base, masks := locateAVX(hash, f.mask, f.probes)
	block := (*[wordsPerBlock]uint64)(f.blocks[base:])
	archsimd.LoadUint64x8(block).Or(masks).Store(block)
	f.NumEntries++
}

// containsAVX answers exactly as [Filter.Contains] does, checking the whole block at once.
// Faults on CPUs without AVX-512.
func (f *Filter) containsAVX(hash SHA1Hash) bool {
	base, masks := locateAVX(hash, f.mask, f.probes)
	block := archsimd.LoadUint64x8((*[wordsPerBlock]uint64)(f.blocks[base:]))
	// The entry can only exist if every probe bit is set in its word
	return masks.And(block).Equal(masks).ToBits() == 0xff
}

// locateAVX mirrors locate, building the eight word masks in one vector.
//
// Lane j of round c covers probe c*8+j, landing on the same word and bit that
// the scalar loop reaches on that iteration, so the masks are bit-identical.
func locateAVX(hash SHA1Hash, mask uint64, probes int) (base uint64, masks archsimd.Uint64x8) {
	blockHash := binary.LittleEndian.Uint64(hash[0:8])
	position := binary.LittleEndian.Uint64(hash[8:16])
	step := uint64(binary.LittleEndian.Uint32(hash[16:20]))<<1 | 1
	base = (blockHash & mask) * wordsPerBlock

	// Eight consecutive probe positions at once: position + {0..7} * step
	positions := archsimd.BroadcastUint64x8(position).
		Add(archsimd.LoadUint64x8(&laneIndex).Mul(archsimd.BroadcastUint64x8(step)))
	stride := archsimd.BroadcastUint64x8(8 * step)
	one := archsimd.BroadcastUint64x8(1)
	low6 := archsimd.BroadcastUint64x8(63)

	for range probes / wordsPerBlock {
		masks = masks.Or(one.ShiftLeft(positions.And(low6)))
		positions = positions.Add(stride)
	}
	// A final partial round must only fill the lanes below probes%8
	if tail := probes % wordsPerBlock; tail != 0 {
		lanes := archsimd.Mask64x8FromBits(uint8(1)<<tail - 1)
		masks = masks.Or(one.ShiftLeft(positions.And(low6)).Masked(lanes))
	}
	return base, masks
}
