// Package logging builds the dual console-and-file logger the CLI uses.
package logging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	charmlog "github.com/charmbracelet/log"
)

// Setup returns a logger that writes pretty output to stderr (level-filtered by level)
// and JSON to logFilePath (always at Debug).
//
// If the log file can't be opened the returned file is nil and a warn is emitted via the
// console handler; logging still works.
func Setup(level slog.Level, logFilePath string) (*slog.Logger, *os.File) {
	console := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		Level:           slogToCharm(level),
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	f, err := os.Create(logFilePath)
	if err != nil {
		consoleOnly := slog.New(console)
		consoleOnly.Warn("could not open log file; console-only logging", "path", logFilePath, "err", err)
		return consoleOnly, nil
	}

	file := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(multiHandler{console, file}), f
}

// slogToCharm maps slog levels onto charmbracelet/log's level enum.
func slogToCharm(level slog.Level) charmlog.Level {
	switch level {
	case slog.LevelDebug:
		return charmlog.DebugLevel
	case slog.LevelWarn:
		return charmlog.WarnLevel
	case slog.LevelError:
		return charmlog.ErrorLevel
	default:
		return charmlog.InfoLevel
	}
}

// multiHandler fans a single slog.Record out to several sub-handlers, each of which keeps its
// own level filter and formatting.
type multiHandler []slog.Handler

func (h multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, sub := range h {
		if sub.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, sub := range h {
		if !sub.Enabled(ctx, r.Level) {
			continue
		}
		// Clone per sub: handlers are free to mutate the record via AddAttrs.
		if err := sub.Handle(ctx, r.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make(multiHandler, len(h))
	for i, sub := range h {
		out[i] = sub.WithAttrs(attrs)
	}
	return out
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	out := make(multiHandler, len(h))
	for i, sub := range h {
		out[i] = sub.WithGroup(name)
	}
	return out
}
