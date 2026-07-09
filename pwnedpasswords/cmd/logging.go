package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	charmlog "github.com/charmbracelet/log"
)

// Structured NDJSON run log, truncated on each run.
const logPath = "pwnedpasswords.log"

// logging holds the loggers built by setupLogging.
type logging struct {
	console *slog.Logger // Friendly, human-readable console output
	file    *slog.Logger // Structured NDJSON records
	logFile *os.File     // Backing file, for the caller to close
}

// newConsoleHandler builds the colourised charm handler on stderr, at debug
// level when verbose. It backs both the full logging fan-out and the
// console-only logger read-only commands use for progress.
func newConsoleHandler(verbose bool) *charmlog.Logger {
	level := charmlog.InfoLevel
	if verbose {
		level = charmlog.DebugLevel
	}
	return charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		Level:           level,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
}

// newConsoleLogger returns a console-only logger on stderr.
//
// Read-only commands such as export report progress through it, so they never
// open or truncate the run log the way a write command does.
func newConsoleLogger(verbose bool) *slog.Logger {
	return slog.New(newConsoleHandler(verbose))
}

// setupLogging installs a fan-out slog default that writes every log to both a
// colourised console handler on stderr and an NDJSON file handler on
// pwnedpasswords.log, truncated each run.
//
// Progress reporting uses the returned loggers to send friendly text to the
// console and the matching structured record to the file. The -v flag raises
// the level to debug.
func setupLogging(verbose bool) (logging, error) {
	fileLevel := slog.LevelInfo
	if verbose {
		fileLevel = slog.LevelDebug
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return logging{}, fmt.Errorf("creating %s: %w", logPath, err)
	}

	console := newConsoleHandler(verbose)
	file := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: fileLevel})

	// Incidental logs go to both; progress is routed explicitly by the reporter
	slog.SetDefault(slog.New(fanout{console, file}))

	return logging{
		console: slog.New(console),
		file:    slog.New(file),
		logFile: logFile,
	}, nil
}

// fanout is an slog.Handler that dispatches each record to every child handler.
type fanout []slog.Handler

func (f fanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f fanout) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range f {
		if h.Enabled(ctx, r.Level) {
			// Clone so a handler that retains the record cannot disturb the others
			errs = append(errs, h.Handle(ctx, r.Clone()))
		}
	}
	return errors.Join(errs...)
}

func (f fanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make(fanout, len(f))
	for i, h := range f {
		next[i] = h.WithAttrs(attrs)
	}
	return next
}

func (f fanout) WithGroup(name string) slog.Handler {
	next := make(fanout, len(f))
	for i, h := range f {
		next[i] = h.WithGroup(name)
	}
	return next
}
