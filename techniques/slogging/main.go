package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	charminglog "github.com/charmbracelet/log"
	"github.com/spf13/pflag"
)

var (
	verbose = pflag.BoolP("verbose", "v", false, "debug-level logging)")
	quiet   = pflag.BoolP("quiet", "q", false, "warnings and errors only")
)

// setupLogging installs a slog default logger that writes to stderr.

// The time attribute is stripped because this CLI is short-lived and the
// per-line timestamp is noise rather than signal.
func setupLogging() {
	// Level is chosen from the -v and -q flags; -v wins if both are set.
	level := slog.LevelInfo
	switch {
	case *verbose:
		level = slog.LevelDebug
	case *quiet:
		level = slog.LevelWarn
	}

	// Set level and drop too-verbose time attribute
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))
	handlerJSON := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})
	slog.SetDefault(slog.New(handlerJSON))
}

// setupCharmingLogging installs a slog default logger that writes to stderr.
func setupCharmingLogging() {
	// Level is chosen from the -v and -q flags; -v wins if both are set.
	var level charminglog.Level
	switch {
	case *verbose:
		level = charminglog.DebugLevel
	case *quiet:
		level = charminglog.WarnLevel
	}

	handler := charminglog.NewWithOptions(os.Stderr, charminglog.Options{
		Level:           level,
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func slogExamples() {
	// Simple examples
	slog.Debug("Debug logging enabled", "cortisol", "low")
	slog.Info("Info message", "cortisol", "medium")
	slog.Warn("Warning message", "cortisol", "high")
	slog.Error("Error message", "cortisol", "EXTREME")

	// Typed, with context
	ctx := context.Background()
	slog.LogAttrs(ctx, slog.LevelDebug, "Just debugging", slog.Int("cortisolLevel", 1))
	slog.LogAttrs(ctx, slog.LevelInfo, "FYI only", slog.Int("cortisolLevel", 2))
	slog.LogAttrs(ctx, slog.LevelWarn, "Watch out!", slog.Int("cortisolLevel", 3))
	slog.LogAttrs(ctx, slog.LevelError, "It done broke", slog.Int("cortisolLevel", 100))
}

func main() {
	pflag.Parse()

	fmt.Println("## slog.TextHandler with timestamps removed")
	setupLogging()
	slogExamples()
	fmt.Println()

	fmt.Println("## charmbraclet.log configured as slog handler")
	setupCharmingLogging()
	slogExamples()
}
