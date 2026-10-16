// Command pwnedpasswords builds breach-frequency password denylists from word lists.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
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

// startProfile begins writing a CPU profile to the given path.
//
// Returns the stop function that ends profiling, closes the file, and reports
// where the profile went. A plain create suffices here, with none of the care
// writeResults takes, as a spoiled profile is simply overwritten next run.
func startProfile(path string) (func(), error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return nil, err
	}
	return func() {
		pprof.StopCPUProfile()
		file.Close()
		fmt.Fprintf(os.Stderr, "CPU profile written to %s\n", path)
	}, nil
}
