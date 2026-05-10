package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// bucketName is versioned so a future schema change can introduce a new bucket
// (e.g. "hashes-v2") and migrate or drop the old one without breaking older
// caches in the field.
const bucketName = "hashes-v1"

// openTimeout bounds bbolt.Open so a second concurrent dupe instance fails
// fast with a clear error instead of blocking on the file lock indefinitely.
const openTimeout = 5 * time.Second

// CacheEntry is the persisted hash record for a single file, keyed by its
// absolute path. Size and ModTime are the verification fields: a cache hit
// requires both to match the current stat result, otherwise the hash is
// recomputed and the entry overwritten.
type CacheEntry struct {
	Size    int64
	ModTime time.Time
	Hash    string
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
	db *bolt.DB
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
func openCache(path string) (*Cache, error) {
	return openCacheWithTimeout(path, openTimeout)
}

// openCacheWithTimeout is openCache with an explicit lock-wait timeout. Split
// out so tests can use a short timeout for the concurrent-open case without
// affecting the production default.
func openCacheWithTimeout(path string, timeout time.Duration) (*Cache, error) {
	if path == "" {
		return &Cache{}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &Cache{}, err
	}
	db, err := bolt.Open(path, 0o644, &bolt.Options{Timeout: timeout})
	if err != nil {
		return &Cache{}, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		return e
	})
	if err != nil {
		db.Close()
		return &Cache{}, err
	}
	slog.Debug("cache opened", "path", path)
	return &Cache{db: db}, nil
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
		if err := gob.NewDecoder(bytes.NewReader(raw)).Decode(&entry); err != nil {
			slog.Warn("cache: failed to decode entry; treating as miss", "path", path, "err", err)
			return nil
		}
		found = true
		return nil
	})
	return entry, found
}

// Set persists entry under path. Concurrent calls from worker goroutines
// coalesce into shared bbolt transactions via Batch (one fsync per batch).
func (c *Cache) Set(path string, entry CacheEntry) error {
	if c == nil || c.db == nil {
		return nil
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(entry); err != nil {
		return fmt.Errorf("encode cache entry: %w", err)
	}
	return c.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.New("cache bucket missing")
		}
		return b.Put([]byte(path), buf.Bytes())
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
