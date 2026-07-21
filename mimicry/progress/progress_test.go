package progress

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	t.Run("counts hashed and cached files", func(t *testing.T) {
		var p Progress
		p.Hashed("/a", 100)
		p.Hashed("/b", 200)
		p.FromCache(3)

		s := p.Snapshot()
		assert.Equal(t, int64(2), s.Hashed)
		assert.Equal(t, int64(300), s.BytesHashed)
		assert.Equal(t, int64(3), s.FromCache)
		assert.Equal(t, int64(5), s.Processed())
		assert.Equal(t, "/b", s.Sample, "sample is the most recent hashed file")
	})

	t.Run("nil progress is a safe no-op", func(t *testing.T) {
		var p *Progress
		assert.NotPanics(t, func() {
			p.Hashed("/a", 1)
			p.FromCache(1)
		})
		assert.Equal(t, Snapshot{}, p.Snapshot())
	})
}

func TestHumanProgress(t *testing.T) {
	// 1,234 of 5,000 processed; 1,000 of those since the previous tick 10s ago -> 100/s.
	tick := Snapshot{Hashed: 200, BytesHashed: 2684354560, FromCache: 1034, Sample: "/data/foo.jpg"}
	prev := Snapshot{Hashed: 34, FromCache: 200} // Processed() == 234
	const (
		total = 5000
		since = 10 * time.Second
	)

	t.Run("tick shows count, percent, rate, bytes, and current file", func(t *testing.T) {
		got := humanProgress("progress", tick, prev, since, total)
		assert.Equal(t, "1,234/5,000 files (24%) | 100 files/s | 2.5 GiB hashed | /data/foo.jpg", got)
	})

	t.Run("tick omits the current file before the first hash", func(t *testing.T) {
		c := tick
		c.Sample = ""
		got := humanProgress("progress", c, prev, since, total)
		assert.Equal(t, "1,234/5,000 files (24%) | 100 files/s | 2.5 GiB hashed", got)
	})

	t.Run("tick shows a zero rate when no time has passed", func(t *testing.T) {
		got := humanProgress("progress", tick, prev, 0, total)
		assert.Equal(t, "1,234/5,000 files (24%) | 0 files/s | 2.5 GiB hashed | /data/foo.jpg", got)
	})

	t.Run("tick omits the percentage when the total is unknown", func(t *testing.T) {
		got := humanProgress("progress", tick, prev, since, 0)
		assert.Equal(t, "1,234/0 files | 100 files/s | 2.5 GiB hashed | /data/foo.jpg", got)
	})

	t.Run("summary shows run totals, average rate, elapsed, and cache hits", func(t *testing.T) {
		run := Snapshot{Hashed: 200, BytesHashed: 2684354560, FromCache: 4800}
		got := humanProgress("summary", run, Snapshot{}, since, total)
		assert.Equal(t, "Finished in 10s: 5,000/5,000 files (100%) | 500 files/s | 2.5 GiB hashed | 4,800 from cache", got)
	})
}
