package main

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"local.dev/mimicry"
	"local.dev/mimicry/logging"
	"local.dev/mimicry/progress"
)

// newScanCmd builds the "scan" sub-command.
func newScanCmd() *cobra.Command {
	var jobs int
	var force bool
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "scan ROOT...",
		Short: "Walk roots, hash files, and populate the cache",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, roots []string) error {
			return runScan(cmd.ErrOrStderr(), logLevel(), roots, jobs, force, progressInterval)
		},
	}
	cmd.Flags().IntVarP(&jobs, "jobs", "j", runtime.NumCPU(), "number of concurrent worker goroutines")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "stat every file, ignoring the folder-mtime cache")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", progress.DefaultInterval,
		"interval between progress reports")
	return cmd
}

// runScan walks roots, hashes files, populates the cache, and sweeps stale entries.
func runScan(stderr io.Writer, level slog.Level, roots []string, jobs int, force bool, progressInterval time.Duration) error {
	cacheFile, err := cachePath()
	if err != nil {
		return fmt.Errorf("cannot resolve cache path: %w", err)
	}
	log, logHandle := logging.Setup(level, filepath.Join(filepath.Dir(cacheFile), "mimicry.log"))
	if logHandle != nil {
		defer logHandle.Close()
	}

	if jobs < 1 {
		log.Warn("invalid --jobs value; clamping to 1", "value", jobs)
		jobs = 1
	}

	log.Info("scanning", "roots", roots)
	collector := mimicry.NewCollector(log)
	if err := collector.Walk(roots...); err != nil {
		return fmt.Errorf("collect roots failed: %w", err)
	}
	if collector.TotalFiles() == 0 {
		log.Warn("no files found")
		return nil
	}

	cache, cacheErr := mimicry.OpenCache(cacheFile, log)
	if cacheErr != nil {
		log.Warn("cache disabled", "path", cacheFile, "err", cacheErr)
	}
	defer cache.Close()

	prog := &progress.Progress{}
	reporter := progress.StartReporter(progressInterval, progress.ReportTo(prog, log, collector.TotalFiles()))

	scanner := mimicry.NewScanner(cache, jobs, log, force, prog)
	files := scanner.Process(collector.Folders)
	reporter.StopAndReport()

	if err := cache.Sweep(collector.Folders, collector.AbsRoots); err != nil {
		log.Warn("cache sweep failed", "err", err)
	}

	log.Info("scan complete", "files", len(files))
	if cacheErr == nil {
		fmt.Fprintf(stderr, "Cache written to %s\n", cacheFile)
	}
	return nil
}
