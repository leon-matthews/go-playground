// Package buildfilter scans the pwnedcache hashes into a split-block Bloom
// filter and writes it to disk.
package buildfilter

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

	"pwnedpasswords/database"
	"pwnedpasswords/filter"
	"pwnedpasswords/logging"
	"pwnedpasswords/progress"
)

// Preset pairs a filter size with the probe count that minimises its
// false-positive rate for the ~2 billion hash pwnedcache corpus. If that corpus
// grows substantially, retune using the sizing table on [filter.SplitBlockBloom].
type Preset struct {
	Blocks uint64
	Probes int
}

// The presets offered to the command, tuned to the model optima for the corpus.
var (
	Preset4GB  = Preset{filter.BlocksForBytes(4 << 30), 8}
	Preset8GB  = Preset{filter.BlocksForBytes(8 << 30), 16}
	Preset16GB = Preset{filter.BlocksForBytes(16 << 30), 24}
)

// Run scans every hash in the pwnedcache database into a split-block Bloom
// filter sized by preset and writes it to disk. It refuses to overwrite an
// existing filter file.
func Run(ctx context.Context, logs logging.Logging, cachePath, filterPath string, preset Preset, interval time.Duration) error {
	if _, err := os.Stat(filterPath); err == nil {
		return fmt.Errorf("filter %q already exists; remove it to rebuild", filterPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	_, cacheDB, err := database.OpenRO(ctx, cachePath, 1)
	if err != nil {
		return err
	}
	defer cacheDB.Close()

	built, err := filter.New(preset.Blocks, preset.Probes)
	if err != nil {
		return err
	}
	// Keep the GC goal near the huge live filter, rather than the default 2x
	debug.SetGCPercent(10)

	slog.Info("building filter",
		"blocks", preset.Blocks,
		"probes", preset.Probes,
		"size_gib", float64(preset.Blocks*64)/(1<<30))

	count, err := scanIntoFilter(ctx, logs, cacheDB, built, interval)
	if err != nil {
		return err
	}

	slog.Info("writing filter",
		"elements", count,
		"bits_per_element", float64(preset.Blocks*512)/float64(count),
		"path", filterPath)
	if err := built.Write(filterPath, cachePath); err != nil {
		return err
	}
	slog.Info("filter complete", "elements", count, "path", filterPath)
	return nil
}

// scanIntoFilter streams every hash from the cache into the filter, reporting
// progress on interval and a scan summary when done. It returns the hash count.
func scanIntoFilter(ctx context.Context, logs logging.Logging, cacheDB *sql.DB, built *filter.Filter, interval time.Duration) (int64, error) {
	rows, err := cacheDB.QueryContext(ctx, "SELECT hash FROM hashes")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	prog := &buildProgress{start: time.Now()}
	rep := progress.StartReporter(interval, prog.reportTo(logs.Console, logs.File))
	defer rep.StopAndReport()

	var count int64
	var hash sql.RawBytes // aliases the driver's row buffer; consumed before the next Scan
	for rows.Next() {
		if err := rows.Scan(&hash); err != nil {
			return count, err
		}
		built.Add(filter.SHA1Hash(hash))
		count++
		if count%progress.FlushEvery == 0 {
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
