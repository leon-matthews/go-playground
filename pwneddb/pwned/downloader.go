package pwned

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"

	"pwneddb/database/sqlite"
)

// Report progress at this interval unless [Downloader.Progress] is set
const defaultProgressInterval = 30 * time.Second

// A Downloader fetches every hash list and stores it in the database.
// Stored ETags let repeat runs skip lists that have not changed, so an
// interrupted download resumes where it left off.
type Downloader struct {
	db      *sql.DB
	queries *sqlite.Queries

	// Limit stops the run after this many prefixes, if greater than zero
	Limit int

	// Progress is the interval between progress reports
	Progress time.Duration
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

// Run fetches and stores every prefix in sequence, returning early if ctx
// is cancelled. Individual fetch failures are logged and skipped.
func (d *Downloader) Run(ctx context.Context) error {
	interval := d.Progress
	if interval <= 0 {
		interval = defaultProgressInterval
	}

	start := time.Now()
	next := start.Add(interval)
	var c counters

	for prefix := range Prefixes() {
		if err := ctx.Err(); err != nil {
			d.report(ctx, start, c, true)
			return err
		}
		if d.Limit > 0 && c.processed >= d.Limit {
			break
		}

		etag, err := d.storedEtag(ctx, prefix)
		if err != nil {
			return err
		}

		resp, err := FetchHashes(ctx, prefix, etag)
		if err != nil {
			// Cancellation mid-fetch is not a failure; loop top returns
			if ctx.Err() != nil {
				continue
			}
			slog.Error("fetch failed", "prefix", prefix, "error", err)
			c.failed++
			continue
		}

		if resp.HTTPStatus == http.StatusNotModified {
			c.unchanged++
		} else {
			if err := d.store(ctx, resp); err != nil {
				return fmt.Errorf("storing prefix %q: %w", prefix, err)
			}
			c.fetched++
			c.downloaded += int64(len(resp.Hashes))
		}

		c.processed++
		if time.Now().After(next) {
			d.report(ctx, start, c, false)
			next = time.Now().Add(interval)
		}
	}

	d.report(ctx, start, c, true)
	return nil
}

// storedEtag returns the ETag held for prefix, or "" if it is not yet stored.
func (d *Downloader) storedEtag(ctx context.Context, prefix Prefix) (string, error) {
	row, err := d.queries.GetPrefix(ctx, string(prefix))
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading prefix %q: %w", prefix, err)
	}
	return row.Etag.String, nil
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
