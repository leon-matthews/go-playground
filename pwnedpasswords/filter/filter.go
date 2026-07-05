// Package filter implements a split-block Bloom filter over 20-byte SHA-1
// hashes, tuned for fast, lock-free, parallel membership queries.
//
// Each element maps to a single 512-bit block (one cache line) and sets a tunable
// number of probe bits spread round-robin across the block's eight 64-bit words. A
// query therefore touches exactly one cache line. Probe positions are generated
// from the SHA-1 by double hashing; the digest is already uniformly random, so
// no separate hash function is needed.
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
	magic         = "PWNEDPASSWORDS"
	magicLength   = len(magic)
	version       = 2
	headerSize    = 4096 // header bytes before the data; multiple of 8 keeps words aligned
	bytesPerBlock = 64   // 512 bits, one x86-64 cache line
	wordsPerBlock = 8
)

// SHA1Hash is a 20-byte SHA-1 digest, the sole element type of the filter.
type SHA1Hash [20]byte

// ErrStale reports that the filter was built from a different source database
// than the one presented, so it must be rebuilt.
var ErrStale = errors.New("filter is stale: source database has changed")

// Filter is a split-block Bloom filter.
// The zero value is not usable; build one in memory with [New] or load one
// memory-mapped with [Open].
type Filter struct {
	blocks []uint64 // NumBlocks * wordsPerBlock words
	mask   uint64   // NumBlocks - 1, for the block index
	probes int      // bits set per element, spread across the eight words
	file   *os.File // backing file, non-nil when opened with Open

	// NumEntries is the number of hashes added to the filter.
	NumEntries uint64

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

// New allocates an empty in-memory filter with numBlocks blocks, which must be
// a power of two, setting probes bits per element.
// Persist a built filter with [Filter.Write].
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

// Add inserts a hash, counting it in NumEntries.
//
// Find the right block and set f.probes of its bits.
// Warning: not safe for concurrent use.
func (f *Filter) Add(hash SHA1Hash) {
	base, masks := locate(hash, f.mask, f.probes)
	for i := range wordsPerBlock {
		f.blocks[base+uint64(i)] |= masks[i]
	}
	f.NumEntries++
}

// Contains reports whether a hash may be present.
//
// May be safely called concurrently.
// False positives are possible; false negatives are not.
func (f *Filter) Contains(hash SHA1Hash) bool {
	base, masks := locate(hash, f.mask, f.probes)
	for i := range wordsPerBlock {
		// If any bits from mask are not set, the entry CANNOT exist
		if f.blocks[base+uint64(i)]&masks[i] != masks[i] {
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

// header is the metadata at the front of a filter file. Field order and types
// define the little-endian on-disk layout; do not reorder. The source fields
// fingerprint the database the filter was built from, letting Open detect a
// stale filter.
type header struct {
	Magic       [magicLength]byte
	Version     uint32
	NumBlocks   uint64
	NumEntries  uint64
	SourceSize  uint64
	SourceMtime uint64
	Probes      uint32
}

// encode lays the header out in its fixed on-disk format.
func (h header) encode() []byte {
	buf := make([]byte, headerSize)
	copy(h.Magic[:], magic)
	h.Version = version
	binary.Encode(buf, binary.LittleEndian, h) // cannot fail: buf exceeds Size(h)
	return buf
}

// decodeHeader parses an on-disk header, rejecting foreign or newer files.
func decodeHeader(buf []byte) (header, error) {
	var h header
	binary.Decode(buf, binary.LittleEndian, &h)
	if string(h.Magic[:]) != magic {
		return header{}, errors.New("not a filter file")
	}
	if h.Version != version {
		return header{}, fmt.Errorf("unsupported filter version %d", h.Version)
	}
	return h, nil
}

// sourceFingerprint returns the size and modification time that identify one
// version of the source database.
func sourceFingerprint(path string) (size, mtime uint64, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, fmt.Errorf("stat source %q: %w", path, err)
	}
	return uint64(info.Size()), uint64(info.ModTime().UnixNano()), nil
}

// Write persists the filter to path, recording sourcePath's fingerprint so a
// later [Open] can detect a stale filter.
// Everything is written to a temporary file beside path, which is then
// atomically renamed into place.
func (f *Filter) Write(path, sourcePath string) error {
	sourceSize, sourceMtime, err := sourceFingerprint(sourcePath)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer os.Remove(tmp) // no-op once renamed below; cleans up in case of an error

	hdr := header{
		NumBlocks:   f.NumBlocks,
		NumEntries:  f.NumEntries,
		SourceSize:  sourceSize,
		SourceMtime: sourceMtime,
		Probes:      uint32(f.probes),
	}
	// blocks is never resliced or reallocated, so reinterpreting it as bytes
	// for one sequential write is safe.
	data := unsafe.Slice((*byte)(unsafe.Pointer(&f.blocks[0])), len(f.blocks)*8)

	_, err = file.Write(hdr.encode())
	if err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Open memory-maps the filter at path read-only.
// It returns [ErrStale] if the filter was built from a different sourcePath
// than the current one.
func Open(path, sourcePath string) (_ *Filter, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			file.Close()
		}
	}()

	var buf [headerSize]byte
	if _, err = io.ReadFull(file, buf[:]); err != nil {
		return nil, fmt.Errorf("reading filter header: %w", err)
	}
	hdr, err := decodeHeader(buf[:])
	if err != nil {
		return nil, fmt.Errorf("%q: %w", path, err)
	}

	sourceSize, sourceMtime, err := sourceFingerprint(sourcePath)
	if err != nil {
		return nil, err
	}
	if sourceSize != hdr.SourceSize || sourceMtime != hdr.SourceMtime {
		return nil, ErrStale
	}

	dataLen := int(hdr.NumBlocks * bytesPerBlock)
	mapLen := headerSize + dataLen
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() != int64(mapLen) {
		return nil, fmt.Errorf("filter %q is truncated", path)
	}

	// Map from offset zero, which is always page-aligned, and index past the
	// header, so the mapping works whatever the system page size.
	data, err := syscall.Mmap(int(file.Fd()), 0, mapLen, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap filter: %w", err)
	}
	// Queries are uniformly random, so suppress read-ahead
	_ = syscall.Madvise(data, syscall.MADV_RANDOM)

	blocks := unsafe.Slice((*uint64)(unsafe.Pointer(&data[headerSize])), dataLen/8)
	return &Filter{
		blocks:     blocks,
		mask:       hdr.NumBlocks - 1,
		probes:     int(hdr.Probes),
		file:       file,
		NumEntries: hdr.NumEntries,
		NumBlocks:  hdr.NumBlocks,
	}, nil
}

// Close closes the backing file of a filter loaded with [Open].
func (f *Filter) Close() error {
	if f.file != nil {
		err := f.file.Close()
		f.file = nil
		return err
	}
	return nil
}
