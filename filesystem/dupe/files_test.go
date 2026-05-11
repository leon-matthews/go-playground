package main

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustHash decodes a 64-char hex string into a [32]byte; for canonical test fixtures.
func mustHash(t *testing.T, s string) [32]byte {
	t.Helper()
	raw, err := hex.DecodeString(s)
	require.NoError(t, err)
	require.Len(t, raw, 32)
	var h [32]byte
	copy(h[:], raw)
	return h
}

// allPaths flattens FolderScan children into full absolute paths.
func allPaths(folders []FolderScan) []string {
	paths := make([]string, 0)
	for _, fs := range folders {
		for _, name := range fs.Children {
			paths = append(paths, filepath.Join(fs.Path, name))
		}
	}
	return paths
}

func TestCollectRoots(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		root := t.TempDir()

		folders, err := newCollector(nil).collectRoots(root)

		require.NoError(t, err)
		assert.Empty(t, allPaths(folders))
	})

	t.Run("flat directory of files", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root, "a.txt", "b.txt", "c.log")

		folders, err := newCollector(nil).collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, allPaths(folders))
	})

	t.Run("nested directories", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root,
			"top.txt",
			"sub/one.txt",
			"sub/two.txt",
			"sub/deep/three.txt",
		)

		folders, err := newCollector(nil).collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, allPaths(folders))
	})

	t.Run("directories are excluded from results", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "empty-dir"), 0o755))
		want := writeFiles(t, root, "only.txt")

		folders, err := newCollector(nil).collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, allPaths(folders))
	})

	t.Run("missing root returns error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "does-not-exist")

		folders, err := newCollector(nil).collectRoots(missing)

		assert.Error(t, err)
		assert.Empty(t, allPaths(folders))
	})

	t.Run("unreadable subdirectory is skipped", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("permission semantics differ on Windows")
		}
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}

		root := t.TempDir()
		readable := writeFiles(t, root, "readable.txt", "locked/hidden.txt")

		locked := filepath.Join(root, "locked")
		require.NoError(t, os.Chmod(locked, 0o000))
		t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

		folders, err := newCollector(nil).collectRoots(root)

		require.NoError(t, err)
		assert.Contains(t, allPaths(folders), readable[0], "readable file should still be returned")
	})

	t.Run("multiple roots are combined", func(t *testing.T) {
		rootA := t.TempDir()
		rootB := t.TempDir()
		wantA := writeFiles(t, rootA, "a.txt", "sub/b.txt")
		wantB := writeFiles(t, rootB, "c.txt")

		folders, err := newCollector(nil).collectRoots(rootA, rootB)

		require.NoError(t, err)
		assert.ElementsMatch(t, append(wantA, wantB...), allPaths(folders))
	})

	t.Run("overlapping roots produce no duplicates", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root, "top.txt", "sub/inner.txt")

		folders, err := newCollector(nil).collectRoots(root, filepath.Join(root, "sub"))

		require.NoError(t, err)
		assert.ElementsMatch(t, want, allPaths(folders))
	})

	t.Run("missing root among others returns error", func(t *testing.T) {
		good := t.TempDir()
		writeFiles(t, good, "a.txt")
		missing := filepath.Join(t.TempDir(), "does-not-exist")

		_, err := newCollector(nil).collectRoots(good, missing)

		assert.Error(t, err)
	})

	t.Run("symlinks are excluded", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink creation requires privileges on Windows")
		}

		root := t.TempDir()
		want := writeFiles(t, root, "real.txt", "sub/target.txt")

		fileLink := filepath.Join(root, "file-link")
		require.NoError(t, os.Symlink(want[0], fileLink))

		dirLink := filepath.Join(root, "dir-link")
		require.NoError(t, os.Symlink(filepath.Join(root, "sub"), dirLink))

		folders, err := newCollector(nil).collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, allPaths(folders))
	})

	t.Run("file root is logged and ignored", func(t *testing.T) {
		root := t.TempDir()
		filePath := writeFiles(t, root, "lonely.txt")[0]

		folders, err := newCollector(nil).collectRoots(filePath)

		require.NoError(t, err)
		assert.Empty(t, folders)
	})

	t.Run("folder mtime is captured", func(t *testing.T) {
		root := t.TempDir()
		writeFiles(t, root, "a.txt")

		folders, err := newCollector(nil).collectRoots(root)
		require.NoError(t, err)
		require.NotEmpty(t, folders)

		info, err := os.Stat(folders[0].Path)
		require.NoError(t, err)
		assert.True(t, info.ModTime().Equal(folders[0].Mtime), "FolderScan.Mtime should match a fresh stat")
	})
}

func TestCollector(t *testing.T) {
	t.Run("happy path populates fields", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root, "a.txt", "sub/b.txt", "sub/c.txt")

		c := newCollector(nil)
		require.NoError(t, c.Walk(root))

		assert.Equal(t, 3, c.TotalFiles())
		assert.ElementsMatch(t, want, allPaths(c.Folders))
		assert.Equal(t, []string{root}, c.AbsRoots)
	})

	t.Run("non-directory root is recorded in AbsRoots but contributes no folders", func(t *testing.T) {
		root := t.TempDir()
		filePath := writeFiles(t, root, "lonely.txt")[0]

		c := newCollector(nil)
		require.NoError(t, c.Walk(filePath))

		assert.Empty(t, c.Folders)
		assert.Equal(t, 0, c.TotalFiles())
		assert.Equal(t, []string{filePath}, c.AbsRoots, "abs-root list still includes a file root for sweep scoping")
	})

	t.Run("missing root errors before populating", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "does-not-exist")

		c := newCollector(nil)
		err := c.Walk(missing)

		assert.Error(t, err)
		assert.Empty(t, c.Folders)
		assert.Empty(t, c.AbsRoots)
	})
}

func TestHashFile(t *testing.T) {
	// SHA-256 of well-known inputs, used to verify hashFile produces the
	// canonical digest rather than something self-consistent but wrong.
	const (
		sha256Empty = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		sha256Hello = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	)

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty")
		require.NoError(t, os.WriteFile(path, nil, 0o644))

		got, err := hashFile(path)

		require.NoError(t, err)
		assert.Equal(t, mustHash(t, sha256Empty), got)
	})

	t.Run("known content matches canonical SHA256", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "hello")
		require.NoError(t, os.WriteFile(path, []byte("hello"), 0o644))

		got, err := hashFile(path)

		require.NoError(t, err)
		assert.Equal(t, mustHash(t, sha256Hello), got)
	})

	t.Run("identical content produces identical hashes", func(t *testing.T) {
		dir := t.TempDir()
		a := filepath.Join(dir, "a")
		b := filepath.Join(dir, "b")
		require.NoError(t, os.WriteFile(a, []byte("same bytes"), 0o644))
		require.NoError(t, os.WriteFile(b, []byte("same bytes"), 0o644))

		hashA, err := hashFile(a)
		require.NoError(t, err)
		hashB, err := hashFile(b)
		require.NoError(t, err)

		assert.Equal(t, hashA, hashB)
	})

	t.Run("different content produces different hashes", func(t *testing.T) {
		dir := t.TempDir()
		a := filepath.Join(dir, "a")
		b := filepath.Join(dir, "b")
		require.NoError(t, os.WriteFile(a, []byte("alpha"), 0o644))
		require.NoError(t, os.WriteFile(b, []byte("beta"), 0o644))

		hashA, err := hashFile(a)
		require.NoError(t, err)
		hashB, err := hashFile(b)
		require.NoError(t, err)

		assert.NotEqual(t, hashA, hashB)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "does-not-exist")

		got, err := hashFile(missing)

		assert.Error(t, err)
		assert.Equal(t, [32]byte{}, got)
	})

	t.Run("directory path returns error", func(t *testing.T) {
		got, err := hashFile(t.TempDir())

		assert.Error(t, err)
		assert.Equal(t, [32]byte{}, got)
	})
}

// writeFiles creates each path (relative to root) as an empty file, making any
// necessary parent directories. Returns the absolute paths created.
func writeFiles(t *testing.T, root string, rels ...string) []string {
	t.Helper()
	abs := make([]string, 0, len(rels))
	for _, rel := range rels {
		full := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		f, err := os.Create(full)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		abs = append(abs, full)
	}
	return abs
}

// writeFile writes contents to root/name and returns the absolute path.
func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	full := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(full, []byte(contents), 0o644))
	return full
}

func TestScannerTrustedPath(t *testing.T) {
	root := t.TempDir()
	pathA := writeFile(t, root, "a.txt", "real content")

	c, _ := newCache(t)
	scans, err := newCollector(nil).collectRoots(root)
	require.NoError(t, err)
	require.Len(t, scans, 1)

	// Plant a deliberately wrong hash for a.txt; trusted-folder path should serve it.
	wrong := makeHash(0xff)
	require.NoError(t, c.Set(pathA, CacheEntry{Size: int64(len("real content")), ModTime: scans[0].Mtime, Hash: wrong}))
	require.NoError(t, c.SetFolderMtime(scans[0].Path, scans[0].Mtime))
	require.NoError(t, c.Flush())

	scanner := newScanner(c, 4, nil, false)
	files := scanner.Process(scans)
	require.Len(t, files, 1)
	// The verifier won't fire because this isn't a duplicate group (singleton).
	assert.Equal(t, wrong, files[0].Hash, "trusted folder should return cached hash unchanged")
}

func TestScannerStalePath(t *testing.T) {
	root := t.TempDir()
	pathA := writeFile(t, root, "a.txt", "real content")

	c, _ := newCache(t)
	scans, err := newCollector(nil).collectRoots(root)
	require.NoError(t, err)

	// Cache the wrong hash and a wrong folder mtime → folder is stale → per-file path runs.
	wrong := makeHash(0xff)
	require.NoError(t, c.Set(pathA, CacheEntry{Size: int64(len("real content")), ModTime: scans[0].Mtime, Hash: wrong}))
	require.NoError(t, c.SetFolderMtime(scans[0].Path, time.Unix(1, 0)))
	require.NoError(t, c.Flush())

	scanner := newScanner(c, 4, nil, false)
	files := scanner.Process(scans)
	require.Len(t, files, 1)
	// Cache file entry matched on size+mtime so the file-level Layer-1 cache served the
	// (wrong) hash. The folder being stale only means we stat'd it — it doesn't force
	// re-hashing of a file whose size+mtime still matches.
	assert.Equal(t, wrong, files[0].Hash)
}

func TestScannerForceFlag(t *testing.T) {
	root := t.TempDir()
	pathA := writeFile(t, root, "a.txt", "real content")

	c, _ := newCache(t)
	scans, err := newCollector(nil).collectRoots(root)
	require.NoError(t, err)

	// Cache wrong hash + matching folder mtime → trusted path would serve the wrong hash.
	// With force=true, the trusted branch is bypassed and we stat each file. The file's
	// size+mtime still match cache so the file-level cache STILL serves the wrong hash —
	// force only disables the folder-mtime layer, not the file-level cache.
	wrong := makeHash(0xff)
	require.NoError(t, c.Set(pathA, CacheEntry{Size: int64(len("real content")), ModTime: scans[0].Mtime, Hash: wrong}))
	require.NoError(t, c.SetFolderMtime(scans[0].Path, scans[0].Mtime))
	require.NoError(t, c.Flush())

	scanner := newScanner(c, 4, nil, true)
	files := scanner.Process(scans)
	require.Len(t, files, 1)
	// Same as the stale-path test — force makes us re-stat but file-level cache still applies.
	assert.Equal(t, wrong, files[0].Hash)
}

func TestScannerVerifyCatchesInPlaceEdit(t *testing.T) {
	root := t.TempDir()
	pathA := writeFile(t, root, "a.txt", "matching content")
	pathB := writeFile(t, root, "b.txt", "matching content")

	c, _ := newCache(t)
	scans, err := newCollector(nil).collectRoots(root)
	require.NoError(t, err)
	require.Len(t, scans, 1)

	scanner := newScanner(c, 4, nil, false)

	// First scan: populates cache + folder mtime.
	firstFiles := scanner.Process(scans)
	require.NoError(t, c.Flush())
	require.Len(t, firstFiles, 2)
	assert.Equal(t, firstFiles[0].Hash, firstFiles[1].Hash, "both should hash identically")

	// Edit a.txt in place; preserve folder mtime by re-stat'ing the folder afterwards.
	folderBefore, err := os.Stat(scans[0].Path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(pathA, []byte("different content!"), 0o644))
	// Linux: writing to an existing file doesn't bump the folder mtime. Sanity-check that.
	folderAfter, err := os.Stat(scans[0].Path)
	require.NoError(t, err)
	require.True(t, folderBefore.ModTime().Equal(folderAfter.ModTime()),
		"in-place write must not change folder mtime for this test to be meaningful")

	// Second scan: folder is trusted, both files served from cache with matching (old) hash —
	// but the verification pass should re-stat them, see a.txt's size changed, re-hash, and
	// emit divergent hashes.
	scans2, err := newCollector(nil).collectRoots(root)
	require.NoError(t, err)
	secondFiles := scanner.Process(scans2)

	hashes := make(map[string][32]byte, len(secondFiles))
	for _, fi := range secondFiles {
		hashes[fi.Path] = fi.Hash
	}
	assert.NotEqual(t, hashes[pathA], hashes[pathB], "verification should have updated a.txt's hash to reflect the in-place edit")
}

func TestScannerVerifyFailureDrops(t *testing.T) {
	root := t.TempDir()
	pathA := writeFile(t, root, "a.txt", "matching content")
	pathB := writeFile(t, root, "b.txt", "matching content")

	c, _ := newCache(t)
	scans, err := newCollector(nil).collectRoots(root)
	require.NoError(t, err)
	require.Len(t, scans, 1)

	// First scan populates cache.
	scanner := newScanner(c, 4, nil, false)
	_ = scanner.Process(scans)
	require.NoError(t, c.Flush())

	// Delete a.txt while the cache still has its entry. The folder mtime WILL bump
	// because deletion bumps it, so the trusted path won't trigger — but we want to
	// test the verify-failure path. We can manually re-set folder mtime to match.
	require.NoError(t, os.Remove(pathA))
	folderInfo, err := os.Stat(scans[0].Path)
	require.NoError(t, err)
	// Re-collect to capture the new mtime, then set cache to match (simulates an
	// in-place modification that happened to remove rather than edit).
	require.NoError(t, c.SetFolderMtime(scans[0].Path, folderInfo.ModTime()))
	require.NoError(t, c.Flush())

	// Now simulate "I saw 2 files in this folder last scan, but only 1 is here now."
	// We can't reproduce that via collectRoots (which sees current state); instead we
	// craft a FolderScan that claims both children still exist.
	scan := FolderScan{
		Path:     scans[0].Path,
		Mtime:    folderInfo.ModTime(),
		Children: []string{filepath.Base(pathA), filepath.Base(pathB)},
	}
	files := scanner.Process([]FolderScan{scan})

	// b.txt should survive; a.txt's claim is no longer verifiable.
	var bSeen, aSeen bool
	for _, fi := range files {
		if fi.Path == pathA {
			aSeen = true
		}
		if fi.Path == pathB {
			bSeen = true
		}
	}
	assert.False(t, aSeen, "deleted file should be dropped by the verifier")
	assert.True(t, bSeen, "surviving duplicate-group member should remain")
}
