package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestUpdateCache(t *testing.T) {
	roots := []string{"/scope"}
	scoped := func(name string) string { return "/scope/" + name }
	external := func(name string) string { return "/elsewhere/" + name }
	t1 := time.Unix(1, 0).UTC()
	t2 := time.Unix(2, 0).UTC()

	t.Run("new file is added", func(t *testing.T) {
		cache := map[string]CacheEntry{}
		files := []FileInfo{{Path: scoped("a"), Size: 100, ModTime: t1, Hash: "abc"}}
		paths := []string{scoped("a")}

		updateCache(cache, files, paths, roots)

		assert.Equal(t, CacheEntry{Size: 100, ModTime: t1, Hash: "abc"}, cache[scoped("a")])
	})

	t.Run("existing entry is overwritten", func(t *testing.T) {
		cache := map[string]CacheEntry{
			scoped("a"): {Size: 50, ModTime: t1, Hash: "old"},
		}
		files := []FileInfo{{Path: scoped("a"), Size: 100, ModTime: t2, Hash: "new"}}
		paths := []string{scoped("a")}

		updateCache(cache, files, paths, roots)

		assert.Equal(t, "new", cache[scoped("a")].Hash)
		assert.Equal(t, int64(100), cache[scoped("a")].Size)
	})

	t.Run("empty hash does not overwrite existing entry", func(t *testing.T) {
		// A hash-failed FileInfo arrives with Path populated but Hash empty.
		// The prior cached hash is the only thing we have, so keep it.
		old := CacheEntry{Size: 50, ModTime: t1, Hash: "old"}
		cache := map[string]CacheEntry{scoped("a"): old}
		files := []FileInfo{{Path: scoped("a"), Size: 100, ModTime: t2, Hash: ""}}
		paths := []string{scoped("a")}

		updateCache(cache, files, paths, roots)

		assert.Equal(t, old, cache[scoped("a")])
	})

	t.Run("in-scope orphan is swept", func(t *testing.T) {
		cache := map[string]CacheEntry{
			scoped("present"): {Size: 1, Hash: "p"},
			scoped("orphan"):  {Size: 2, Hash: "o"},
		}
		files := []FileInfo{{Path: scoped("present"), Size: 1, Hash: "p"}}
		paths := []string{scoped("present")}

		updateCache(cache, files, paths, roots)

		assert.Contains(t, cache, scoped("present"))
		assert.NotContains(t, cache, scoped("orphan"))
	})

	t.Run("out-of-scope orphan is preserved", func(t *testing.T) {
		cache := map[string]CacheEntry{
			scoped("present"):     {Size: 1, Hash: "p"},
			external("elsewhere"): {Size: 99, Hash: "e"},
		}
		files := []FileInfo{{Path: scoped("present"), Size: 1, Hash: "p"}}
		paths := []string{scoped("present")}

		updateCache(cache, files, paths, roots)

		assert.Contains(t, cache, external("elsewhere"))
	})

	t.Run("stat-failed path keeps its cache entry", func(t *testing.T) {
		// collectRoots saw the path (so it's in paths), but processFile got a
		// stat error and the file never reached results. The entry is in
		// `seen` and must not be swept.
		old := CacheEntry{Size: 50, Hash: "keep"}
		cache := map[string]CacheEntry{scoped("transient"): old}
		var files []FileInfo
		paths := []string{scoped("transient")}

		updateCache(cache, files, paths, roots)

		assert.Equal(t, old, cache[scoped("transient")])
	})
}

func TestLoadCache(t *testing.T) {
	t.Run("missing file returns empty cache", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nonexistent.gob")

		cache := loadCache(path)

		assert.Empty(t, cache)
	})

	t.Run("corrupt file returns empty cache", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "corrupt.gob")
		require.NoError(t, os.WriteFile(path, []byte("not a valid gob stream"), 0o644))

		cache := loadCache(path)

		assert.Empty(t, cache)
	})

	t.Run("round-trips data written by saveCache", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cache.gob")
		original := map[string]CacheEntry{
			"/foo/a": {Size: 100, ModTime: time.Unix(1, 0).UTC(), Hash: "abc"},
			"/foo/b": {Size: 200, ModTime: time.Unix(2, 0).UTC(), Hash: "def"},
		}
		require.NoError(t, saveCache(path, original))

		loaded := loadCache(path)

		// Compare field-wise: time.Time after gob round-trip can fail
		// reflect.DeepEqual due to internal location-pointer differences even
		// when .Equal() is true.
		require.Len(t, loaded, len(original))
		for k, want := range original {
			got, ok := loaded[k]
			require.True(t, ok, "missing key %q", k)
			assert.Equal(t, want.Size, got.Size)
			assert.Equal(t, want.Hash, got.Hash)
			assert.True(t, want.ModTime.Equal(got.ModTime), "ModTime mismatch for %q", k)
		}
	})
}

func TestSaveCache(t *testing.T) {
	t.Run("creates missing parent directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested", "deeper", "cache.gob")

		err := saveCache(path, map[string]CacheEntry{"x": {Size: 1, Hash: "h"}})

		require.NoError(t, err)
		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("second save overwrites the first", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cache.gob")
		require.NoError(t, saveCache(path, map[string]CacheEntry{"a": {Hash: "old"}}))
		require.NoError(t, saveCache(path, map[string]CacheEntry{"a": {Hash: "new"}}))

		loaded := loadCache(path)

		assert.Equal(t, "new", loaded["a"].Hash)
	})

	t.Run("empty map round-trips", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cache.gob")

		require.NoError(t, saveCache(path, map[string]CacheEntry{}))
		loaded := loadCache(path)

		assert.Empty(t, loaded)
	})

	t.Run("does not leave temp files behind on success", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cache.gob")

		require.NoError(t, saveCache(path, map[string]CacheEntry{"x": {Hash: "h"}}))

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "only the cache file should remain")
		assert.Equal(t, "cache.gob", entries[0].Name())
	})
}
