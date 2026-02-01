package main

import (
	"context"
	"log"
	"log/slog"

	"pwneddb/pwned"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prefix, err := pwned.NewPrefix("cafe5")
	prefix2, err := pwned.NewPrefix("cafe6")
	if err != nil {
		log.Fatal(err)
	}

	// Load ETags
	etags, err := pwned.ETagsLoad("etags.txt")
	if err != nil {
		etags = make(pwned.ETags)
	}

	// Fetch page of password hashes
	r, err := pwned.FetchHashes(ctx, prefix, "")
	if err != nil {
		panic(err)
	}
	etags[prefix] = r.Etag

	r, err = pwned.FetchHashes(ctx, prefix2, "")
	if err != nil {
		panic(err)
	}
	etags[prefix2] = r.Etag

	// Save ETag
	etags.Save("etags.txt")

	// Refetching with eTag should return no body
	_, err = pwned.FetchHashes(ctx, prefix, etags[prefix])
	if err != nil {
		panic(err)
	}
	_, err = pwned.FetchHashes(ctx, prefix2, etags[prefix2])
	if err != nil {
		panic(err)
	}
}
