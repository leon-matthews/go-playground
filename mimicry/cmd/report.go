package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"local.dev/mimicry"
	"local.dev/mimicry/logging"
	"local.dev/mimicry/tui"
)

// reportOptions collects the report command's flags.
type reportOptions struct {
	minSize    int64
	extensions bool
	plain      bool
}

// newReportCmd builds the "report" sub-command.
func newReportCmd() *cobra.Command {
	var opts reportOptions
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Browse duplicate files from the cache",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReport(cmd.OutOrStdout(), cmd.ErrOrStderr(), logLevel(), opts)
		},
	}
	cmd.Flags().Int64VarP(&opts.minSize, "min-size", "m", 1024, "ignore duplicates smaller than this many bytes")
	cmd.Flags().BoolVarP(&opts.extensions, "extensions", "e", false, "report the per-extension breakdown instead of duplicates")
	cmd.Flags().BoolVar(&opts.plain, "plain", false, "print plain text instead of the interactive browser")
	return cmd
}

// runReport reads the cache and either launches the interactive browser or prints a plain report.
func runReport(stdout, stderr io.Writer, level slog.Level, opts reportOptions) error {
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

	summary := mimicry.Summarize(files)
	interactive := !opts.plain && isTTY(stdout)
	if opts.extensions {
		return reportExtensions(stdout, files, summary, interactive)
	}
	return reportDuplicates(stdout, files, summary, opts.minSize, interactive)
}

// reportDuplicates browses or prints the duplicate groups at or above minSize.
func reportDuplicates(stdout io.Writer, files []mimicry.FileInfo, summary mimicry.Summary, minSize int64, interactive bool) error {
	groups := aboveMinSize(mimicry.DuplicateGroups(files), minSize)
	if len(groups) == 0 {
		fmt.Fprintf(stdout, "No duplicates found (>= %s).\n", formatSize(minSize))
		return nil
	}
	if interactive {
		return tui.RunDuplicates(groups, summary)
	}
	printSummary(stdout, summary)
	printDuplicates(stdout, groups, minSize)
	return nil
}

// reportExtensions browses or prints the per-extension breakdown.
func reportExtensions(stdout io.Writer, files []mimicry.FileInfo, summary mimicry.Summary, interactive bool) error {
	stats := mimicry.ExtensionStats(files)
	if interactive {
		return tui.RunExtensions(stats, summary)
	}
	printSummary(stdout, summary)
	printExtensions(stdout, stats)
	return nil
}

// aboveMinSize keeps only the duplicate groups whose per-file size is at least minSize.
func aboveMinSize(groups []mimicry.DuplicateGroup, minSize int64) []mimicry.DuplicateGroup {
	var kept []mimicry.DuplicateGroup
	for _, g := range groups {
		if g.Size >= minSize {
			kept = append(kept, g)
		}
	}
	return kept
}

// isTTY reports whether w is a terminal, i.e. whether the interactive browser can run.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && isatty.IsTerminal(f.Fd())
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

// printDuplicates writes the given duplicate groups, most reclaimable first.
func printDuplicates(w io.Writer, groups []mimicry.DuplicateGroup, minSize int64) {
	fmt.Fprintf(w, "\nDuplicates (>= %s):\n", formatSize(minSize))
	for _, g := range groups {
		name := filepath.Base(g.Files[0].Path)
		fmt.Fprintf(w, "  %s (%d copies, %s each, %s reclaimable)\n",
			name, len(g.Files), formatSize(g.Size), formatSize(g.Reclaimable()))
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
