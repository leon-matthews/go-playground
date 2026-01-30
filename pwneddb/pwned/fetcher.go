package pwned

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// A HashResponse is returned by [FetchHashes]
type HashResponse struct {
	Etag       string   // As provided from upstream
	Hashes     HashList // Colon-separated SHA1(password):count pairs
	Prefix     Prefix   // The hexadecimal prefix shared by these hashes
	HTTPStatus int      // Hopefully 200 or 304
}

// BuildURL produces URL from a hash prefix
func BuildURL(prefix Prefix) string {
	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	return url
}

// FetchHashes gets passwords hashes from the pwned passwords API
// An ETag can optionally be provided to avoid having to re-download a
// hash list that we already have.
func FetchHashes(ctx context.Context, prefix Prefix, etag string) (*HashResponse, error) {
	url := BuildURL(prefix)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("If-None-Match", etag)

	start := time.Now()
	r, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer r.Body.Close()

	// Not modified?
	if r.StatusCode == 304 {
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

	// Body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
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
