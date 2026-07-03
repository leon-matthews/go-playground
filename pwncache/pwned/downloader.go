package pwned

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/dustin/go-humanize"

	"pwncache/database"
	"pwncache/database/sqlite"
)

// Report progress at this interval unless [Downloader.Progress] is set
const defaultProgressInterval = 10 * time.Second

// Run this many fetch workers unless [Downloader.Concurrency] is set
const defaultConcurrency = 64

// Small channel buffers decouple the pipeline stages a little
const channelBuffer = 64

// A Downloader fetches every hash list and stores it in the database.
// Stored ETags let repeat runs skip lists that have not changed, so an
// interrupted download resumes where it left off.
type Downloader struct {
	db      *sql.DB
	queries *sqlite.Queries

	// Concurrency is the number of parallel fetch workers
	Concurrency int

	// Limit stops the run after attempting this many prefixes, if greater than zero
	Limit int

	// Progress is the interval between progress reports
	Progress time.Duration

	// MaxRetries is how many times to retry a failed fetch before skipping it
	MaxRetries int

	// ConsoleLog receives friendly one-line progress; FileLog receives the
	// matching structured record. Both fall back to slog.Default() when nil.
	ConsoleLog *slog.Logger
	FileLog    *slog.Logger
}

// A result carries one prefix's fetch outcome from a worker to the collector
type result struct {
	prefix Prefix
	resp   *HashResponse
	err    error
}

// counters accumulates totals over a whole download run
type counters struct {
	processed  int   // Prefixes handled, one way or another
	fetched    int   // Fresh hash lists downloaded and stored
	unchanged  int   // Hash lists skipped thanks to a matching ETag
	failed     int   // Fetches that failed and were skipped
	downloaded int64 // Bytes of hash list data received
}

// NewDownloader builds a Downloader backed by the given database.
func NewDownloader(db *sql.DB, queries *sqlite.Queries) *Downloader {
	return &Downloader{db: db, queries: queries}
}

// Run fetches every prefix using parallel workers, returning early if ctx
// is cancelled. Individual fetch failures are logged and skipped.
func (d *Downloader) Run(ctx context.Context) error {
	interval := d.Progress
	if interval <= 0 {
		interval = defaultProgressInterval
	}
	workers := d.Concurrency
	if workers <= 0 {
		workers = defaultConcurrency
	}

	etags, cached, err := d.loadEtags(ctx)
	if err != nil {
		return err
	}
	d.reportCacheLoaded(ctx, cached)

	inserter, err := database.NewHashInserter(ctx, d.db)
	if err != nil {
		return err
	}
	defer inserter.Close()

	// Cancelling shuts down the generator and workers if storing fails
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	prefixes := d.generate(ctx)
	results := fetchWorkers(ctx, workers, d.MaxRetries, etags, prefixes)

	// Run's own goroutine collects results, and is the sole database writer,
	// so the counters need no locks and SQLite sees no write contention.
	start := time.Now()
	since, sinceProcessed := start, 0
	next := start.Add(interval)
	var c counters
	for res := range results {
		// Drain remaining results without processing once cancelled
		if ctx.Err() != nil {
			continue
		}
		if res.err != nil {
			slog.Error("fetch failed", "prefix", res.prefix, "error", res.err)
			c.failed++
			continue
		}

		if res.resp.HTTPStatus == http.StatusNotModified {
			c.unchanged++
		} else {
			if err := d.store(ctx, inserter, res.resp); err != nil {
				// Cancellation mid-store is not a failure either
				if ctx.Err() != nil {
					continue
				}
				return fmt.Errorf("storing prefix %q: %w", res.prefix, err)
			}
			c.fetched++
			c.downloaded += int64(len(res.resp.Hashes))
		}

		c.processed++
		if now := time.Now(); now.After(next) {
			rate := ratePerSecond(c.processed-sinceProcessed, now.Sub(since))
			d.report(ctx, c, rate, now.Sub(start), false)
			since, sinceProcessed = now, c.processed
			next = now.Add(interval)
		}
	}

	// The summary covers the whole run, so its rate is the overall average
	end := time.Now()
	d.report(ctx, c, ratePerSecond(c.processed, end.Sub(start)), end.Sub(start), true)
	return ctx.Err()
}

// loadEtags reads every stored ETag into a slice indexed by prefix value, and
// reports how many prefixes are cached. The dense prefix keyspace makes a slice
// cheaper than a map: the index is the key, so no keys are stored.
func (d *Downloader) loadEtags(ctx context.Context) ([]string, int, error) {
	rows, err := d.queries.GetEtags(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("loading etags: %w", err)
	}
	etags := make([]string, PrefixCount)
	cached := 0
	for _, row := range rows {
		if !row.Etag.Valid {
			continue
		}
		index, err := Prefix(row.Prefix).Index()
		if err != nil {
			// An unplaceable prefix is skipped; it simply gets re-fetched
			slog.Warn("skipping malformed cached prefix", "prefix", row.Prefix, "error", err)
			continue
		}
		etags[index] = row.Etag.String
		cached++
	}
	return etags, cached, nil
}

// etagFor returns the stored ETag for a prefix, or "" when it is not cached.
func etagFor(etags []string, prefix Prefix) string {
	index, err := prefix.Index()
	if err != nil {
		return ""
	}
	return etags[index]
}

// reportCacheLoaded announces how much of the database is already cached, both
// as a sign of life and as a completeness figure before the download begins.
func (d *Downloader) reportCacheLoaded(ctx context.Context, cached int) {
	percent := 100 * float64(cached) / PrefixCount
	d.fileLogger().LogAttrs(
		ctx, slog.LevelInfo, "cache loaded",
		slog.Int("cached", cached),
		slog.Int("total", PrefixCount),
		slog.Float64("percent", math.Round(percent*10)/10),
	)
	d.consoleLogger().Info(fmt.Sprintf(
		"%s of %s prefixes already cached (%.1f%%) - starting download",
		humanize.Comma(int64(cached)), humanize.Comma(int64(PrefixCount)), percent,
	))
}

// generate feeds prefixes into a channel until done, limited, or cancelled.
func (d *Downloader) generate(ctx context.Context) <-chan Prefix {
	out := make(chan Prefix, channelBuffer)
	go func() {
		defer close(out)
		count := 0
		for prefix := range Prefixes() {
			if d.Limit > 0 && count >= d.Limit {
				return
			}
			select {
			case out <- prefix:
				count++
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// fetchWorkers starts count parallel workers to fetch each prefix's hash list.
// The results channel closes once the last worker has finished.
func fetchWorkers(
	ctx context.Context,
	count int,
	maxRetries int,
	etags []string,
	prefixes <-chan Prefix,
) <-chan result {
	out := make(chan result, channelBuffer)
	var wg sync.WaitGroup
	for range count {
		wg.Go(func() {
			for prefix := range prefixes {
				resp, err := fetchWithRetry(ctx, prefix, etagFor(etags, prefix), maxRetries)
				select {
				case out <- result{prefix: prefix, resp: resp, err: err}:
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// store replaces the prefix's hashes and download metadata in one transaction.
func (d *Downloader) store(ctx context.Context, inserter *database.HashInserter, resp *HashResponse) error {
	hashes, err := ParseHashList(resp.Prefix, resp.Hashes)
	if err != nil {
		return err
	}
	lower, upper, err := resp.Prefix.HashRange()
	if err != nil {
		return err
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// A no-op after commit, but unwinds cleanly on any error return
	defer tx.Rollback()

	qtx := d.queries.WithTx(tx)
	bounds := sqlite.DeleteHashRangeParams{Lower: lower, Upper: upper}
	if err := qtx.DeleteHashRange(ctx, bounds); err != nil {
		return fmt.Errorf("deleting old hashes: %w", err)
	}
	rows := make([]sqlite.InsertHashParams, len(hashes))
	for i, hash := range hashes {
		rows[i] = sqlite.InsertHashParams{Hash: hash.SHA1, Count: hash.Count}
	}
	if err := inserter.Insert(ctx, tx, rows); err != nil {
		return err
	}

	err = qtx.UpsertPrefix(ctx, sqlite.UpsertPrefixParams{
		Prefix:  string(resp.Prefix),
		Updated: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		Etag:    sql.NullString{String: resp.Etag, Valid: resp.Etag != ""},
	})
	if err != nil {
		return fmt.Errorf("updating metadata: %w", err)
	}
	return tx.Commit()
}

// report writes a friendly line to the console and the matching structured
// record to the file, as either "progress" or a final "summary". rate is the
// per-second rate over the reporting window; elapsed is the whole run so far.
func (d *Downloader) report(ctx context.Context, c counters, rate float64, elapsed time.Duration, final bool) {
	percent := 100 * float64(c.processed) / PrefixCount
	remaining, etaKnown := eta(c.processed, d.Limit, rate)

	// Structured record for the log file
	attrs := []slog.Attr{
		slog.Float64("percent", math.Round(percent*10)/10),
		slog.Int("processed", c.processed),
		slog.Int("fetched", c.fetched),
		slog.Int("unchanged", c.unchanged),
		slog.Int("failed", c.failed),
		slog.String("downloaded", humanize.Bytes(uint64(c.downloaded))),
		slog.Float64("per_second", math.Round(rate*10)/10),
		slog.Duration("elapsed", elapsed.Round(time.Second)),
	}
	kind := "summary"
	if !final {
		kind = "progress"
		// A stalled window has no estimate, so omit eta rather than log a false 0
		if etaKnown {
			attrs = append(attrs, slog.Duration("eta", remaining))
		}
	}
	d.fileLogger().LogAttrs(ctx, slog.LevelInfo, kind, attrs...)

	// Friendly one-line version for the console
	d.consoleLogger().Info(humanReport(c, rate, percent, elapsed, remaining, etaKnown, final))
}

// humanReport renders a friendly one-line progress or summary message.
func humanReport(c counters, rate, percent float64, elapsed, remaining time.Duration, etaKnown, final bool) string {
	counts := fmt.Sprintf(
		"%s fetched, %s unchanged, %s failed",
		humanize.Comma(int64(c.fetched)),
		humanize.Comma(int64(c.unchanged)),
		humanize.Comma(int64(c.failed)),
	)
	size := humanize.Bytes(uint64(c.downloaded))
	perSecond := humanize.Comma(int64(math.Round(rate)))

	if final {
		return fmt.Sprintf(
			"Finished %s of %s prefixes: %s - %s at %s/s in %s",
			humanize.Comma(int64(c.processed)), humanize.Comma(int64(PrefixCount)),
			counts, size, perSecond, elapsed.Round(time.Second),
		)
	}
	etaText := "unknown"
	if etaKnown {
		etaText = remaining.String()
	}
	return fmt.Sprintf(
		"Processed %s prefixes (%.1f%%): %s - %s at %s/s, ETA %s",
		humanize.Comma(int64(c.processed)),
		percent, counts, size, perSecond, etaText,
	)
}

// consoleLogger returns the friendly console logger, or the default if unset.
func (d *Downloader) consoleLogger() *slog.Logger {
	if d.ConsoleLog != nil {
		return d.ConsoleLog
	}
	return slog.Default()
}

// fileLogger returns the structured file logger, or the default if unset.
func (d *Downloader) fileLogger() *slog.Logger {
	if d.FileLog != nil {
		return d.FileLog
	}
	return slog.Default()
}

// ratePerSecond is processed items per second over window, guarding a zero window.
func ratePerSecond(processed int, window time.Duration) float64 {
	if window <= 0 {
		return 0
	}
	return float64(processed) / window.Seconds()
}

// eta estimates the time remaining to finish the run at the given rate. ok is
// false when the rate is too low to estimate, i.e. the run has stalled.
func eta(processed, limit int, rate float64) (remaining time.Duration, ok bool) {
	target := PrefixCount
	if limit > 0 && limit < PrefixCount {
		target = limit
	}
	left := target - processed
	if left <= 0 {
		return 0, true // essentially done, so zero is honest
	}
	if rate <= 0 {
		return 0, false // stalled: no basis for an estimate
	}
	return time.Duration(float64(left) / rate * float64(time.Second)).Round(time.Second), true
}
