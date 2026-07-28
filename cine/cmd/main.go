// Command cine builds and queries a local SQLite database holding the IMDb
// non-commercial datasets.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
)

func main() {
	// Cancel the run cleanly on Ctrl-C: a build discards its temporary file, so
	// any database already at the target path is left untouched.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := newRootCmd().ExecuteContext(ctx)
	// Stopped here, not in a PersistentPostRunE, so the profile still flushes
	// if RunE returns an error, which skips persistent post-run hooks.
	if stopProfile != nil {
		stopProfile()
	}
	if err != nil {
		// Ctrl-C surfaces as a cancelled context at the end of a long error chain,
		// which reads like a crash rather than the deliberate stop it was.
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "interrupted")
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

// startProfile begins writing a CPU profile to the given path.
//
// Returns the stop function that ends profiling, closes the file, and reports
// where the profile went. A plain create suffices, as a spoiled profile is
// simply overwritten by the next run.
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
