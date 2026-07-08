package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/filter"
)

// filterPreset pairs a filter size with the probe count that minimises its
// false-positive rate for the ~2 billion hash pwnedcache corpus. If that corpus
// grows substantially, retune the probe counts against the new element count.
type filterPreset struct {
	blocks uint64
	probes int
}

var (
	preset4GB  = filterPreset{filter.BlocksForBytes(4 << 30), 10}
	preset8GB  = filterPreset{filter.BlocksForBytes(8 << 30), 16}
	preset16GB = filterPreset{filter.BlocksForBytes(16 << 30), 21}
)

// newBuildFilterCmd builds the "buildfilter" sub-command.
func newBuildFilterCmd() *cobra.Command {
	var use4GB, use8GB, use16GB bool
	var filterPath string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "buildfilter",
		Short: "Build the membership filter from the pwnedcache hashes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var preset filterPreset
			switch {
			case use4GB:
				preset = preset4GB
			case use8GB:
				preset = preset8GB
			case use16GB:
				preset = preset16GB
			}
			logs, err := setupLogging(verbose)
			if err != nil {
				return err
			}
			defer logs.logFile.Close()
			return runBuildFilter(cmd.Context(), logs, pwnedcachePath, filterPath, preset, progressInterval)
		},
	}
	cmd.Flags().BoolVar(&use4GB, "4GB", false, "4 GiB filter (false positives ~1 in 1,200)")
	cmd.Flags().BoolVar(&use8GB, "8GB", false, "8 GiB filter, suggested (false positives ~1 in 250,000)")
	cmd.Flags().BoolVar(&use16GB, "16GB", false, "16 GiB filter (false positives ~1 in 130 million)")
	cmd.MarkFlagsMutuallyExclusive("4GB", "8GB", "16GB")
	cmd.MarkFlagsOneRequired("4GB", "8GB", "16GB")
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter", "output filter file path")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
	return cmd
}

// runBuildFilter scans every hash in the pwnedcache database into a split-block
// Bloom filter sized by preset and writes it to disk. It refuses to overwrite an
// existing filter file.
func runBuildFilter(ctx context.Context, logs logging, cachePath, filterPath string, preset filterPreset, interval time.Duration) error {
	if _, err := os.Stat(filterPath); err == nil {
		return fmt.Errorf("filter %q already exists; remove it to rebuild", filterPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	_, cacheDB, err := database.OpenHashes(ctx, cachePath)
	if err != nil {
		return err
	}
	defer cacheDB.Close()

	built, err := filter.New(preset.blocks, preset.probes)
	if err != nil {
		return err
	}
	// Keep the GC goal near the huge live filter, rather than the default 2x
	debug.SetGCPercent(10)

	slog.Info("building filter",
		"blocks", preset.blocks,
		"probes", preset.probes,
		"size_gib", float64(preset.blocks*64)/(1<<30))

	count, err := scanIntoFilter(ctx, logs, cacheDB, built, interval)
	if err != nil {
		return err
	}

	slog.Info("writing filter",
		"elements", count,
		"bits_per_element", float64(preset.blocks*512)/float64(count),
		"path", filterPath)
	if err := built.Write(filterPath, cachePath); err != nil {
		return err
	}
	slog.Info("filter complete", "elements", count, "path", filterPath)
	return nil
}

// scanIntoFilter streams every hash from the cache into the filter, reporting
// progress on interval and a scan summary when done. It returns the hash count.
func scanIntoFilter(ctx context.Context, logs logging, cacheDB *sql.DB, built *filter.Filter, interval time.Duration) (int64, error) {
	rows, err := cacheDB.QueryContext(ctx, "SELECT hash FROM hashes")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	prog := &buildProgress{start: time.Now()}
	rep := startReporter(interval, prog.reportTo(logs.console, logs.file))
	defer rep.stopAndReport()

	var count int64
	var hash sql.RawBytes // aliases the driver's row buffer; consumed before the next Scan
	for rows.Next() {
		if err := rows.Scan(&hash); err != nil {
			return count, err
		}
		built.Add(filter.SHA1Hash(hash))
		count++
		if count%flushEvery == 0 {
			prog.added.Store(count)
		}
	}
	if err := rows.Err(); err != nil {
		return count, err
	}
	prog.added.Store(count)
	return count, nil
}

// buildProgress tracks the hashes scanned into the filter so the reporter can
// read the running total without contending on the scan loop.
type buildProgress struct {
	added atomic.Int64
	start time.Time
}

// reportTo returns a report function that logs scan progress as a friendly
// console line and the matching structured record to the file.
func (b *buildProgress) reportTo(console, file *slog.Logger) func(kind string) {
	return func(kind string) {
		added := b.added.Load()
		elapsed := time.Since(b.start)
		var rate int64
		if elapsed > 0 {
			rate = int64(float64(added) / elapsed.Seconds())
		}
		file.Info(kind, "added", added, "rate_per_sec", rate, "elapsed_sec", int64(elapsed.Seconds()))
		if kind == "summary" {
			console.Info(fmt.Sprintf("scanned %s hashes in %s (%s/s)",
				humanize.Comma(added), elapsed.Round(time.Second), humanize.Comma(rate)))
			return
		}
		console.Info(fmt.Sprintf("scanned %s hashes (%s/s)",
			humanize.Comma(added), humanize.Comma(rate)))
	}
}
