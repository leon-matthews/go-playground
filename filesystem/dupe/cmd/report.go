package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"local.dev/monarch"
)

// ExtensionStats tracks total per extension
type ExtensionStats struct {
	Count int
	Size  int64
}

// analyse prints summary, per-extension breakdown, and duplicate groups.
func analyse(files []monarch.FileInfo, minSize int64) {
	printSummary(files)
	printByExtension(files)
	printDuplicates(files, minSize)
}

// printSummary writes the total file count and combined size.
func printSummary(files []monarch.FileInfo) {
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	fmt.Printf("Found %d files (%s)\n\n", len(files), formatSize(totalSize))
}

// printByExtension writes a per-extension breakdown sorted by file count desc.
func printByExtension(files []monarch.FileInfo) {
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
func printDuplicates(files []monarch.FileInfo, minSize int64) {
	groups := make(map[[32]byte][]monarch.FileInfo)
	for _, f := range files {
		if f.Hash != ([32]byte{}) {
			groups[f.Hash] = append(groups[f.Hash], f)
		}
	}

	var dups [][]monarch.FileInfo
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
