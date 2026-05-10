package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// FileInfo track per-file statistics
type FileInfo struct {
	Path      string
	Size      int64
	ModTime   time.Time
	Extension string
	Hash      string
}

// collectRoot returns absolute paths to every regular file under root.
// Symlinks and unreadable subtrees are skipped; an unreadable root errors.
func collectRoot(root string) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var paths []string
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Missing root?
			if path == absRoot {
				return walkErr
			}

			// skip files we can't read
			return nil
		}

		if d.Type().IsRegular() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

// collectRoots walks each root concurrently and returns deduplicated absolute
// paths to every regular file found. Files reachable via more than one root
// are returned once.
func collectRoots(roots ...string) ([]string, error) {
	type result struct {
		paths []string
		err   error
	}
	results := make([]result, len(roots))

	var wg sync.WaitGroup
	for i, root := range roots {
		wg.Go(func() {
			results[i].paths, results[i].err = collectRoot(root)
		})
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
	}

	seen := make(map[string]struct{})
	var paths []string
	for _, r := range results {
		for _, p := range r.paths {
			if _, dup := seen[p]; !dup {
				seen[p] = struct{}{}
				paths = append(paths, p)
			}
		}
	}
	return paths, nil
}

// hashFile calculates a SHA-256 hash for the file with the given path
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// processFile stats and hashes a single path. A stat failure returns a zero
// FileInfo and the error, signalling the caller to skip the file. A hash
// failure returns a populated FileInfo with an empty Hash AND the error: the
// FileInfo still feeds summary stats, while the empty Hash excludes the file
// from duplicate detection. Callers distinguish the two cases by inspecting
// info.Path: empty means stat failed.
//
// A cache hit (matching Size and ModTime) skips the hash read entirely. On a
// fresh hash the entry is written back via cache.Set so the work is durable
// before the function returns; a Set error is logged at warn level but does
// not fail the file.
func processFile(path string, cache *Cache) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	fi := FileInfo{
		Path:      path,
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		Extension: filepath.Ext(path),
	}
	if e, ok := cache.Get(path); ok && e.Size == fi.Size && e.ModTime.Equal(fi.ModTime) {
		fi.Hash = e.Hash
		return fi, nil
	}
	hash, err := hashFile(path)
	if err != nil {
		return fi, err
	}
	fi.Hash = hash
	if err := cache.Set(path, CacheEntry{Size: fi.Size, ModTime: fi.ModTime, Hash: fi.Hash}); err != nil {
		slog.Warn("cache: failed to write entry", "path", path, "err", err)
	}
	return fi, nil
}

// processFiles stats and hashes every path using a fixed pool of workers.
// Workers consult cache concurrently; cache writes are coalesced internally
// via bbolt's Batch.
func processFiles(paths []string, cache *Cache) []FileInfo {
	numWorkers := min(len(paths), runtime.NumCPU())
	jobs := make(chan string, numWorkers)
	results := make(chan FileInfo, numWorkers)
	var wg sync.WaitGroup

	// Start a fixed pool of workers
	slog.Info("starting workers", "count", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			for path := range jobs {
				slog.Debug("Reading file", "worker", i, "file", path)
				info, err := processFile(path, cache)
				if err != nil {
					if info.Path == "" {
						slog.Warn("stat failed; skipping file", "path", path, "err", err)
						continue
					}
					slog.Warn("hash failed; excluding from duplicates", "path", path, "err", err)
				}
				results <- info
			}
		})
	}

	// Feeder: send every path, then close jobs
	go func() {
		for _, p := range paths {
			jobs <- p
		}
		close(jobs)
	}()

	// Closer: when all workers exit, close results
	go func() {
		wg.Wait()
		close(results)
	}()

	var files []FileInfo
	for f := range results {
		files = append(files, f)
	}
	return files
}
