// Package logging sets up the dual console-and-file loggers the commands share.
package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	charmlog "github.com/charmbracelet/log"
)

// Structured JSONL run log, truncated on each run.
const logPath = "cine.log"

// Logging holds the loggers built by Setup, one per destination a message can
// want: progress goes to All, a short warning to Console, and the full detail
// behind it to File.
type Logging struct {
	All     *slog.Logger // both destinations
	Console *slog.Logger // friendly, human-readable console output
	File    *slog.Logger // structured JSONL records, down to debug level
	LogFile *os.File     // backing file, for the caller to close
}

// Setup builds the loggers, truncating cine.log.
//
// The console carries info and above, so it stays short enough to read while a
// build runs; the file carries debug as well, so the detail behind a summary is
// always recorded somewhere. It also installs the fan-out as the slog default,
// so a log written through the package-level functions reaches both.
func Setup() (Logging, error) {
	logFile, err := os.Create(logPath)
	if err != nil {
		return Logging{}, fmt.Errorf("creating %s: %w", logPath, err)
	}

	console := newConsoleHandler()
	file := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug})
	all := fanout{console, file}
	slog.SetDefault(slog.New(all))

	return Logging{
		All:     slog.New(all),
		Console: slog.New(console),
		File:    slog.New(file),
		LogFile: logFile,
	}, nil
}

// Discard returns loggers that write nowhere, for tests.
func Discard() Logging {
	discard := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return Logging{All: discard, Console: discard, File: discard}
}

// newConsoleHandler builds the colourised charm handler on stderr.
func newConsoleHandler() *charmlog.Logger {
	return charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		Level:           charmlog.InfoLevel,
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
	})
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
