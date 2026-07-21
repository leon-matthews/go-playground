// Package mimicry scans directory trees for duplicate files via SHA-256 and a SQLite hash cache.
package mimicry

import (
	"crypto/sha256"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"local.dev/mimicry/progress"
)

// FileInfo track per-file statistics
type FileInfo struct {
	Path      string
	Size      int64
	ModTime   time.Time
	Extension string
	Hash      [32]byte
}

// FolderInfo is one folder visited during the walk: path, current mtime, and file names
type FolderInfo struct {
	Path     string
	Mtime    time.Time
	Children []string
}

// Collector owns the results of a directory walk: the per-folder scans, the absolute paths
// of the original roots, and a shared logger. One-shot - call Walk once, then read the fields.
type Collector struct {
	Folders  []FolderInfo
	AbsRoots []string
	log      *slog.Logger
}

// NewCollector returns a Collector; a nil logger is replaced with a discard logger.
func NewCollector(log *slog.Logger) *Collector {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Collector{log: log}
}

// Walk visits each root and populates Folders + AbsRoots. Errors only on missing or
// unreadable roots; non-directory roots are logged and skipped.
func (c *Collector) Walk(roots ...string) error {
	start := time.Now()
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
	c.log.Info("found files", slog.Int("count", c.TotalFiles()), slog.Duration("elapsed", time.Since(start)))
	return nil
}

// TotalFiles sums the children of every FolderInfo.
func (c *Collector) TotalFiles() int {
	n := 0
	for _, f := range c.Folders {
		n += len(f.Children)
	}
	return n
}

// collectRoots walks each root concurrently and returns deduplicated FolderInfo values. Non-directory
// roots are logged and skipped; the first missing/unreadable root errors out.
func (c *Collector) collectRoots(roots ...string) ([]FolderInfo, error) {
	type result struct {
		folders []FolderInfo
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
	var folders []FolderInfo
	for _, r := range results {
		for _, f := range r.folders {
			if _, dup := seenFolders[f.Path]; dup {
				continue
			}
			seenFolders[f.Path] = struct{}{}
			folders = append(folders, f)
		}
	}
	return folders, nil
}

// collectRoot walks root and returns one FolderInfo per directory below it. Roots that exist
// but aren't directories are logged as errors and produce no scans; missing roots still error.
// Symlinks and unreadable subtrees are skipped.
func (c *Collector) collectRoot(root string) ([]FolderInfo, error) {
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

	folders := make(map[string]*FolderInfo)
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
			folders[path] = &FolderInfo{Path: path, Mtime: info.ModTime()}
			return nil
		}
		if d.Type().IsRegular() {
			parent := filepath.Dir(path)
			if f, ok := folders[parent]; ok {
				f.Children = append(f.Children, d.Name())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := make([]FolderInfo, 0, len(folders))
	for _, f := range folders {
		result = append(result, *f)
	}
	return result, nil
}

// hashFile calculates the SHA-256 hash for the file with the given path.
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

// Scanner owns the worker pool and per-scan dependencies (cache, logger, progress).
type Scanner struct {
	cache      *Cache
	log        *slog.Logger
	prog       *progress.Progress
	maxWorkers int
	force      bool
}

// NewScanner returns a Scanner; a nil logger is replaced with a discard logger, and a nil
// progress disables progress counting.
func NewScanner(cache *Cache, maxWorkers int, log *slog.Logger, force bool, prog *progress.Progress) *Scanner {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Scanner{cache: cache, log: log, prog: prog, maxWorkers: maxWorkers, force: force}
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
		s.prog.FromCache(1)
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
	s.prog.Hashed(path, fi.Size)
	return fi, nil
}

// folderStats aggregates per-folder outcomes so Process can report a single summary line.
type folderStats struct {
	trustedFolders int
	staleFolders   int
	cachedFiles    int
	fellThrough    int
	dispatched     int
}

func (s *folderStats) merge(o folderStats) {
	s.trustedFolders += o.trustedFolders
	s.staleFolders += o.staleFolders
	s.cachedFiles += o.cachedFiles
	s.fellThrough += o.fellThrough
	s.dispatched += o.dispatched
}

// Process stats and hashes every file under folderInfos. Folders whose mtime matches the cache
// are served directly from cached entries; other folders go through the per-file worker pool.
// After collecting all results, files appearing in candidate duplicate groups are verified by
// re-stat (and re-hashed on drift) to catch in-place modifications.
func (s *Scanner) Process(folderInfos []FolderInfo) []FileInfo {
	start := time.Now()
	totalJobs := 0
	for _, f := range folderInfos {
		totalJobs += len(f.Children)
	}
	numWorkers := max(min(totalJobs, s.maxWorkers), 1)

	jobs := make(chan string, numWorkers)
	results := make(chan FileInfo, numWorkers)
	fromCache := make(map[string]bool)
	var totals folderStats
	var producers sync.WaitGroup

	s.log.Info("starting workers", "count", numWorkers)
	for range numWorkers {
		producers.Go(func() {
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

	producers.Go(func() {
		for _, scan := range folderInfos {
			stats, ok := s.tryTrustedFolder(scan, jobs, results, fromCache)
			if !ok {
				stats = s.dispatchStaleFolder(scan, jobs)
			}
			totals.merge(stats)
		}
		close(jobs)
	})

	go func() {
		producers.Wait()
		close(results)
	}()

	var files []FileInfo
	for f := range results {
		files = append(files, f)
	}

	verified := s.verifyDuplicates(files, fromCache)

	s.log.Info(
		"scanner finished",
		slog.Duration("elapsed", time.Since(start)),
		"folders_trusted", totals.trustedFolders,
		"folders_stale", totals.staleFolders,
		"files_from_cache", totals.cachedFiles,
		"files_fell_through", totals.fellThrough,
		"files_dispatched", totals.dispatched,
	)
	return verified
}

// tryTrustedFolder serves scan from the cache if its mtime hasn't drifted. Returns ok=true with
// the per-folder stats when handled; ok=false to fall through to dispatchStaleFolder.
func (s *Scanner) tryTrustedFolder(scan FolderInfo, jobs chan<- string, results chan<- FileInfo, fromCache map[string]bool) (folderStats, bool) {
	if s.force {
		return folderStats{}, false
	}
	mtime, ok := s.cache.GetFolderMtime(scan.Path)
	if !ok || !mtime.Equal(scan.Mtime) {
		return folderStats{}, false
	}
	entries, err := s.cache.GetFilesInFolder(scan.Path)
	if err != nil {
		s.log.Warn("trusted-folder bulk fetch failed; falling through", "folder", scan.Path, "err", err)
		return folderStats{}, false
	}
	stats := folderStats{trustedFolders: 1}
	for _, name := range scan.Children {
		fullPath := filepath.Join(scan.Path, name)
		if entry, ok := entries[name]; ok {
			results <- FileInfo{
				Path:      fullPath,
				Size:      entry.Size,
				ModTime:   entry.ModTime,
				Extension: filepath.Ext(name),
				Hash:      entry.Hash,
			}
			fromCache[fullPath] = true
			stats.cachedFiles++
		} else {
			jobs <- fullPath
			stats.fellThrough++
		}
	}
	s.prog.FromCache(int64(stats.cachedFiles))
	return stats, true
}

// dispatchStaleFolder enqueues every child for hashing and updates the folder's mtime so the
// next scan can trust it.
func (s *Scanner) dispatchStaleFolder(scan FolderInfo, jobs chan<- string) folderStats {
	stats := folderStats{staleFolders: 1}
	for _, name := range scan.Children {
		jobs <- filepath.Join(scan.Path, name)
		stats.dispatched++
	}
	_ = s.cache.SetFolderMtime(scan.Path, scan.Mtime)
	return stats
}

// verifyDuplicates re-stats files from cache that appear in candidate duplicate groups; on
// drift it re-hashes inline and pushes the new entry back to cache. On stat/hash failure the
// file is dropped from the results.
func (s *Scanner) verifyDuplicates(infos []FileInfo, fromCache map[string]bool) []FileInfo {
	if len(fromCache) == 0 {
		return infos
	}

	start := time.Now()
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
	s.log.Info("verifying duplicates finished", slog.Duration("elapsed", time.Since(start)))
	return cleaned
}
