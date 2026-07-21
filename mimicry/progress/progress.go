// Package progress accumulates the running totals of a scan and periodically reports them.
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

// Progress accumulates the running totals of a scan: files hashed, bytes hashed, and files
// served straight from the cache. Safe for concurrent use; a nil *Progress is a no-op.
type Progress struct {
	hashed      atomic.Int64
	bytesHashed atomic.Int64
	fromCache   atomic.Int64
	sample      atomic.Pointer[string] // most recent file hashed, for display
}

// Snapshot is a plain-value read of a Progress at one instant.
type Snapshot struct {
	Hashed      int64
	BytesHashed int64
	FromCache   int64
	Sample      string
}

// Processed is the total files accounted for: hashed plus served from cache.
func (s Snapshot) Processed() int64 {
	return s.Hashed + s.FromCache
}

// Hashed records one freshly hashed file of the given size.
func (p *Progress) Hashed(path string, size int64) {
	if p == nil {
		return
	}
	p.hashed.Add(1)
	p.bytesHashed.Add(size)
	p.sample.Store(&path)
}

// FromCache records n files served directly from the cache.
func (p *Progress) FromCache(n int64) {
	if p == nil {
		return
	}
	p.fromCache.Add(n)
}

// Snapshot reads the current totals.
func (p *Progress) Snapshot() Snapshot {
	if p == nil {
		return Snapshot{}
	}
	s := Snapshot{
		Hashed:      p.hashed.Load(),
		BytesHashed: p.bytesHashed.Load(),
		FromCache:   p.fromCache.Load(),
	}
	if v := p.sample.Load(); v != nil {
		s.Sample = *v
	}
	return s
}

// Reporter periodically invokes a report function and, once stopped, invokes it a final time
// for a summary. The report function renders and emits one line; its argument is "progress" for
// a periodic tick or "summary" for the final call.
type Reporter struct {
	report func(kind string)
	stop   chan struct{}
	done   chan struct{}
}

// StartReporter calls report("progress") every interval until the returned Reporter is stopped
// with StopAndReport. A non-positive interval falls back to DefaultInterval.
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

// ReportTo returns a report function that logs the scan's running totals as a friendly message
// plus structured attributes. total is the number of files the scan expects to process, for the
// determinate "X/Y" display.
func ReportTo(p *Progress, log *slog.Logger, total int) func(kind string) {
	start := p.Snapshot()
	startAt := time.Now()
	prev, prevAt := start, startAt
	return func(kind string) {
		now := time.Now()
		c := p.Snapshot()
		// A tick rates against the previous tick; the summary averages over the whole run.
		baseline, since := prev, now.Sub(prevAt)
		if kind == "summary" {
			baseline, since = start, now.Sub(startAt)
		}
		log.Info(
			humanProgress(kind, c, baseline, since, total),
			"processed", c.Processed(),
			"total", total,
			"hashed", c.Hashed,
			"bytes_hashed", c.BytesHashed,
			"from_cache", c.FromCache,
		)
		prev, prevAt = c, now
	}
}

// humanProgress renders a friendly one-line progress or summary message.
//
// A tick shows the processed/total count, percentage, files-per-second since the previous tick,
// bytes hashed, and the file currently being hashed. The summary shows the same totals with the
// average rate over the whole run, prefixed with the elapsed time and suffixed with the cache
// hit count.
func humanProgress(kind string, c, baseline Snapshot, since time.Duration, total int) string {
	processed := c.Processed()
	rate := perSecond(processed-baseline.Processed(), since)
	line := fmt.Sprintf("%s/%s files", humanize.Comma(processed), humanize.Comma(int64(total)))
	if total > 0 {
		line += fmt.Sprintf(" (%d%%)", processed*100/int64(total))
	}
	line += fmt.Sprintf(" | %s files/s | %s hashed", humanize.Comma(rate), humanize.IBytes(uint64(c.BytesHashed)))
	if kind == "summary" {
		return fmt.Sprintf("Finished in %s: %s | %s from cache",
			since.Round(time.Second), line, humanize.Comma(c.FromCache))
	}
	if c.Sample != "" {
		line += " | " + c.Sample
	}
	return line
}

// perSecond returns the whole-number rate of delta over d, or 0 when d is not positive.
func perSecond(delta int64, d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return int64(float64(delta) / d.Seconds())
}
