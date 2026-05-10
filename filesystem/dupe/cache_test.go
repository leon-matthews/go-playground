package main

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeHash returns a 64-char hex digest filled with seed; distinct seeds differ.
func makeHash(seed byte) string {
	return strings.Repeat(fmt.Sprintf("%02x", seed), sha256.Size)
}

// newCache opens a Cache backed by a fresh DB inside a per-test temp dir and
// schedules its closure via t.Cleanup. Returns the path so individual subtests
// can reopen it to verify persistence.
func newCache(t *testing.T) (*Cache, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := openCache(path, nil)
	require.NoError(t, err)
	tuneForTest(c)
	t.Cleanup(func() { _ = c.Close() })
	return c, path
}

// tuneForTest drops MaxBatchDelay to keep isolated Set→Get tests fast.
func tuneForTest(c *Cache) {
	if c != nil && c.db != nil {
		c.db.MaxBatchDelay = time.Millisecond
	}
}

func TestPathInRoots(t *testing.T) {
	t.Run("path equal to root", func(t *testing.T) {
		assert.True(t, pathInRoots("/foo", []string{"/foo"}))
	})

	t.Run("path nested under root", func(t *testing.T) {
		assert.True(t, pathInRoots("/foo/bar", []string{"/foo"}))
		assert.True(t, pathInRoots("/foo/bar/baz/qux", []string{"/foo"}))
	})

	t.Run("path not under root", func(t *testing.T) {
		assert.False(t, pathInRoots("/baz", []string{"/foo"}))
	})

	t.Run("sibling with shared prefix is excluded", func(t *testing.T) {
		// The separator-suffix check exists for this case: /foo must not match
		// /foobar.
		assert.False(t, pathInRoots("/foobar", []string{"/foo"}))
		assert.False(t, pathInRoots("/foobar/baz", []string{"/foo"}))
	})

	t.Run("matches any of multiple roots", func(t *testing.T) {
		roots := []string{"/foo", "/bar"}
		assert.True(t, pathInRoots("/foo/x", roots))
		assert.True(t, pathInRoots("/bar/y", roots))
		assert.False(t, pathInRoots("/baz/z", roots))
	})

	t.Run("empty roots never match", func(t *testing.T) {
		assert.False(t, pathInRoots("/anything", nil))
		assert.False(t, pathInRoots("/anything", []string{}))
	})
}

func TestOpenCache(t *testing.T) {
	t.Run("missing file is created", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested", "deeper", "cache.db")
		c, err := openCache(path, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = c.Close() })

		assert.FileExists(t, path)
	})

	t.Run("empty path returns no-op cache", func(t *testing.T) {
		c, err := openCache("", nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = c.Close() })

		_, ok := c.Get("/anything")
		assert.False(t, ok)
		assert.NoError(t, c.Set("/anything", CacheEntry{Hash: "h"}))
		assert.NoError(t, c.Sweep(nil, nil))
	})

	t.Run("concurrent open times out", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cache.db")
		first, err := openCache(path, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = first.Close() })

		// Use a short timeout so the test doesn't sit idle for the production
		// default. The behaviour we care about is that opening a locked file
		// returns an error and a usable no-op cache, not the exact wait length.
		second, err := openCacheWithTimeout(path, 100*time.Millisecond, nil)
		t.Cleanup(func() { _ = second.Close() })

		assert.Error(t, err)
		assert.NotNil(t, second, "should return a usable no-op cache even on error")
		_, ok := second.Get("/anything")
		assert.False(t, ok)
	})
}

func TestCacheEntryBinary(t *testing.T) {
	t.Run("round-trips", func(t *testing.T) {
		want := CacheEntry{Size: 12345, ModTime: time.Unix(1700000000, 42).UTC(), Hash: makeHash(0xab)}
		blob, err := want.MarshalBinary()
		require.NoError(t, err)
		assert.Len(t, blob, entrySize)

		var got CacheEntry
		require.NoError(t, got.UnmarshalBinary(blob))
		assert.Equal(t, want.Size, got.Size)
		assert.Equal(t, want.Hash, got.Hash)
		assert.True(t, want.ModTime.Equal(got.ModTime))
	})

	t.Run("marshal rejects non-hex hash", func(t *testing.T) {
		_, err := CacheEntry{Hash: "not-hex"}.MarshalBinary()
		assert.Error(t, err)
	})

	t.Run("marshal rejects wrong-length hash", func(t *testing.T) {
		_, err := CacheEntry{Hash: "abcd"}.MarshalBinary()
		assert.Error(t, err)
	})

	t.Run("unmarshal rejects wrong-length input", func(t *testing.T) {
		var got CacheEntry
		assert.Error(t, got.UnmarshalBinary(make([]byte, entrySize-1)))
		assert.Error(t, got.UnmarshalBinary(make([]byte, entrySize+1)))
	})
}

func TestCacheSet(t *testing.T) {
	t1 := time.Unix(1, 0).UTC()
	t2 := time.Unix(2, 0).UTC()

	t.Run("set then get round-trips", func(t *testing.T) {
		c, _ := newCache(t)
		want := CacheEntry{Size: 100, ModTime: t1, Hash: makeHash(0x01)}

		require.NoError(t, c.Set("/foo/a", want))
		got, ok := c.Get("/foo/a")

		require.True(t, ok)
		assertEntryEqual(t, want, got)
	})

	t.Run("missing key returns ok=false", func(t *testing.T) {
		c, _ := newCache(t)
		_, ok := c.Get("/never-set")
		assert.False(t, ok)
	})

	t.Run("set overwrites previous value", func(t *testing.T) {
		c, _ := newCache(t)
		require.NoError(t, c.Set("/foo/a", CacheEntry{Size: 50, ModTime: t1, Hash: makeHash(0x01)}))
		require.NoError(t, c.Set("/foo/a", CacheEntry{Size: 100, ModTime: t2, Hash: makeHash(0x02)}))

		got, ok := c.Get("/foo/a")

		require.True(t, ok)
		assert.Equal(t, makeHash(0x02), got.Hash)
		assert.Equal(t, int64(100), got.Size)
	})

	t.Run("entries persist across reopen", func(t *testing.T) {
		c, path := newCache(t)
		want := CacheEntry{Size: 100, ModTime: t1, Hash: makeHash(0x03)}
		require.NoError(t, c.Set("/foo/a", want))
		require.NoError(t, c.Close())

		reopened, err := openCache(path, nil)
		require.NoError(t, err)
		tuneForTest(reopened)
		t.Cleanup(func() { _ = reopened.Close() })

		got, ok := reopened.Get("/foo/a")
		require.True(t, ok)
		assertEntryEqual(t, want, got)
	})
}

func TestCacheSweep(t *testing.T) {
	roots := []string{"/scope"}
	scoped := func(name string) string { return "/scope/" + name }
	external := func(name string) string { return "/elsewhere/" + name }

	t.Run("in-scope orphan is removed", func(t *testing.T) {
		c, _ := newCache(t)
		require.NoError(t, c.Set(scoped("present"), CacheEntry{Size: 1, Hash: makeHash(0x10)}))
		require.NoError(t, c.Set(scoped("orphan"), CacheEntry{Size: 2, Hash: makeHash(0x11)}))

		seen := map[string]struct{}{scoped("present"): {}}
		require.NoError(t, c.Sweep(seen, roots))

		_, ok := c.Get(scoped("present"))
		assert.True(t, ok, "seen entry should remain")
		_, ok = c.Get(scoped("orphan"))
		assert.False(t, ok, "in-scope unseen entry should be swept")
	})

	t.Run("out-of-scope orphan is preserved", func(t *testing.T) {
		c, _ := newCache(t)
		require.NoError(t, c.Set(scoped("present"), CacheEntry{Size: 1, Hash: makeHash(0x10)}))
		require.NoError(t, c.Set(external("elsewhere"), CacheEntry{Size: 99, Hash: makeHash(0x12)}))

		seen := map[string]struct{}{scoped("present"): {}}
		require.NoError(t, c.Sweep(seen, roots))

		_, ok := c.Get(external("elsewhere"))
		assert.True(t, ok, "out-of-scope entry should be preserved")
	})

	t.Run("seen-but-stat-failed path keeps its entry", func(t *testing.T) {
		// collectRoots saw the path so it's in `seen`, but processFile got a
		// stat error and never produced a fresh result. The entry must not be
		// swept just because no Set was made for it this scan.
		c, _ := newCache(t)
		want := CacheEntry{Size: 50, Hash: makeHash(0x20)}
		require.NoError(t, c.Set(scoped("transient"), want))

		seen := map[string]struct{}{scoped("transient"): {}}
		require.NoError(t, c.Sweep(seen, roots))

		got, ok := c.Get(scoped("transient"))
		require.True(t, ok)
		assert.Equal(t, want.Hash, got.Hash)
	})

	t.Run("sweep on empty cache is harmless", func(t *testing.T) {
		c, _ := newCache(t)
		require.NoError(t, c.Sweep(map[string]struct{}{}, roots))
	})
}

// assertEntryEqual compares CacheEntry values; ModTime uses Equal (location-agnostic).
func assertEntryEqual(t *testing.T, want, got CacheEntry) {
	t.Helper()
	assert.Equal(t, want.Size, got.Size)
	assert.Equal(t, want.Hash, got.Hash)
	assert.True(t, want.ModTime.Equal(got.ModTime), "ModTime mismatch: want %v got %v", want.ModTime, got.ModTime)
}
