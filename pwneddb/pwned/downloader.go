package pwned

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"pwneddb/database/sqlite"
)

// Log progress after this many prefixes have been processed
const progressInterval = 10_000

// A Downloader fetches every hash list and stores it in the database.
// Stored ETags let repeat runs skip lists that have not changed, so an
// interrupted download resumes where it left off.
type Downloader struct {
	queries *sqlite.Queries
}

// NewDownloader builds a Downloader backed by the given query store.
func NewDownloader(queries *sqlite.Queries) *Downloader {
	return &Downloader{queries: queries}
}

// Run fetches and stores every prefix in sequence, returning early if ctx
// is cancelled. Individual fetch failures are logged and skipped.
func (d *Downloader) Run(ctx context.Context) error {
	start := time.Now()
	var processed, fetched, unchanged int

	for prefix := range Prefixes() {
		if err := ctx.Err(); err != nil {
			return err
		}

		etag, err := d.storedEtag(ctx, prefix)
		if err != nil {
			return err
		}

		resp, err := FetchHashes(ctx, prefix, etag)
		if err != nil {
			slog.Error("fetch failed", "prefix", prefix, "error", err)
			continue
		}

		if resp.HTTPStatus == http.StatusNotModified {
			unchanged++
		} else {
			if err := d.store(ctx, resp); err != nil {
				return err
			}
			fetched++
		}

		processed++
		if processed%progressInterval == 0 {
			d.logProgress(ctx, start, processed, fetched, unchanged)
		}
	}

	d.logProgress(ctx, start, processed, fetched, unchanged)
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

// store writes a freshly fetched hash list to the database.
func (d *Downloader) store(ctx context.Context, resp *HashResponse) error {
	err := d.queries.UpsertPrefix(ctx, sqlite.UpsertPrefixParams{
		Prefix:  string(resp.Prefix),
		Updated: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		Etag:    sql.NullString{String: resp.Etag, Valid: resp.Etag != ""},
		Hashes:  string(resp.Hashes),
	})
	if err != nil {
		return fmt.Errorf("storing prefix %q: %w", resp.Prefix, err)
	}
	return nil
}

// logProgress reports how far the run has got, and its overall rate.
func (d *Downloader) logProgress(ctx context.Context, start time.Time, processed, fetched, unchanged int) {
	elapsed := time.Since(start)
	rate := float64(processed) / elapsed.Seconds()
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"progress",
		slog.Int("processed", processed),
		slog.Int("fetched", fetched),
		slog.Int("unchanged", unchanged),
		slog.Duration("elapsed", elapsed.Round(time.Second)),
		slog.Float64("per_second", rate),
	)
}
