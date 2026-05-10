package main

import (
	"encoding/gob"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CacheEntry is the persisted hash record for a single file, keyed by its
// absolute path. Size and ModTime are the verification fields: a cache hit
// requires both to match the current stat result, otherwise the hash is
// recomputed and the entry overwritten.
type CacheEntry struct {
	Size    int64
	ModTime time.Time
	Hash    string
}

// cachePath returns the absolute path of the persistent hash cache.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dupe", "cache.gob"), nil
}

// loadCache reads the persistent hash cache from path. A missing file yields
// an empty cache without error; a corrupt file is logged and replaced with an
// empty cache so a single bad write never blocks future runs.
func loadCache(path string) map[string]CacheEntry {
	empty := map[string]CacheEntry{}
	f, err := os.Open(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Warn("cache disabled: cannot open file", "path", path, "err", err)
		}
		return empty
	}
	defer f.Close()

	var m map[string]CacheEntry
	if err := gob.NewDecoder(f).Decode(&m); err != nil {
		slog.Warn("cache corrupt; starting empty", "path", path, "err", err)
		return empty
	}
	slog.Debug("cache loaded", "path", path, "entries", len(m))
	return m
}

// saveCache writes the cache to path atomically via temp file + rename, so a
// crash mid-write cannot leave a half-written file in place of the previous
// good one.
func saveCache(path string, m map[string]CacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "cache-*.gob")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if err := gob.NewEncoder(tmp).Encode(m); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
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

// updateCache writes fresh hashes back to cache and sweeps in-scope orphans.
// An entry is in scope when its path lives under one of the current roots; a
// scoped entry not present in paths (so the file no longer exists) is removed,
// while out-of-scope entries from prior runs are left untouched.
func updateCache(cache map[string]CacheEntry, files []FileInfo, paths, roots []string) {
	for _, f := range files {
		if f.Hash != "" {
			cache[f.Path] = CacheEntry{Size: f.Size, ModTime: f.ModTime, Hash: f.Hash}
		}
	}

	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		seen[p] = struct{}{}
	}

	absRoots := make([]string, 0, len(roots))
	for _, r := range roots {
		a, err := filepath.Abs(r)
		if err != nil {
			slog.Warn("cannot resolve absolute root for cache sweep", "root", r, "err", err)
			continue
		}
		absRoots = append(absRoots, a)
	}

	for p := range cache {
		if _, ok := seen[p]; ok {
			continue
		}
		if pathInRoots(p, absRoots) {
			delete(cache, p)
		}
	}
}
