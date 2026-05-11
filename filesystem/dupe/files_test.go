package main

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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

func TestCollectRoots(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		root := t.TempDir()

		paths, err := collectRoots(root)

		require.NoError(t, err)
		assert.Empty(t, paths)
	})

	t.Run("flat directory of files", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root, "a.txt", "b.txt", "c.log")

		paths, err := collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, paths)
	})

	t.Run("nested directories", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root,
			"top.txt",
			"sub/one.txt",
			"sub/two.txt",
			"sub/deep/three.txt",
		)

		paths, err := collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, paths)
	})

	t.Run("directories are excluded from results", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "empty-dir"), 0o755))
		want := writeFiles(t, root, "only.txt")

		paths, err := collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, paths)
	})

	t.Run("missing root returns error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "does-not-exist")

		paths, err := collectRoots(missing)

		assert.Error(t, err)
		assert.Empty(t, paths)
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

		paths, err := collectRoots(root)

		require.NoError(t, err)
		assert.Contains(t, paths, readable[0], "readable file should still be returned")
	})

	t.Run("multiple roots are combined", func(t *testing.T) {
		rootA := t.TempDir()
		rootB := t.TempDir()
		wantA := writeFiles(t, rootA, "a.txt", "sub/b.txt")
		wantB := writeFiles(t, rootB, "c.txt")

		paths, err := collectRoots(rootA, rootB)

		require.NoError(t, err)
		assert.ElementsMatch(t, append(wantA, wantB...), paths)
	})

	t.Run("overlapping roots produce no duplicates", func(t *testing.T) {
		root := t.TempDir()
		want := writeFiles(t, root, "top.txt", "sub/inner.txt")

		paths, err := collectRoots(root, filepath.Join(root, "sub"))

		require.NoError(t, err)
		assert.ElementsMatch(t, want, paths)
	})

	t.Run("missing root among others returns error", func(t *testing.T) {
		good := t.TempDir()
		writeFiles(t, good, "a.txt")
		missing := filepath.Join(t.TempDir(), "does-not-exist")

		_, err := collectRoots(good, missing)

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

		paths, err := collectRoots(root)

		require.NoError(t, err)
		assert.ElementsMatch(t, want, paths)
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
