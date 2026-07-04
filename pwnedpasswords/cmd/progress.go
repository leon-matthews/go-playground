package main

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

// Report progress at this interval when none is given.
const defaultProgressInterval = 10 * time.Second

// Fold a single-goroutine tally into the shared progress this often, so the
// reporter sees movement without the loop touching a shared atomic every step.
const flushEvery = 1 << 14

// progress accumulates the running totals shared by the import and bruteforce
// commands: filter lookups, hashes-table lookups, and passwords found.
type progress struct {
	filterQueries atomic.Int64
	hashQueries   atomic.Int64
	found         atomic.Int64
}

// tally is a goroutine-local counter set, folded into a progress periodically
// so the hot path never contends on a shared atomic.
type tally struct {
	filterQueries int64
	hashQueries   int64
	found         int64
}

// add folds t into p and resets t.
func (p *progress) add(t *tally) {
	p.filterQueries.Add(t.filterQueries)
	p.hashQueries.Add(t.hashQueries)
	p.found.Add(t.found)
	*t = tally{}
}

// snapshot reads the current totals as a plain tally.
func (p *progress) snapshot() tally {
	return tally{
		filterQueries: p.filterQueries.Load(),
		hashQueries:   p.hashQueries.Load(),
		found:         p.found.Load(),
	}
}

// reporter periodically logs a progress line and, once stopped, a final summary.
type reporter struct {
	prog    *progress
	console *slog.Logger
	file    *slog.Logger
	stop    chan struct{}
	done    chan struct{}
}

// startReporter begins logging prog every interval until the returned reporter
// is stopped. A friendly line goes to the console and the matching structured
// record to the file.
func startReporter(prog *progress, console, file *slog.Logger, interval time.Duration) *reporter {
	if interval <= 0 {
		interval = defaultProgressInterval
	}
	r := &reporter{
		prog:    prog,
		console: console,
		file:    file,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	go func() {
		defer close(r.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-r.stop:
				return
			case <-ticker.C:
				r.report("progress")
			}
		}
	}()
	return r
}

// stopAndReport halts the ticker and logs a final summary of the whole run.
func (r *reporter) stopAndReport() {
	close(r.stop)
	<-r.done
	r.report("summary")
}

// report writes one progress or summary line to the console and the matching
// structured record to the file.
func (r *reporter) report(kind string) {
	c := r.prog.snapshot()
	r.file.Info(
		kind,
		"filter_queries", c.filterQueries,
		"hash_queries", c.hashQueries,
		"found", c.found,
	)
	r.console.Info(humanProgress(kind, c))
}

// humanProgress renders a friendly one-line progress or summary message.
func humanProgress(kind string, c tally) string {
	verb := "Progress"
	if kind == "summary" {
		verb = "Finished"
	}
	return fmt.Sprintf(
		"%s: %s filter queries, %s hashes-table queries, %s passwords found",
		verb,
		humanize.Comma(c.filterQueries),
		humanize.Comma(c.hashQueries),
		humanize.Comma(c.found),
	)
}
