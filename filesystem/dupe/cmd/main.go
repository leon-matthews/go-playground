// Package main implements a File Duplicate Scanner.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	charminglog "github.com/charmbracelet/log"
	flag "github.com/spf13/pflag"

	"local.dev/monarch"
)

var (
	verbose = flag.BoolP("verbose", "v", false, "verbose output (debug-level logging)")
	quiet   = flag.BoolP("quiet", "q", false, "quiet output (warnings and errors only)")
	minSize = flag.Int64P("min-size", "m", 1024, "ignore duplicates smaller than this many bytes")
	jobs    = flag.IntP("jobs", "j", runtime.NumCPU(), "number of concurrent worker goroutines")
	force   = flag.BoolP("force", "f", false, "stat every file, ignoring the folder-mtime cache")
)

// cachePath returns the absolute path of the persistent hash cache.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dupe", "cache.db"), nil
}

// setupLogger returns a logger that writes pretty output to stderr (level-filtered by `level`)
// and JSON to logFilePath (always at Debug). If the log file can't be opened the returned file
// is nil and a warn is emitted via the console handler; logging still works.
func setupLogger(level slog.Level, logFilePath string) (*slog.Logger, *os.File) {
	console := charminglog.NewWithOptions(os.Stderr, charminglog.Options{
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
	return slog.New(MultiHandler{console, file}), f
}

// slogToCharm maps slog levels onto charmbracelet/log's level enum.
func slogToCharm(level slog.Level) charminglog.Level {
	switch level {
	case slog.LevelDebug:
		return charminglog.DebugLevel
	case slog.LevelWarn:
		return charminglog.WarnLevel
	case slog.LevelError:
		return charminglog.ErrorLevel
	default:
		return charminglog.InfoLevel
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: dupe [-v|-q] [-m bytes] [-j N] [-f] FOLDER(S)...")
		os.Exit(1)
	}

	level := slog.LevelInfo
	switch {
	case *verbose:
		level = slog.LevelDebug
	case *quiet:
		level = slog.LevelWarn
	}

	cacheFile, err := cachePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve cache path: %v\n", err)
		os.Exit(1)
	}
	log, logHandle := setupLogger(level, filepath.Join(filepath.Dir(cacheFile), "dupe.log"))
	if logHandle != nil {
		defer logHandle.Close()
	}

	if *jobs < 1 {
		log.Warn("invalid --jobs value; clamping to 1", "value", *jobs)
		*jobs = 1
	}

	// Collect paths under given roots
	roots := flag.Args()
	log.Info("scanning", "roots", roots)
	collector := monarch.NewCollector(log)
	if err := collector.Walk(roots...); err != nil {
		log.Error("collect roots failed", "err", err)
		os.Exit(1)
	}
	if collector.TotalFiles() == 0 {
		log.Warn("no files found")
		return
	}

	cache, err := monarch.OpenCache(cacheFile, log)
	if err != nil {
		log.Warn("cache disabled", "path", cacheFile, "err", err)
	}
	defer cache.Close()

	scanner := monarch.NewScanner(cache, *jobs, log, *force)
	files := scanner.Process(collector.Folders)

	if err := cache.Sweep(collector.Folders, collector.AbsRoots); err != nil {
		log.Warn("cache sweep failed", "err", err)
	}

	printReport(os.Stdout, files, *minSize)
}
