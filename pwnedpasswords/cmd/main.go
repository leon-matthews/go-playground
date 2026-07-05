// Command pwnedpasswords builds breach-frequency password denylists from word lists.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	// Cancel the run cleanly on Ctrl-C; state is safe in the database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := newRootCmd().ExecuteContext(ctx)
	// Stopped here, not in a PersistentPostRunE, so the profile still flushes
	// if RunE returns an error, which skips persistent post-run hooks.
	if stopProfile != nil {
		stopProfile()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
