// Package progress accumulates the running totals shared by the scanning
// commands and periodically reports them to the console and the run log.
package progress

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

// DefaultInterval is the progress-report interval used when none is given.
const DefaultInterval = 10 * time.Second

// FlushEvery is how often a single-goroutine tally is folded into the shared
// progress, so the reporter sees movement without the loop touching a shared
// atomic every step.
const FlushEvery = 1 << 14

// Progress accumulates the running totals shared by the import and bruteforce
// commands: filter lookups, hashes-table lookups, passwords found, and of those
// the number that added or changed a row in the output database.
type Progress struct {
	filterQueries atomic.Int64
	hashQueries   atomic.Int64
	found         atomic.Int64
	changed       atomic.Int64
	sample        atomic.Pointer[string] // most recent database hit, for display
}

// Tally is a goroutine-local counter set, folded into a Progress periodically
// so the hot path never contends on a shared atomic. Its fields are exported so
// the checker in another package can record into it directly.
type Tally struct {
	FilterQueries int64
	HashQueries   int64
	Found         int64
	Changed       int64
	Sample        string // most recent database hit since the last fold
}

// Add folds t into p and resets t.
func (p *Progress) Add(t *Tally) {
	p.filterQueries.Add(t.FilterQueries)
	p.hashQueries.Add(t.HashQueries)
	p.found.Add(t.Found)
	p.changed.Add(t.Changed)
	if t.Sample != "" {
		s := t.Sample
		p.sample.Store(&s)
	}
	*t = Tally{}
}

// Snapshot reads the current totals as a plain Tally.
func (p *Progress) Snapshot() Tally {
	t := Tally{
		FilterQueries: p.filterQueries.Load(),
		HashQueries:   p.hashQueries.Load(),
		Found:         p.found.Load(),
		Changed:       p.changed.Load(),
	}
	if s := p.sample.Load(); s != nil {
		t.Sample = *s
	}
	return t
}

// Reporter periodically invokes a report function and, once stopped, invokes it
// a final time for a summary. The report function renders and emits one line; its
// argument is "progress" for a periodic tick or "summary" for the final call.
type Reporter struct {
	report func(kind string)
	stop   chan struct{}
	done   chan struct{}
}

// StartReporter calls report("progress") every interval until the returned
// Reporter is stopped with StopAndReport.
func StartReporter(interval time.Duration, report func(kind string)) *Reporter {
	if interval <= 0 {
		interval = DefaultInterval
	}
	r := &Reporter{
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

// StopAndReport halts the ticker and emits a final summary.
func (r *Reporter) StopAndReport() {
	close(r.stop)
	<-r.done
	r.report("summary")
}

// ReportTo returns a report function that logs the shared progress as a friendly
// console line and the matching structured record to the file. filtered selects
// which counter is the candidate total.
func (p *Progress) ReportTo(console, file *slog.Logger, filtered bool) func(kind string) {
	prev := p.Snapshot()
	prevAt := time.Now()
	return func(kind string) {
		c := p.Snapshot()
		now := time.Now()
		file.Info(
			kind,
			"filter_queries", c.FilterQueries,
			"hash_queries", c.HashQueries,
			"found", c.Found,
			"changed", c.Changed,
		)
		console.Info(humanProgress(kind, c, prev, now.Sub(prevAt), filtered))
		prev, prevAt = c, now
	}
}

// humanProgress renders a friendly one-line progress or summary message.
//
// A tick shows the candidate total with its per-second rate, the database-read
// rate, and the write total, then the most recent match. The summary shows
// cumulative totals only.
func humanProgress(kind string, c, prev Tally, since time.Duration, filtered bool) string {
	if kind == "summary" {
		return fmt.Sprintf(
			"Finished: %s candidates > %s db reads > %s writes",
			humanize.Comma(candidateCount(c, filtered)),
			humanize.Comma(c.HashQueries),
			humanize.Comma(c.Changed),
		)
	}
	line := fmt.Sprintf(
		"%s candidates (%s/s) > %s db reads/s > %s writes",
		humanize.Comma(candidateCount(c, filtered)),
		humanize.Comma(perSecond(candidateCount(c, filtered)-candidateCount(prev, filtered), since)),
		humanize.Comma(perSecond(c.HashQueries-prev.HashQueries, since)),
		humanize.Comma(c.Changed),
	)
	if c.Sample != "" {
		line += " > found: " + c.Sample
	}
	return line
}

// candidateCount returns the candidate total: filter lookups when a filter is in
// use, otherwise hashes-table lookups.
func candidateCount(t Tally, filtered bool) int64 {
	if filtered {
		return t.FilterQueries
	}
	return t.HashQueries
}

// perSecond returns the whole-number rate of delta over d, or 0 when d is not positive.
func perSecond(delta int64, d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return int64(float64(delta) / d.Seconds())
}
