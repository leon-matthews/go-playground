// Command pwnedcache maintains and queries a local mirror of the Have I Been Pwned password database.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	// Cancel the run cleanly on Ctrl-C; state is safe in the database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
