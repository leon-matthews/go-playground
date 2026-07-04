// Package filter implements a split-block Bloom filter over 20-byte SHA-1
// hashes, tuned for fast, lock-free, parallel membership queries.
//
// Each element maps to a single 512-bit block (one cache line) and sets a tunable
// number of probe bits spread round-robin across the block's eight 64-bit words. A
// query therefore touches exactly one cache line. Probe positions are generated
// from the SHA-1 by enhanced double hashing; the digest is already uniformly
// random, so no separate hash function is needed.
//
// A built filter is immutable: [Filter.Contains] only reads, so any number of
// goroutines may query it concurrently without synchronisation.
package filter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"os"
	"syscall"
	"unsafe"
)

const (
	magic         = "PWNDBF01"
	version       = 2
	headerSize    = 4096 // one page, so the mmap'd data region is page-aligned
	bytesPerBlock = 64   // 512 bits, one x86-64 cache line
	wordsPerBlock = 8
)

// ErrStale reports that the filter was built from a different source database
// than the one presented, so it must be rebuilt.
var ErrStale = errors.New("filter is stale: source database has changed")

// Filter is a split-block Bloom filter. The zero value is not usable; build one
// with [New] or load one with [Open].
type Filter struct {
	blocks  []uint64 // NumBlocks * wordsPerBlock words
	mask    uint64   // NumBlocks - 1, for the block index
	probes  int      // bits set per element, spread across the eight words
	mmapped []byte   // backing mmap region, non-nil when loaded from disk
	file    *os.File // backing file, non-nil when loaded from disk

	// Elements is the number of hashes added to the filter.
	Elements uint64
	// NumBlocks is the block count, always a power of two.
	NumBlocks uint64
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

// New allocates an empty in-memory filter with numBlocks blocks, which must be a
// power of two, setting probes bits per element.
func New(numBlocks uint64, probes int) (*Filter, error) {
	if numBlocks == 0 || numBlocks&(numBlocks-1) != 0 {
		return nil, fmt.Errorf("numBlocks must be a power of two, got %d", numBlocks)
	}
	if probes < 1 {
		return nil, fmt.Errorf("probes must be at least 1, got %d", probes)
	}
	return &Filter{
		blocks:    make([]uint64, numBlocks*wordsPerBlock),
		mask:      numBlocks - 1,
		probes:    probes,
		NumBlocks: numBlocks,
	}, nil
}

// locate derives the block's word offset and the per-word bit masks from a hash.
// The block index comes from the first 8 bytes; the probe positions are generated
// by enhanced double hashing over the remaining 12 and dealt round-robin to the
// eight words, so each word receives one to three bits.
func locate(hash []byte, mask uint64, probes int) (base uint64, masks [wordsPerBlock]uint64) {
	h1 := binary.LittleEndian.Uint64(hash[0:8])
	pos := binary.LittleEndian.Uint64(hash[8:16])
	step := uint64(binary.LittleEndian.Uint32(hash[16:20]))<<1 | 1 // odd, coprime to 64
	base = (h1 & mask) * wordsPerBlock
	for j := range probes {
		masks[j&(wordsPerBlock-1)] |= uint64(1) << (pos & 63)
		pos += step
	}
	return base, masks
}

// Add inserts a 20-byte hash. It is not safe for concurrent use and is only
// called while building, before the filter is published to readers.
func (f *Filter) Add(hash []byte) {
	base, masks := locate(hash, f.mask, f.probes)
	for i := range wordsPerBlock {
		f.blocks[base+uint64(i)] |= masks[i]
	}
}

// Contains reports whether a 20-byte hash may be present. False positives are
// possible; false negatives are not. It is read-only and safe for any number of
// goroutines to call concurrently.
func (f *Filter) Contains(hash []byte) bool {
	base, masks := locate(hash, f.mask, f.probes)
	for i := range wordsPerBlock {
		if f.blocks[base+uint64(i)]&masks[i] != masks[i] {
			return false
		}
	}
	return true
}

// Write serialises the filter to path via a temporary file and an atomic rename.
// The header records the size and modification time of sourcePath so a later
// [Open] can detect a stale filter.
func (f *Filter) Write(path, sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", sourcePath, err)
	}

	tmp := path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	var header [headerSize]byte
	copy(header[0:8], magic)
	binary.LittleEndian.PutUint32(header[8:12], version)
	binary.LittleEndian.PutUint64(header[12:20], f.NumBlocks)
	binary.LittleEndian.PutUint64(header[20:28], f.Elements)
	binary.LittleEndian.PutUint64(header[28:36], uint64(info.Size()))
	binary.LittleEndian.PutUint64(header[36:44], uint64(info.ModTime().UnixNano()))
	binary.LittleEndian.PutUint32(header[44:48], uint32(f.probes))
	if _, err := file.Write(header[:]); err != nil {
		file.Close()
		return err
	}

	// Reinterpret the words as bytes and write in bounded chunks
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&f.blocks[0])), len(f.blocks)*8)
	const chunk = 128 << 20
	for off := 0; off < len(raw); off += chunk {
		end := min(off+chunk, len(raw))
		if _, err := file.Write(raw[off:end]); err != nil {
			file.Close()
			return err
		}
	}

	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Open memory-maps the filter at path read-only. It returns [ErrStale] if the
// filter was built from a different sourcePath than the current one.
func Open(path, sourcePath string) (*Filter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var header [headerSize]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		file.Close()
		return nil, fmt.Errorf("reading filter header: %w", err)
	}
	if string(header[0:8]) != magic {
		file.Close()
		return nil, fmt.Errorf("%q is not a filter file", path)
	}
	if v := binary.LittleEndian.Uint32(header[8:12]); v != version {
		file.Close()
		return nil, fmt.Errorf("unsupported filter version %d", v)
	}
	numBlocks := binary.LittleEndian.Uint64(header[12:20])
	elements := binary.LittleEndian.Uint64(header[20:28])
	sourceSize := binary.LittleEndian.Uint64(header[28:36])
	sourceMtime := binary.LittleEndian.Uint64(header[36:44])
	probes := binary.LittleEndian.Uint32(header[44:48])

	info, err := os.Stat(sourcePath)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("stat source %q: %w", sourcePath, err)
	}
	if uint64(info.Size()) != sourceSize || uint64(info.ModTime().UnixNano()) != sourceMtime {
		file.Close()
		return nil, ErrStale
	}

	dataLen := int(numBlocks * bytesPerBlock)
	if fi, err := file.Stat(); err != nil {
		file.Close()
		return nil, err
	} else if fi.Size() != headerSize+int64(dataLen) {
		file.Close()
		return nil, fmt.Errorf("filter %q is truncated", path)
	}

	data, err := syscall.Mmap(int(file.Fd()), headerSize, dataLen, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("mmap filter: %w", err)
	}
	// Queries are uniformly random, so suppress read-ahead
	_ = syscall.Madvise(data, syscall.MADV_RANDOM)

	blocks := unsafe.Slice((*uint64)(unsafe.Pointer(&data[0])), len(data)/8)
	return &Filter{
		blocks:    blocks,
		mask:      numBlocks - 1,
		probes:    int(probes),
		mmapped:   data,
		file:      file,
		Elements:  elements,
		NumBlocks: numBlocks,
	}, nil
}

// Close releases a mmap-backed filter's resources. It is a no-op for an
// in-memory filter built with [New].
func (f *Filter) Close() error {
	if f.mmapped != nil {
		if err := syscall.Munmap(f.mmapped); err != nil {
			return err
		}
		f.mmapped = nil
	}
	if f.file != nil {
		err := f.file.Close()
		f.file = nil
		return err
	}
	return nil
}
