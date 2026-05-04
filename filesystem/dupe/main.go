// Package main implements a File Duplicate Scanner.
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
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

// analyse prints stats to terminal from given data
func analyse(files []FileInfo) {
	var totalSize int64
	extMap := make(map[string]*ExtensionStats)
	hashMap := make(map[string][]FileInfo)

	for _, f := range files {
		totalSize += f.Size

		if _, ok := extMap[f.Extension]; !ok {
			extMap[f.Extension] = &ExtensionStats{}
		}
		extMap[f.Extension].Count++
		extMap[f.Extension].Size += f.Size

		if f.Hash != "" {
			hashMap[f.Hash] = append(hashMap[f.Hash], f)
		}
	}

	fmt.Printf("Found %d files (%s)\n\n", len(files), formatSize(totalSize))

	// Sort extensions by file count
	type extEntry struct {
		Ext   string
		Stats *ExtensionStats
	}
	var exts []extEntry
	for ext, stats := range extMap {
		name := ext
		if name == "" {
			name = "(none)"
		}
		exts = append(exts, extEntry{name, stats})
	}
	sort.Slice(exts, func(i, j int) bool {
		return exts[i].Stats.Count > exts[j].Stats.Count
	})

	fmt.Println("By extension:")
	for _, e := range exts {
		fmt.Printf("  %-10s %4d files   %s\n", e.Ext, e.Stats.Count, formatSize(e.Stats.Size))
	}

	// Find duplicates
	fmt.Println("\nDuplicates:")
	found := false
	for _, group := range hashMap {
		if len(group) < 2 {
			continue
		}
		found = true
		name := filepath.Base(group[0].Path)
		fmt.Printf("  %s (%d copies, %s each)\n", name, len(group), formatSize(group[0].Size))
	}
	if !found {
		fmt.Println("  No duplicates found.")
	}
}

// collectFiles builds a slice of absolute paths to all the files under root
func collectFiles(root string) ([]string, error) {
	var paths []string
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Missing root?
			if path == absRoot {
				return err
			}

			// skip files we can't read
			return nil
		}

		if d.Type().IsRegular() {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
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

// processFiles creates a goroutine for every path
func processFiles(paths []string) []FileInfo {
	// Create a huge buffered channel
	results := make(chan FileInfo, len(paths))
	var wg sync.WaitGroup

	// Start a goroutine for every entry in paths
	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			info, err := os.Stat(path)
			if err != nil {
				return
			}

			hash, _ := hashFile(path)

			results <- FileInfo{
				Path:      path,
				Size:      info.Size(),
				Extension: filepath.Ext(path),
				Hash:      hash,
			}
		}(p)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Empty channel into a slice
	var files []FileInfo
	for f := range results {
		files = append(files, f)
	}
	return files
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: filescan <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fmt.Printf("Scanning: %s\n", root)

	paths, err := collectFiles(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(paths) == 0 {
		fmt.Println("No files found.")
		return
	}

	files := processFiles(paths)
	analyse(files)
}
