package pwned

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"time"
)

// buildURL produces absolute URL from password hash prefix
func BuildURL(prefix string) (string, error) {
	if len(prefix) != 5 {
		return "", errors.New("prefix wrong length")
	}
	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	return url, nil
}

type basicResponse struct {
	Text       string
	Etag       string
	StatusCode int
}

// GetHashes fetches passwords hashes from the pwned passwords API
func GetHashes(ctx context.Context, url, etag string) (*basicResponse, error) {
	//~ client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("If-None-Match", etag)
	//~ req.Header.Add("Accept-Encoding", "gzip")

	start := time.Now()
	//~ r, err := client.Do(req)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer r.Body.Close()

	// Content not modified?
	if r.StatusCode == 304 {
		elapsed := time.Since(start)
		slog.Debug("get response", "status", r.StatusCode, "elapsed", elapsed)
		res := basicResponse{
			"",
			etag,
			r.StatusCode,
		}
		return &res, nil
	}

	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	elapsed := time.Since(start)
	rEtag := r.Header.Get("ETag")
	slog.Debug("get response", "status", r.StatusCode, "bytes", len(body), "elapsed", elapsed, "etag", rEtag, "url", url, "compressed", r.Uncompressed)

	res := basicResponse{
		string(body), // We assume body is ASCII
		rEtag,
		r.StatusCode,
	}
	return &res, nil
}

// HexStrings generates all hexadecimal strings of the given length, zero-padded.
func HexStrings(length int) iter.Seq[string] {
	limit := 0x01 << (length * 4)
	return func(yield func(string) bool) {
		for v := range limit {
			hex := fmt.Sprintf("%0*x", length, v)
			if !yield(hex) {
				return
			}
		}
	}
}
