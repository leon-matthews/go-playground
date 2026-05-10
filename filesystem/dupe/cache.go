package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// bucketName is versioned so schema changes can land without breaking old caches.
const bucketName = "hashes-v2"

// entrySize: 8B Size + 8B ModTime unix-nanos + 32B raw SHA-256 digest, big-endian.
const entrySize = 8 + 8 + sha256.Size

// openTimeout bounds bbolt.Open so a second concurrent dupe instance fails fast.
const openTimeout = 5 * time.Second

// maxBatchDelay caps the wait before a Set batch commits (bbolt default: 10ms).
const maxBatchDelay = 1 * time.Second

// maxBatchSize caps callbacks per batch (bbolt default: 1000).
const maxBatchSize = 10000

// CacheEntry is the persisted hash record for a single file, keyed by its
// absolute path. Size and ModTime are the verification fields: a cache hit
// requires both to match the current stat result, otherwise the hash is
// recomputed and the entry overwritten.
type CacheEntry struct {
	Size    int64
	ModTime time.Time
	Hash    string
}

// MarshalBinary packs CacheEntry into a fixed-width 48-byte big-endian record.
func (e CacheEntry) MarshalBinary() ([]byte, error) {
	digest, err := hex.DecodeString(e.Hash)
	if err != nil {
		return nil, fmt.Errorf("decode hash: %w", err)
	}
	if len(digest) != sha256.Size {
		return nil, fmt.Errorf("hash length %d, want %d", len(digest), sha256.Size)
	}
	buf := make([]byte, entrySize)
	binary.BigEndian.PutUint64(buf[0:8], uint64(e.Size))
	binary.BigEndian.PutUint64(buf[8:16], uint64(e.ModTime.UnixNano()))
	copy(buf[16:entrySize], digest)
	return buf, nil
}

// UnmarshalBinary parses a fixed-width 48-byte record produced by MarshalBinary.
func (e *CacheEntry) UnmarshalBinary(data []byte) error {
	if len(data) != entrySize {
		return fmt.Errorf("entry length %d, want %d", len(data), entrySize)
	}
	e.Size = int64(binary.BigEndian.Uint64(data[0:8]))
	e.ModTime = time.Unix(0, int64(binary.BigEndian.Uint64(data[8:16]))).UTC()
	e.Hash = hex.EncodeToString(data[16:entrySize])
	return nil
}

// Cache is a persistent hash cache backed by bbolt. Each Set is durable when
// the next batch transaction commits, so an interrupted scan only loses work
// for files that were mid-hash at the moment of interruption.
//
// A Cache with a nil db is a no-op: Get always misses, Set/Sweep are silent
// no-ops, Close returns nil. openCache returns one of these when the on-disk
// database can't be opened, so callers don't need to special-case "cache
// disabled" — they just keep calling the same methods.
type Cache struct {
	db  *bolt.DB
	log *slog.Logger
}

// cachePath returns the absolute path of the persistent hash cache.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dupe", "cache.db"), nil
}

// openCache opens or creates the bbolt database at path and ensures the
// hash bucket exists. On any error (including timeout from a concurrent
// instance holding the lock) it returns a no-op Cache plus the error, so
// the caller can log and continue without caching rather than abort.
func openCache(path string, log *slog.Logger) (*Cache, error) {
	return openCacheWithTimeout(path, openTimeout, log)
}

// openCacheWithTimeout is openCache with an explicit lock-wait timeout.
func openCacheWithTimeout(path string, timeout time.Duration, log *slog.Logger) (*Cache, error) {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	if path == "" {
		return &Cache{log: log}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &Cache{log: log}, err
	}
	db, err := bolt.Open(path, 0o644, &bolt.Options{Timeout: timeout})
	if err != nil {
		return &Cache{log: log}, err
	}
	db.MaxBatchDelay = maxBatchDelay
	db.MaxBatchSize = maxBatchSize
	err = db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		return e
	})
	if err != nil {
		db.Close()
		return &Cache{log: log}, err
	}
	log.Debug("cache opened", "path", path)
	return &Cache{db: db, log: log}, nil
}

// Close releases the underlying database file. Safe to call on a no-op Cache.
func (c *Cache) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}

// Get returns the cache entry for path, if present.
func (c *Cache) Get(path string) (CacheEntry, bool) {
	if c == nil || c.db == nil {
		return CacheEntry{}, false
	}
	var entry CacheEntry
	found := false
	_ = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}
		raw := b.Get([]byte(path))
		if raw == nil {
			return nil
		}
		if err := entry.UnmarshalBinary(raw); err != nil {
			c.log.Warn("cache: failed to decode entry; treating as miss", "path", path, "err", err)
			return nil
		}
		found = true
		return nil
	})
	return entry, found
}

// Set persists entry under path; Hash must be hex of a 32-byte SHA-256 digest.
func (c *Cache) Set(path string, entry CacheEntry) error {
	if c == nil || c.db == nil {
		return nil
	}
	blob, err := entry.MarshalBinary()
	if err != nil {
		return err
	}
	return c.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.New("cache bucket missing")
		}
		return b.Put([]byte(path), blob)
	})
}

// Sweep removes entries whose key is under any of roots but not in seen.
// Out-of-scope entries (under no current root) are left untouched so that
// scanning one tree today doesn't invalidate cached entries from a different
// tree scanned yesterday.
func (c *Cache) Sweep(seen map[string]struct{}, roots []string) error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}
		var toDelete [][]byte
		_ = b.ForEach(func(k, _ []byte) error {
			path := string(k)
			if _, ok := seen[path]; ok {
				return nil
			}
			if pathInRoots(path, roots) {
				toDelete = append(toDelete, append([]byte(nil), k...))
			}
			return nil
		})
		for _, k := range toDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// pathInRoots reports whether path is equal to, or nested under, any of roots.
// Roots must already be absolute and cleaned; the separator boundary check
// prevents "/foo" from matching "/foobar".
func pathInRoots(path string, roots []string) bool {
	sep := string(filepath.Separator)
	for _, r := range roots {
		if path == r || strings.HasPrefix(path, r+sep) {
			return true
		}
	}
	return false
}
