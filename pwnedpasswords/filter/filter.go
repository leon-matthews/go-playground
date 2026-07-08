// Package filter implements a split-block Bloom filter over 20-byte SHA-1
// hashes, tuned for fast, lock-free, parallel membership queries.
//
// [SplitBlockBloom] is the in-memory data structure; [Filter] adds an on-disk
// format, persisted with [Filter.Write] and memory-mapped back with [Open].
package filter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

const (
	magic       = "PWNEDPASSWORDS"
	magicLength = len(magic)
	version     = 3    // bump on any change to block layout or probe placement
	headerSize  = 4096 // header bytes before the data; multiple of 8 keeps words aligned
)

// ErrStale reports that the filter was built from a different source database
// than the one presented, so it must be rebuilt.
var ErrStale = errors.New("filter is stale: source database has changed")

// Filter is a [SplitBlockBloom] with an on-disk representation.
// The zero value is not usable; build one in memory with [New] or load one
// memory-mapped with [Open].
type Filter struct {
	SplitBlockBloom
	file *os.File // backing file, non-nil when opened with Open
}

// New allocates an empty in-memory filter with numBlocks blocks, which must be
// a power of two, setting probes bits per element.
// Persist a built filter with [Filter.Write].
func New(numBlocks uint64, probes int) (*Filter, error) {
	bloom, err := newSplitBlockBloom(numBlocks, probes)
	if err != nil {
		return nil, err
	}
	return &Filter{SplitBlockBloom: bloom}, nil
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
		SplitBlockBloom: SplitBlockBloom{
			blocks:     blocks,
			mask:       hdr.NumBlocks - 1,
			probes:     int(hdr.Probes),
			NumEntries: hdr.NumEntries,
			NumBlocks:  hdr.NumBlocks,
		},
		file: file,
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
