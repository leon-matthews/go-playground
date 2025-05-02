package main

import (
	"context"
	"log/slog"

	"main/pwned"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const prefix = "cafe5"
	url, _ := pwned.BuildURL(prefix)

	// Fetch page of password hashes
	r, err := pwned.GetHashes(ctx, url, "")
	if err != nil {
		panic(err)
	}

	// Refetching with eTag should return no body
	_, err = pwned.GetHashes(ctx, url, r.Etag)
	if err != nil {
		panic(err)
	}
}
