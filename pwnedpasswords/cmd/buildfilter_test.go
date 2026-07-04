package main

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildProgressReport(t *testing.T) {
	logTo := func(buf *bytes.Buffer) *slog.Logger {
		return slog.New(slog.NewTextHandler(buf, nil))
	}

	t.Run("progress line shows the humanized count and rate", func(t *testing.T) {
		var console, file bytes.Buffer
		prog := &buildProgress{start: time.Now().Add(-10 * time.Second)}
		prog.added.Store(1_234_567)
		prog.reportTo(logTo(&console), logTo(&file))("progress")

		assert.Regexp(t, `scanned 1,234,567 hashes \([\d,]+/s\)`, console.String())
		assert.Contains(t, file.String(), "added=1234567")
	})

	t.Run("summary line adds the elapsed time", func(t *testing.T) {
		var console, file bytes.Buffer
		prog := &buildProgress{start: time.Now().Add(-10 * time.Second)}
		prog.added.Store(2_000_000)
		prog.reportTo(logTo(&console), logTo(&file))("summary")

		assert.Regexp(t, `scanned 2,000,000 hashes in \d+s \([\d,]+/s\)`, console.String())
	})
}
