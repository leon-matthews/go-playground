package main

import (
	"fmt"
	"io"
	"path/filepath"

	"local.dev/monarch"
)

// printReport writes the scan summary, per-extension breakdown, and duplicate groups.
func printReport(w io.Writer, files []monarch.FileInfo, minSize int64) {
	printSummary(w, monarch.Summarize(files))
	printExtensions(w, monarch.ExtensionStats(files))
	printDuplicates(w, monarch.DuplicateGroups(files), minSize)
}

// printSummary writes the total file count and combined size.
func printSummary(w io.Writer, s monarch.Summary) {
	fmt.Fprintf(w, "Found %d files (%s)\n\n", s.Count, formatSize(s.Size))
}

// printExtensions writes a per-extension breakdown.
func printExtensions(w io.Writer, stats []monarch.ExtensionStat) {
	fmt.Fprintln(w, "By extension:")
	for _, s := range stats {
		name := s.Extension
		if name == "" {
			name = "(none)"
		}
		fmt.Fprintf(w, "  %-10s %4d files   %s\n", name, s.Count, formatSize(s.Size))
	}
}

// printDuplicates writes duplicate groups at or above minSize, largest first.
func printDuplicates(w io.Writer, groups []monarch.DuplicateGroup, minSize int64) {
	fmt.Fprintf(w, "\nDuplicates (>= %s):\n", formatSize(minSize))
	var shown int
	for _, g := range groups {
		if g.Size < minSize {
			continue
		}
		name := filepath.Base(g.Files[0].Path)
		fmt.Fprintf(w, "  %s (%d copies, %s each)\n", name, len(g.Files), formatSize(g.Size))
		shown++
	}
	if shown == 0 {
		fmt.Fprintln(w, "  No duplicates found.")
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
