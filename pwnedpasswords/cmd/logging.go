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

// setupLogging installs a fan-out slog default that writes every log to both a
// colourised console handler on stderr and an NDJSON file handler on
// pwnedpasswords.log, truncated each run.
//
// Progress reporting uses the returned loggers to send friendly text to the
// console and the matching structured record to the file. The -v flag raises
// the level to debug.
func setupLogging(verbose bool) (logging, error) {
	consoleLevel, fileLevel := charmlog.InfoLevel, slog.LevelInfo
	if verbose {
		consoleLevel, fileLevel = charmlog.DebugLevel, slog.LevelDebug
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return logging{}, fmt.Errorf("creating %s: %w", logPath, err)
	}

	console := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		Level:           consoleLevel,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
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
