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

// FolderScan is one folder visited during the walk: path, current mtime, and basenames of its
// direct regular-file children.
type FolderScan struct {
	Path     string
	Mtime    time.Time
	Children []string
}

// Collector owns the results of a directory walk: the per-folder scans, the absolute paths
// of the original roots, and a shared logger. One-shot — call Walk once, then read the fields.
type Collector struct {
	Folders  []FolderScan
	AbsRoots []string

	log *slog.Logger
}

// newCollector returns a Collector; a nil logger is replaced with a discard logger.
func newCollector(log *slog.Logger) *Collector {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Collector{log: log}
}

// Walk visits each root and populates Folders + AbsRoots. Errors only on missing or
// unreadable roots; non-directory roots are logged and skipped.
func (c *Collector) Walk(roots ...string) error {
	folders, err := c.collectRoots(roots...)
	if err != nil {
		return err
	}
	c.Folders = folders
	c.AbsRoots = make([]string, 0, len(roots))
	for _, r := range roots {
		abs, absErr := filepath.Abs(r)
		if absErr != nil {
			c.log.Warn("cannot resolve absolute root", "root", r, "err", absErr)
			continue
		}
		c.AbsRoots = append(c.AbsRoots, abs)
	}
	c.log.Info("found files", "count", c.TotalFiles())
	return nil
}

// TotalFiles sums the children of every FolderScan.
func (c *Collector) TotalFiles() int {
	n := 0
	for _, fs := range c.Folders {
		n += len(fs.Children)
	}
	return n
}

// collectRoots walks each root concurrently and returns deduplicated FolderScans. Non-directory
// roots are logged and skipped; the first missing/unreadable root errors out.
func (c *Collector) collectRoots(roots ...string) ([]FolderScan, error) {
	type result struct {
		folders []FolderScan
		err     error
	}
	results := make([]result, len(roots))

	var wg sync.WaitGroup
	for i, root := range roots {
		wg.Go(func() {
			results[i].folders, results[i].err = c.collectRoot(root)
		})
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
	}

	seenFolders := make(map[string]struct{})
	var folders []FolderScan
	for _, r := range results {
		for _, fs := range r.folders {
			if _, dup := seenFolders[fs.Path]; dup {
				continue
			}
			seenFolders[fs.Path] = struct{}{}
			folders = append(folders, fs)
		}
	}
	return folders, nil
}

// collectRoot walks root and returns one FolderScan per directory below it. Roots that exist
// but aren't directories are logged as errors and produce no scans; missing roots still error.
// Symlinks and unreadable subtrees are skipped.
func (c *Collector) collectRoot(root string) ([]FolderScan, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	rootInfo, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	if !rootInfo.IsDir() {
		c.log.Error("root is not a directory; ignoring", "root", absRoot)
		return nil, nil
	}

	folders := make(map[string]*FolderScan)
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == absRoot {
				return walkErr
			}
			return nil
		}
		if d.IsDir() {
			info, infoErr := d.Info()
			if infoErr != nil {
				return nil
			}
			folders[path] = &FolderScan{Path: path, Mtime: info.ModTime()}
			return nil
		}
		if d.Type().IsRegular() {
			parent := filepath.Dir(path)
			if fs, ok := folders[parent]; ok {
				fs.Children = append(fs.Children, d.Name())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := make([]FolderScan, 0, len(folders))
	for _, fs := range folders {
		result = append(result, *fs)
	}
	return result, nil
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
	force      bool
}

// newScanner returns a Scanner; a nil logger is replaced with a discard logger.
func newScanner(cache *Cache, maxWorkers int, log *slog.Logger, force bool) *Scanner {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Scanner{cache: cache, log: log, maxWorkers: maxWorkers, force: force}
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

// Process stats and hashes every file under folderScans. Folders whose mtime matches the cache
// are served directly from cached entries; other folders go through the per-file worker pool.
// After collecting all results, files appearing in candidate duplicate groups are verified by
// re-stat (and re-hashed on drift) to catch in-place modifications.
func (s *Scanner) Process(folderScans []FolderScan) []FileInfo {
	totalJobs := 0
	for _, fs := range folderScans {
		totalJobs += len(fs.Children)
	}
	numWorkers := min(totalJobs, s.maxWorkers)
	if numWorkers < 1 {
		numWorkers = 1
	}

	jobs := make(chan string, numWorkers)
	results := make(chan FileInfo, numWorkers)
	var wg sync.WaitGroup

	s.log.Info("starting workers", "count", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			for path := range jobs {
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

	var (
		cachedResults   []FileInfo
		fromCache       = make(map[string]bool)
		trustedFolders  int
		staleFolders    int
		cachedFiles     int
		fellThroughHits int
		dispatchedFiles int
	)
	dispatcherDone := make(chan struct{})

	go func() {
		defer close(dispatcherDone)
		for _, scan := range folderScans {
			trusted := false
			if !s.force {
				if cachedMtime, ok := s.cache.GetFolderMtime(scan.Path); ok && cachedMtime.Equal(scan.Mtime) {
					entries, err := s.cache.GetFilesInFolder(scan.Path)
					if err != nil {
						s.log.Warn("trusted-folder bulk fetch failed; falling through", "folder", scan.Path, "err", err)
					} else {
						trusted = true
						for _, name := range scan.Children {
							fullPath := filepath.Join(scan.Path, name)
							if entry, ok := entries[name]; ok {
								cachedResults = append(cachedResults, FileInfo{
									Path:      fullPath,
									Size:      entry.Size,
									ModTime:   entry.ModTime,
									Extension: filepath.Ext(name),
									Hash:      entry.Hash,
								})
								fromCache[fullPath] = true
								cachedFiles++
							} else {
								jobs <- fullPath
								fellThroughHits++
							}
						}
					}
				}
			}
			if !trusted {
				for _, name := range scan.Children {
					jobs <- filepath.Join(scan.Path, name)
					dispatchedFiles++
				}
				_ = s.cache.SetFolderMtime(scan.Path, scan.Mtime)
			}
			if trusted {
				trustedFolders++
			} else {
				staleFolders++
			}
		}
		close(jobs)
		s.log.Info("scan summary",
			"folders_trusted", trustedFolders,
			"folders_stale", staleFolders,
			"files_from_cache", cachedFiles,
			"files_fell_through", fellThroughHits,
			"files_dispatched", dispatchedFiles,
		)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var files []FileInfo
	for f := range results {
		files = append(files, f)
	}
	<-dispatcherDone
	files = append(files, cachedResults...)

	return s.verifyDuplicates(files, fromCache)
}

// verifyDuplicates re-stats files from cache that appear in candidate duplicate groups; on
// drift it re-hashes inline and pushes the new entry back to cache. On stat/hash failure the
// file is dropped from the results.
func (s *Scanner) verifyDuplicates(infos []FileInfo, fromCache map[string]bool) []FileInfo {
	if len(fromCache) == 0 {
		return infos
	}

	groups := make(map[[32]byte][]int)
	for i, fi := range infos {
		if fi.Hash == ([32]byte{}) {
			continue
		}
		groups[fi.Hash] = append(groups[fi.Hash], i)
	}

	drops := make(map[int]bool)
	for _, indices := range groups {
		if len(indices) < 2 {
			continue
		}
		for _, idx := range indices {
			fi := &infos[idx]
			if !fromCache[fi.Path] {
				continue
			}
			st, err := os.Stat(fi.Path)
			if err != nil {
				s.log.Warn("verify: stat failed; dropping from duplicate group", "path", fi.Path, "err", err)
				drops[idx] = true
				continue
			}
			if st.Size() == fi.Size && st.ModTime().Equal(fi.ModTime) {
				continue
			}
			newHash, err := hashFile(fi.Path)
			if err != nil {
				s.log.Warn("verify: rehash failed; dropping from duplicate group", "path", fi.Path, "err", err)
				drops[idx] = true
				continue
			}
			fi.Size = st.Size()
			fi.ModTime = st.ModTime()
			fi.Hash = newHash
			_ = s.cache.Set(fi.Path, CacheEntry{Size: st.Size(), ModTime: st.ModTime(), Hash: newHash})
		}
	}

	if len(drops) == 0 {
		return infos
	}
	cleaned := make([]FileInfo, 0, len(infos)-len(drops))
	for i, fi := range infos {
		if !drops[i] {
			cleaned = append(cleaned, fi)
		}
	}
	return cleaned
}
