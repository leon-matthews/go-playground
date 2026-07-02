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

	"pwncache/database/sqlite"
)

// Report progress at this interval unless [Downloader.Progress] is set
const defaultProgressInterval = 30 * time.Second

// Run this many fetch workers unless [Downloader.Concurrency] is set
const defaultConcurrency = 32

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

	etags, err := d.loadEtags(ctx)
	if err != nil {
		return err
	}

	// Cancelling shuts down the generator and workers if storing fails
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	prefixes := d.generate(ctx)
	results := fetchWorkers(ctx, workers, d.MaxRetries, etags, prefixes)

	// Run's own goroutine collects results, and is the sole database writer,
	// so the counters need no locks and SQLite sees no write contention.
	start := time.Now()
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
			if err := d.store(ctx, res.resp); err != nil {
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
		if time.Now().After(next) {
			d.report(ctx, start, c, false)
			next = time.Now().Add(interval)
		}
	}

	d.report(ctx, start, c, true)
	return ctx.Err()
}

// loadEtags reads every stored ETag into memory for the fetch workers.
func (d *Downloader) loadEtags(ctx context.Context) (map[Prefix]string, error) {
	rows, err := d.queries.GetEtags(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading etags: %w", err)
	}
	etags := make(map[Prefix]string, len(rows))
	for _, row := range rows {
		if row.Etag.Valid {
			etags[Prefix(row.Prefix)] = row.Etag.String
		}
	}
	return etags, nil
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
	etags map[Prefix]string,
	prefixes <-chan Prefix,
) <-chan result {
	out := make(chan result, channelBuffer)
	var wg sync.WaitGroup
	for range count {
		wg.Go(func() {
			for prefix := range prefixes {
				resp, err := fetchWithRetry(ctx, prefix, etags[prefix], maxRetries)
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
func (d *Downloader) store(ctx context.Context, resp *HashResponse) error {
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
	for _, hash := range hashes {
		row := sqlite.InsertHashParams{Hash: hash.SHA1, Count: hash.Count}
		if err := qtx.InsertHash(ctx, row); err != nil {
			return fmt.Errorf("inserting hash: %w", err)
		}
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

// report logs overall counters and rates, as either "progress" or a final "summary".
func (d *Downloader) report(ctx context.Context, start time.Time, c counters, final bool) {
	elapsed := time.Since(start)
	rate := float64(c.processed) / elapsed.Seconds()
	percent := 100 * float64(c.processed) / PrefixCount
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

	msg := "summary"
	if !final {
		msg = "progress"
		attrs = append(attrs, slog.Duration("eta", eta(c.processed, d.Limit, rate)))
	}
	slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// eta estimates the time remaining to finish the run at the given rate.
func eta(processed, limit int, rate float64) time.Duration {
	target := PrefixCount
	if limit > 0 && limit < PrefixCount {
		target = limit
	}
	remaining := target - processed
	if rate <= 0 || remaining <= 0 {
		return 0
	}
	return time.Duration(float64(remaining) / rate * float64(time.Second)).Round(time.Second)
}
