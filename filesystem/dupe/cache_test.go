package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeHash returns a 32-byte digest filled with seed; distinct seeds differ.
func makeHash(seed byte) [32]byte {
	var h [32]byte
	for i := range h {
		h[i] = seed
	}
	return h
}

// newCache opens a Cache backed by a fresh DB inside a per-test temp dir.
func newCache(t *testing.T) (*Cache, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := openCache(path, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c, path
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
		// The separator-suffix check exists for this case: /foo must not match /foobar.
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
		assert.NoError(t, c.Set("/anything", CacheEntry{}))
		assert.NoError(t, c.Sweep(nil, nil))
	})

	t.Run("non-sqlite file is replaced", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cache.db")
		require.NoError(t, os.WriteFile(path, []byte("not-a-sqlite-file\n"), 0o644))

		c, err := openCache(path, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = c.Close() })

		header, err := os.ReadFile(path)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(header), len(sqliteMagic))
		assert.Equal(t, sqliteMagic, string(header[:len(sqliteMagic)]))

		want := CacheEntry{Size: 1, ModTime: time.Unix(1, 0).UTC(), Hash: makeHash(0x42)}
		require.NoError(t, c.Set("/a", want))
		require.NoError(t, c.Flush())
		_, ok := c.Get("/a")
		assert.True(t, ok, "recreated cache should be usable")
	})
}

func TestCacheSet(t *testing.T) {
	t1 := time.Unix(1, 0).UTC()
	t2 := time.Unix(2, 0).UTC()

	t.Run("set then get round-trips", func(t *testing.T) {
		c, _ := newCache(t)
		want := CacheEntry{Size: 100, ModTime: t1, Hash: makeHash(0x01)}

		require.NoError(t, c.Set("/foo/a", want))
		require.NoError(t, c.Flush())
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
		require.NoError(t, c.Flush())

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
		// collectRoots saw the path so it's in `seen`, but processFile got a stat
		// error and never produced a fresh result. The entry must not be swept.
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
