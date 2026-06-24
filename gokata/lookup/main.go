// Lookup looks up IP addresses concurrently. Hosts are supplied as STDIN stream.
// Adapted from https://youtu.be/woCg2zaIVzQ.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

const (
	lookupTimeout = 10 * time.Second
	numWorkers = 10
)

type lookup struct {
	host  string
	addrs []string
	err   error
}

func main() {
	// Cancel the whole pipeline on Ctrl-C; both stages watch ctx.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// A second layer the reader can cancel, carrying its error as the cause.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// First Ctrl-C cancels ctx for a graceful shutdown; restore the default
	// handler so a second Ctrl-C force-quits if shutdown can't make progress.
	go func() {
		<-ctx.Done()
		stop()
	}()

	hosts := readHosts(ctx, cancel, os.Stdin)
	results := lookupHosts(ctx, hosts)

	// Print errors on stderr and lookups on stdout.
	for lookup := range results {
		if lookup.err != nil {
			fmt.Fprintf(os.Stderr, "lookup: %v\n", lookup.err)
		} else {
			fmt.Printf("%s: %s\n", lookup.host, strings.Join(lookup.addrs, ", "))
		}
	}

	// A non-nil cause means the reader failed or we were interrupted.
	if err := context.Cause(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// readHosts pushes lines from stdin into the returned channel until ctx is cancelled.
func readHosts(ctx context.Context, cancel context.CancelCauseFunc, r io.Reader) <-chan string {
	hosts := make(chan string)
	go func() {
		defer close(hosts)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			// The blocking Scan can't be cancelled, but the send can.
			select {
			case hosts <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
		// Fail through the context so the pipeline drains instead of exiting here.
		if err := scanner.Err(); err != nil {
			cancel(fmt.Errorf("reading: %w", err))
		}
	}()
	return hosts
}

func lookupHosts(ctx context.Context, hosts <-chan string) <-chan lookup {
	// Spin-up workers to look-up hosts
	lookups := make(chan lookup)

	// Each worker pulls work from hosts (until its closed) then writes results to lookups.
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for host := range hosts {
				// Don't start a new lookup if we're already cancelled
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Deadline derived from ctx, so the 10s timeout is per lookup
				lookupCtx, cancel := context.WithTimeout(ctx, lookupTimeout)
				addrs, err := net.DefaultResolver.LookupHost(lookupCtx, host)
				cancel()

				select {
				case lookups <- lookup{host: host, addrs: addrs, err: err}:
				case <-ctx.Done():
					return
				}
			}
		})
	}

	// Close lookups once all the works are finished
	go func() {
		wg.Wait()
		close(lookups)
	}()

	return lookups
}
