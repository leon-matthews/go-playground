package mimicry

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

// sweepInputs synthesises the FolderInfo values Sweep expects from a flat list of file paths.
func sweepInputs(paths []string) []FolderInfo {
	byFolder := make(map[string][]string)
	for _, p := range paths {
		folder, name := splitFolderPath(p)
		byFolder[folder] = append(byFolder[folder], name)
	}
	folders := make([]FolderInfo, 0, len(byFolder))
	for folder, names := range byFolder {
		folders = append(folders, FolderInfo{Path: folder, Children: names})
	}
	return folders
}

// newCache opens a Cache backed by a fresh DB inside a per-test temp dir.
func newCache(t *testing.T) (*Cache, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := OpenCache(path, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c, path
}

func TestOpenCache(t *testing.T) {
	t.Run("missing file is created", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested", "deeper", "cache.db")
		c, err := OpenCache(path, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = c.Close() })

		assert.FileExists(t, path)
	})

	t.Run("empty path returns no-op cache", func(t *testing.T) {
		c, err := OpenCache("", nil)
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

		c, err := OpenCache(path, nil)
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

		reopened, err := OpenCache(path, nil)
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
		require.NoError(t, c.Flush())

		require.NoError(t, c.Sweep(sweepInputs([]string{scoped("present")}), roots))

		_, ok := c.Get(scoped("present"))
		assert.True(t, ok, "seen entry should remain")
		_, ok = c.Get(scoped("orphan"))
		assert.False(t, ok, "in-scope unseen entry should be swept")
	})

	t.Run("out-of-scope orphan is preserved", func(t *testing.T) {
		c, _ := newCache(t)
		require.NoError(t, c.Set(scoped("present"), CacheEntry{Size: 1, Hash: makeHash(0x10)}))
		require.NoError(t, c.Set(external("elsewhere"), CacheEntry{Size: 99, Hash: makeHash(0x12)}))
		require.NoError(t, c.Flush())

		require.NoError(t, c.Sweep(sweepInputs([]string{scoped("present")}), roots))

		_, ok := c.Get(external("elsewhere"))
		assert.True(t, ok, "out-of-scope entry should be preserved")
	})

	t.Run("seen-but-stat-failed path keeps its entry", func(t *testing.T) {
		// collectRoots saw the path so it's in the seen set, but processFile got a stat
		// error and never produced a fresh result. The entry must not be swept.
		c, _ := newCache(t)
		want := CacheEntry{Size: 50, Hash: makeHash(0x20)}
		require.NoError(t, c.Set(scoped("transient"), want))
		require.NoError(t, c.Flush())

		require.NoError(t, c.Sweep(sweepInputs([]string{scoped("transient")}), roots))

		got, ok := c.Get(scoped("transient"))
		require.True(t, ok)
		assert.Equal(t, want.Hash, got.Hash)
	})

	t.Run("sweep on empty cache is harmless", func(t *testing.T) {
		c, _ := newCache(t)
		require.NoError(t, c.Sweep(nil, roots))
	})
}

// assertEntryEqual compares CacheEntry values; ModTime uses Equal (location-agnostic).
func assertEntryEqual(t *testing.T, want, got CacheEntry) {
	t.Helper()
	assert.Equal(t, want.Size, got.Size)
	assert.Equal(t, want.Hash, got.Hash)
	assert.True(t, want.ModTime.Equal(got.ModTime), "ModTime mismatch: want %v got %v", want.ModTime, got.ModTime)
}

func TestCacheFolderMtime(t *testing.T) {
	t.Run("set then get round-trips", func(t *testing.T) {
		c, _ := newCache(t)
		mtime := time.Unix(1700000000, 42).UTC()

		require.NoError(t, c.SetFolderMtime("/foo", mtime))
		require.NoError(t, c.Flush())

		got, ok := c.GetFolderMtime("/foo")
		require.True(t, ok)
		assert.True(t, mtime.Equal(got))
	})

	t.Run("missing folder returns ok=false", func(t *testing.T) {
		c, _ := newCache(t)
		_, ok := c.GetFolderMtime("/never-set")
		assert.False(t, ok)
	})

	t.Run("set updates existing mtime", func(t *testing.T) {
		c, _ := newCache(t)
		t1 := time.Unix(1, 0).UTC()
		t2 := time.Unix(2, 0).UTC()

		require.NoError(t, c.SetFolderMtime("/foo", t1))
		require.NoError(t, c.SetFolderMtime("/foo", t2))
		require.NoError(t, c.Flush())

		got, ok := c.GetFolderMtime("/foo")
		require.True(t, ok)
		assert.True(t, t2.Equal(got))
	})
}

func TestCacheGetFilesInFolder(t *testing.T) {
	c, _ := newCache(t)
	t1 := time.Unix(1, 0).UTC()

	require.NoError(t, c.Set("/scope/a.txt", CacheEntry{Size: 1, ModTime: t1, Hash: makeHash(0x01)}))
	require.NoError(t, c.Set("/scope/b.txt", CacheEntry{Size: 2, ModTime: t1, Hash: makeHash(0x02)}))
	require.NoError(t, c.Set("/scope/sub/c.txt", CacheEntry{Size: 3, ModTime: t1, Hash: makeHash(0x03)}))
	require.NoError(t, c.Set("/other/d.txt", CacheEntry{Size: 4, ModTime: t1, Hash: makeHash(0x04)}))
	require.NoError(t, c.Flush())

	got, err := c.GetFilesInFolder("/scope")
	require.NoError(t, err)
	assert.Len(t, got, 2, "should return only direct children of /scope, not /scope/sub")
	assert.Equal(t, int64(1), got["a.txt"].Size)
	assert.Equal(t, int64(2), got["b.txt"].Size)

	got, err = c.GetFilesInFolder("/scope/sub")
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, int64(3), got["c.txt"].Size)

	got, err = c.GetFilesInFolder("/nowhere")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCacheAllFiles(t *testing.T) {
	t.Run("returns every cached file as FileInfo", func(t *testing.T) {
		c, _ := newCache(t)
		t1 := time.Unix(1, 0).UTC()
		require.NoError(t, c.Set("/scope/a.txt", CacheEntry{Size: 1, ModTime: t1, Hash: makeHash(0x01)}))
		require.NoError(t, c.Set("/scope/sub/c.md", CacheEntry{Size: 3, ModTime: t1, Hash: makeHash(0x03)}))
		require.NoError(t, c.Set("/other/noext", CacheEntry{Size: 4, ModTime: t1, Hash: makeHash(0x04)}))
		require.NoError(t, c.Flush())

		files, err := c.AllFiles()
		require.NoError(t, err)
		require.Len(t, files, 3)

		byPath := make(map[string]FileInfo, len(files))
		for _, f := range files {
			byPath[f.Path] = f
		}
		require.Contains(t, byPath, "/scope/a.txt")
		assert.Equal(t, int64(1), byPath["/scope/a.txt"].Size)
		assert.Equal(t, ".txt", byPath["/scope/a.txt"].Extension)
		assert.Equal(t, makeHash(0x01), byPath["/scope/a.txt"].Hash)
		assert.Equal(t, ".md", byPath["/scope/sub/c.md"].Extension)
		assert.Empty(t, byPath["/other/noext"].Extension)
	})

	t.Run("empty cache returns no files", func(t *testing.T) {
		c, _ := newCache(t)
		files, err := c.AllFiles()
		require.NoError(t, err)
		assert.Empty(t, files)
	})
}

func TestCacheMigrationFromV1(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")

	// Build a v1-shape cache: hashes table + user_version=1.
	c, err := OpenCache(path, nil)
	require.NoError(t, err)
	require.NoError(t, c.Close())
	// Manually fake a v1 schema by stamping the file via the sqlite driver-less path:
	// we open through our own code which always migrates, so instead we drop the v2
	// tables and create a v1 hashes table, then reset user_version. This exercises the
	// "had != 0 → warn + rebuild" branch on the next OpenCache.
	scratch, err := OpenCache(path, nil)
	require.NoError(t, err)
	_, err = scratch.db.Exec("DROP TABLE files")
	require.NoError(t, err)
	_, err = scratch.db.Exec("DROP TABLE folders")
	require.NoError(t, err)
	_, err = scratch.db.Exec(`CREATE TABLE hashes (path TEXT PRIMARY KEY, size INTEGER, modtime INTEGER, hash BLOB)`)
	require.NoError(t, err)
	_, err = scratch.db.Exec("PRAGMA user_version = 1")
	require.NoError(t, err)
	require.NoError(t, scratch.Close())

	// Reopen with current code - should detect mismatch and rebuild.
	reopened, err := OpenCache(path, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = reopened.Close() })

	var version int
	require.NoError(t, reopened.db.QueryRow("PRAGMA user_version").Scan(&version))
	assert.Equal(t, schemaVersion, version)

	// Tables should be folders + files, hashes gone.
	rows, err := reopened.db.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	require.NoError(t, err)
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	assert.ElementsMatch(t, []string{"folders", "files"}, tables)
}
