// Package main implements a File Duplicate Scanner.
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
	"sort"
	"strings"
	"sync"

	flag "github.com/spf13/pflag"
)

var (
	verbose = flag.BoolP("verbose", "v", false, "verbose output (debug-level logging)")
	quiet   = flag.BoolP("quiet", "q", false, "quiet output (warnings and errors only)")
	minSize = flag.Int64P("min-size", "m", 1024, "ignore duplicates smaller than this many bytes")
)

// FileInfo track per-file statistics
type FileInfo struct {
	Path      string
	Size      int64
	Extension string
	Hash      string
}

// ExtensionStats tracks total per extension
type ExtensionStats struct {
	Count int
	Size  int64
}

// analyse prints summary, per-extension breakdown, and duplicate groups.
func analyse(files []FileInfo, minSize int64) {
	printSummary(files)
	printByExtension(files)
	printDuplicates(files, minSize)
}

// printSummary writes the total file count and combined size.
func printSummary(files []FileInfo) {
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	fmt.Printf("Found %d files (%s)\n\n", len(files), formatSize(totalSize))
}

// printByExtension writes a per-extension breakdown sorted by file count desc.
func printByExtension(files []FileInfo) {
	stats := make(map[string]*ExtensionStats)
	for _, f := range files {
		if _, ok := stats[f.Extension]; !ok {
			stats[f.Extension] = &ExtensionStats{}
		}
		stats[f.Extension].Count++
		stats[f.Extension].Size += f.Size
	}

	type extEntry struct {
		Ext   string
		Stats *ExtensionStats
	}
	exts := make([]extEntry, 0, len(stats))
	for ext, s := range stats {
		name := ext
		if name == "" {
			name = "(none)"
		}
		exts = append(exts, extEntry{name, s})
	}
	sort.Slice(exts, func(i, j int) bool {
		return exts[i].Stats.Count > exts[j].Stats.Count
	})

	fmt.Println("By extension:")
	for _, e := range exts {
		fmt.Printf("  %-10s %4d files   %s\n", e.Ext, e.Stats.Count, formatSize(e.Stats.Size))
	}
}

// printDuplicates writes duplicate groups at or above minSize, largest first.
func printDuplicates(files []FileInfo, minSize int64) {
	groups := make(map[string][]FileInfo)
	for _, f := range files {
		if f.Hash != "" {
			groups[f.Hash] = append(groups[f.Hash], f)
		}
	}

	var dups [][]FileInfo
	for _, group := range groups {
		if len(group) < 2 || group[0].Size < minSize {
			continue
		}
		dups = append(dups, group)
	}
	sort.Slice(dups, func(i, j int) bool {
		if dups[i][0].Size != dups[j][0].Size {
			return dups[i][0].Size > dups[j][0].Size
		}
		return filepath.Base(dups[i][0].Path) < filepath.Base(dups[j][0].Path)
	})

	fmt.Printf("\nDuplicates (>= %s):\n", formatSize(minSize))
	if len(dups) == 0 {
		fmt.Println("  No duplicates found.")
		return
	}
	for _, group := range dups {
		name := filepath.Base(group[0].Path)
		fmt.Printf("  %s (%d copies, %s each)\n", name, len(group), formatSize(group[0].Size))
	}
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

// formatSize produces human-formatted file size string
func formatSize(bytes int64) string {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
	)
	switch {
	case bytes >= GiB:
		return fmt.Sprintf("%.1f GiB", float64(bytes)/float64(GiB))
	case bytes >= MiB:
		return fmt.Sprintf("%.1f MiB", float64(bytes)/float64(MiB))
	case bytes >= KiB:
		return fmt.Sprintf("%.1f KiB", float64(bytes)/float64(KiB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
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
func processFile(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	fi := FileInfo{
		Path:      path,
		Size:      info.Size(),
		Extension: filepath.Ext(path),
	}
	hash, err := hashFile(path)
	if err != nil {
		return fi, err
	}
	fi.Hash = hash
	return fi, nil
}

// processFiles stats and hashes every path using a fixed pool of workers
func processFiles(paths []string) []FileInfo {
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
				info, err := processFile(path)
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

// setupLogging installs a slog default logger that writes to stderr at the
// given level. The time attribute is stripped because this CLI is short-lived
// and the per-line timestamp is noise rather than signal.
func setupLogging(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
}

func main() {
	flag.Parse()

	level := slog.LevelInfo
	switch {
	case *verbose:
		level = slog.LevelDebug
	case *quiet:
		level = slog.LevelWarn
	}
	setupLogging(level)

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: filescan [-v|-q] [-m bytes] FOLDER(S)...")
		os.Exit(1)
	}

	roots := flag.Args()
	fmt.Printf("Scanning: %s\n", strings.Join(roots, ", "))

	paths, err := collectRoots(roots...)
	fmt.Printf("Found %d files\n", len(paths))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(paths) == 0 {
		fmt.Println("No files found.")
		return
	}

	files := processFiles(paths)
	analyse(files, *minSize)
}
