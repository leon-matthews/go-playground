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

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
