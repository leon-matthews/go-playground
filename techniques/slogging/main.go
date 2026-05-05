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
	verbose = pflag.BoolP("verbose", "v", false, "debug-level logging")
	quiet   = pflag.BoolP("quiet", "q", false, "warnings and errors only")
)

// setupLogging installs a slog default logger that writes to stderr.
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
}

// setupJSONLogging installs a slog default logger that writes line-delimited
// JSON to stderr — the format most log aggregators expect.
func setupJSONLogging() {
	level := slog.LevelInfo
	switch {
	case *verbose:
		level = slog.LevelDebug
	case *quiet:
		level = slog.LevelWarn
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

// setupCharmingLogging installs a slog default backed by Charmbracelet's
// log package — colourised, with caller info and a short Kitchen timestamp.
func setupCharmingLogging() {
	level := charminglog.InfoLevel
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
	slog.SetDefault(slog.New(handler))
}

// setupMultiHandler installs a slog default that writes pretty output to
// stderr (level driven by -v/-q) and JSON to a freshly-truncated log file
// (always at Debug). The caller is responsible for closing the file.
// Note that this time, we're not printing the source file, but are logging it.
func setupMultiHandler(path string) (*os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	consoleLevel := charminglog.InfoLevel
	switch {
	case *verbose:
		consoleLevel = charminglog.DebugLevel
	case *quiet:
		consoleLevel = charminglog.WarnLevel
	}
	console := charminglog.NewWithOptions(os.Stderr, charminglog.Options{
		Level:           consoleLevel,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	file := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	slog.SetDefault(slog.New(MultiHandler{console, file}))
	return f, nil
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

	fmt.Println("## slog.JSONHandler with default options")
	setupJSONLogging()
	slogExamples()
	fmt.Println()

	fmt.Println("## charmbraclet.log configured as slog handler")
	setupCharmingLogging()
	slogExamples()
	fmt.Println()

	fmt.Println("## MultiHandler: charm to stderr, JSON to slogging.log")
	f, err := setupMultiHandler("slogging.log")
	if err != nil {
		fmt.Fprintln(os.Stderr, "open log file:", err)
		os.Exit(1)
	}
	defer f.Close()
	slogExamples()
}
