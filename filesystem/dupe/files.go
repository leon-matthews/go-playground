package main

import (
	"crypto/sha256"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileInfo track per-file statistics
type FileInfo struct {
	Path      string
	Size      int64
	ModTime   time.Time
	Extension string
	Hash      [32]byte
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

// hashFile calculates a SHA-256 hash for the file with the given path.
func hashFile(path string) ([32]byte, error) {
	var out [32]byte
	f, err := os.Open(path)
	if err != nil {
		return out, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return out, err
	}
	h.Sum(out[:0])
	return out, nil
}

// Scanner owns the worker pool and per-scan dependencies (cache, logger).
type Scanner struct {
	cache      *Cache
	log        *slog.Logger
	maxWorkers int
}

// newScanner returns a Scanner; a nil logger is replaced with a discard logger.
func newScanner(cache *Cache, maxWorkers int, log *slog.Logger) *Scanner {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Scanner{cache: cache, log: log, maxWorkers: maxWorkers}
}

// processFile stats and hashes a single path; see Process for the contract.
func (s *Scanner) processFile(path string) (FileInfo, error) {
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

	if e, ok := s.cache.Get(path); ok && e.Size == fi.Size && e.ModTime.Equal(fi.ModTime) {
		fi.Hash = e.Hash
		return fi, nil
	}

	hash, err := hashFile(path)
	if err != nil {
		return fi, err
	}
	fi.Hash = hash
	if err := s.cache.Set(path, CacheEntry{Size: fi.Size, ModTime: fi.ModTime, Hash: fi.Hash}); err != nil {
		s.log.Warn("cache: failed to write entry", "path", path, "err", err)
	}
	return fi, nil
}

// Process stats and hashes every path using a fixed pool of workers.
func (s *Scanner) Process(paths []string) []FileInfo {
	numWorkers := min(len(paths), s.maxWorkers)
	jobs := make(chan string, numWorkers)
	results := make(chan FileInfo, numWorkers)
	var wg sync.WaitGroup

	s.log.Info("starting workers", "count", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			for path := range jobs {
				s.log.Debug("Reading file", "worker", i, "file", path)
				info, err := s.processFile(path)
				if err != nil {
					if info.Path == "" {
						s.log.Warn("stat failed; skipping file", "path", path, "err", err)
						continue
					}
					s.log.Warn("hash failed; excluding from duplicates", "path", path, "err", err)
				}
				results <- info
			}
		})
	}

	go func() {
		for _, p := range paths {
			jobs <- p
		}
		close(jobs)
	}()

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
