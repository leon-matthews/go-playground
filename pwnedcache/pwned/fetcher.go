package pwned

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Give up on any single request after this long
const fetchTimeout = 30 * time.Second

// Identify ourselves to the pwned passwords API, as good manners request
const userAgent = "pwnedcache/0.1"

// Back off between retries starting here, doubling up to maxRetryDelay
const (
	retryBaseDelay = 1 * time.Second
	maxRetryDelay  = 10 * time.Second
)

// client is shared by all fetch workers
var client = newClient()

// newClient builds an HTTP client that suits many concurrent workers.
// The default transport keeps only two idle connections per host, which
// would force most workers to redo the TCP and TLS handshakes every fetch.
func newClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 128
	transport.MaxIdleConnsPerHost = 128
	return &http.Client{
		Timeout:   fetchTimeout,
		Transport: transport,
	}
}

// A fetchError reports an unexpected HTTP status from the pwned API.
// It carries any Retry-After hint so a caller can back off politely.
type fetchError struct {
	StatusCode int
	RetryAfter time.Duration
	URL        string
}

func (e *fetchError) Error() string {
	return fmt.Sprintf("unexpected status %d from %q", e.StatusCode, e.URL)
}

// A HashResponse is returned by [FetchHashes]
type HashResponse struct {
	Etag       string   // As provided from upstream
	Hashes     HashList // Colon-separated SHA1(password):count pairs
	Prefix     Prefix   // The hexadecimal prefix shared by these hashes
	HTTPStatus int      // Either 200 or 304
}

// BuildURL produces URL from a hash prefix
func BuildURL(prefix Prefix) string {
	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	return url
}

// FetchHashes gets password hashes from the pwned passwords API.
//
// An optional ETag avoids re-downloading a hash list we already have. The body
// is read into buf, so the returned Hashes stay valid only until buf is reused.
func FetchHashes(ctx context.Context, prefix Prefix, etag string, buf *bytes.Buffer) (*HashResponse, error) {
	url := BuildURL(prefix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("If-None-Match", etag)

	start := time.Now()
	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer r.Body.Close()

	// Not modified?
	if r.StatusCode == http.StatusNotModified {
		elapsed := time.Since(start)
		slog.LogAttrs(
			ctx,
			slog.LevelDebug,
			"hashes unchanged",
			slog.Duration("elapsed", elapsed),
			slog.Int("status", r.StatusCode),
			slog.String("url", url),
		)
		res := HashResponse{
			Etag:       etag,
			HTTPStatus: r.StatusCode,
			Prefix:     prefix,
		}
		return &res, nil
	}

	// Anything other than a fresh hash list is an error, eg. 429 or 500
	if r.StatusCode != http.StatusOK {
		return nil, &fetchError{
			StatusCode: r.StatusCode,
			RetryAfter: parseRetryAfter(r.Header.Get("Retry-After")),
			URL:        url,
		}
	}

	// Reusing buf spares a hot worker from regrowing a read buffer every prefix
	buf.Reset()
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	body := buf.Bytes()
	elapsed := time.Since(start)
	rEtag := r.Header.Get("Etag")
	slog.LogAttrs(
		ctx,
		slog.LevelDebug,
		"fetched new hashes",
		slog.Int("bytes", len(body)),
		slog.Duration("elapsed", elapsed),
		slog.Int("status", r.StatusCode),
		slog.String("url", url),
	)

	res := HashResponse{
		Etag:       rEtag,
		Hashes:     HashList(body),
		Prefix:     prefix,
		HTTPStatus: r.StatusCode,
	}
	return &res, nil
}

// fetchWithRetry fetches a prefix's hash list, retrying transient failures with
// a capped backoff. It gives up after maxRetries retries, or as soon as ctx is
// cancelled, since an aborted run is not a failure worth retrying.
func fetchWithRetry(ctx context.Context, prefix Prefix, etag string, maxRetries int, buf *bytes.Buffer) (*HashResponse, error) {
	var err error
	for attempt := 0; ; attempt++ {
		var resp *HashResponse
		resp, err = FetchHashes(ctx, prefix, etag, buf)
		if err == nil {
			return resp, nil
		}
		if ctx.Err() != nil || attempt >= maxRetries {
			return nil, err
		}

		delay := retryDelay(attempt, err)
		slog.LogAttrs(
			ctx, slog.LevelDebug, "retrying fetch",
			slog.String("prefix", string(prefix)),
			slog.Int("attempt", attempt+1),
			slog.Duration("delay", delay),
			slog.String("error", err.Error()),
		)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// retryDelay returns how long to wait before the next attempt. A server's
// Retry-After takes precedence; otherwise the delay doubles each attempt, both
// capped at maxRetryDelay.
func retryDelay(attempt int, err error) time.Duration {
	var fe *fetchError
	if errors.As(err, &fe) && fe.RetryAfter > 0 {
		return min(fe.RetryAfter, maxRetryDelay)
	}
	// A large attempt shifts past int64, so treat overflow as the cap
	delay := retryBaseDelay << attempt
	if delay <= 0 || delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}

// parseRetryAfter interprets a Retry-After header, which is either a whole
// number of seconds or an HTTP date. It returns zero when absent or invalid.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		if delay := time.Until(when); delay > 0 {
			return delay
		}
	}
	return 0
}
