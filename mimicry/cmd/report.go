package main

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	"local.dev/mimicry"
	"local.dev/mimicry/logging"
)

// newReportCmd builds the "report" sub-command.
func newReportCmd() *cobra.Command {
	var minSize int64
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Read the cache and report duplicate files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReport(cmd.OutOrStdout(), cmd.ErrOrStderr(), logLevel(), minSize)
		},
	}
	cmd.Flags().Int64VarP(&minSize, "min-size", "m", 1024, "ignore duplicates smaller than this many bytes")
	return cmd
}

// runReport reads every cached file and prints the summary, per-extension, and duplicate report.
func runReport(stdout, stderr io.Writer, level slog.Level, minSize int64) error {
	cacheFile, err := cachePath()
	if err != nil {
		return fmt.Errorf("cannot resolve cache path: %w", err)
	}
	fmt.Fprintf(stderr, "Reading cache from %s\n", cacheFile)

	cache, err := mimicry.OpenCache(cacheFile, logging.NewConsoleLogger(level))
	if err != nil {
		return fmt.Errorf("open cache: %w", err)
	}
	defer cache.Close()

	files, err := cache.AllFiles()
	if err != nil {
		return fmt.Errorf("read cache: %w", err)
	}
	if len(files) == 0 {
		fmt.Fprintln(stdout, "Cache is empty; run 'mimicry scan' first.")
		return nil
	}

	printReport(stdout, files, minSize)
	return nil
}

// printReport writes the scan summary, per-extension breakdown, and duplicate groups.
func printReport(w io.Writer, files []mimicry.FileInfo, minSize int64) {
	printSummary(w, mimicry.Summarize(files))
	printExtensions(w, mimicry.ExtensionStats(files))
	printDuplicates(w, mimicry.DuplicateGroups(files), minSize)
}

// printSummary writes the total file count and combined size.
func printSummary(w io.Writer, s mimicry.Summary) {
	fmt.Fprintf(w, "Found %d files (%s)\n\n", s.Count, formatSize(s.Size))
}

// printExtensions writes a per-extension breakdown.
func printExtensions(w io.Writer, stats []mimicry.ExtensionStat) {
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
func printDuplicates(w io.Writer, groups []mimicry.DuplicateGroup, minSize int64) {
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
