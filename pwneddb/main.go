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
	etagger := pwned.NewETagStore()
	etagger.Load("etags.txt")

	// Fetch page of password hashes
	r, err := pwned.FetchHashes(ctx, prefix, "")
	if err != nil {
		panic(err)
	}
	etagger[prefix] = r.Etag

	r, err = pwned.FetchHashes(ctx, prefix2, "")
	if err != nil {
		panic(err)
	}
	etagger[prefix2] = r.Etag

	// Save ETag
	etagger.Save("etags.txt")

	// Refetching with eTag should return no body
	_, err = pwned.FetchHashes(ctx, prefix, etagger[prefix])
	if err != nil {
		panic(err)
	}
	_, err = pwned.FetchHashes(ctx, prefix2, etagger[prefix2])
	if err != nil {
		panic(err)
	}
}
