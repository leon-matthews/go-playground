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
// commands: filter lookups, hashes-table lookups, passwords found, and of those
// the number that added or changed a row in the output database.
type progress struct {
	filterQueries atomic.Int64
	hashQueries   atomic.Int64
	found         atomic.Int64
	changed       atomic.Int64
	sample        atomic.Pointer[string] // most recent database hit, for display
}

// tally is a goroutine-local counter set, folded into a progress periodically
// so the hot path never contends on a shared atomic.
type tally struct {
	filterQueries int64
	hashQueries   int64
	found         int64
	changed       int64
	sample        string // most recent database hit since the last fold
}

// add folds t into p and resets t.
func (p *progress) add(t *tally) {
	p.filterQueries.Add(t.filterQueries)
	p.hashQueries.Add(t.hashQueries)
	p.found.Add(t.found)
	p.changed.Add(t.changed)
	if t.sample != "" {
		s := t.sample
		p.sample.Store(&s)
	}
	*t = tally{}
}

// snapshot reads the current totals as a plain tally.
func (p *progress) snapshot() tally {
	t := tally{
		filterQueries: p.filterQueries.Load(),
		hashQueries:   p.hashQueries.Load(),
		found:         p.found.Load(),
		changed:       p.changed.Load(),
	}
	if s := p.sample.Load(); s != nil {
		t.sample = *s
	}
	return t
}

// reporter periodically invokes a report function and, once stopped, invokes it
// a final time for a summary. The report function renders and emits one line; its
// argument is "progress" for a periodic tick or "summary" for the final call.
type reporter struct {
	report func(kind string)
	stop   chan struct{}
	done   chan struct{}
}

// startReporter calls report("progress") every interval until the returned
// reporter is stopped with stopAndReport.
func startReporter(interval time.Duration, report func(kind string)) *reporter {
	if interval <= 0 {
		interval = defaultProgressInterval
	}
	r := &reporter{
		report: report,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
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

// stopAndReport halts the ticker and emits a final summary.
func (r *reporter) stopAndReport() {
	close(r.stop)
	<-r.done
	r.report("summary")
}

// reportTo returns a report function that logs the shared progress as a friendly
// console line and the matching structured record to the file. filtered selects
// the leading count's label.
func (p *progress) reportTo(console, file *slog.Logger, filtered bool) func(kind string) {
	return func(kind string) {
		c := p.snapshot()
		file.Info(
			kind,
			"filter_queries", c.filterQueries,
			"hash_queries", c.hashQueries,
			"found", c.found,
			"changed", c.changed,
		)
		console.Info(humanProgress(kind, c, filtered))
	}
}

// humanProgress renders a friendly one-line progress or summary message. The
// leading count is candidates seen: filter lookups when a filter is in use,
// otherwise candidates processed directly. It is followed by the hashes-table
// reads and row writes, then the most recent matched password as the current
// sample.
func humanProgress(kind string, c tally, filtered bool) string {
	count, label := c.filterQueries, "filtered"
	if !filtered {
		count, label = c.hashQueries, "processed"
	}
	line := fmt.Sprintf(
		"%s %s > %s database reads > %s writes",
		humanize.Comma(count),
		label,
		humanize.Comma(c.hashQueries),
		humanize.Comma(c.changed),
	)
	if c.sample != "" {
		line += " > current: " + c.sample
	}
	if kind == "summary" {
		return "Finished: " + line
	}
	return line
}
