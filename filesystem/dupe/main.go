// Package main implements a File Duplicate Scanner.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	flag "github.com/spf13/pflag"
)

var (
	verbose = flag.BoolP("verbose", "v", false, "verbose output (debug-level logging)")
	quiet   = flag.BoolP("quiet", "q", false, "quiet output (warnings and errors only)")
	minSize = flag.Int64P("min-size", "m", 1024, "ignore duplicates smaller than this many bytes")
	jobs    = flag.IntP("jobs", "j", runtime.NumCPU(), "number of concurrent worker goroutines")
)

// setupLogging installs a slog default logger that writes to stderr at the
// given level. The time attribute is stripped because this CLI is short-lived
// and the per-line timestamp is noise rather than signal.
func setupLogging(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
}

func main() {
	flag.Parse()

	level := slog.LevelInfo
	switch {
	case *verbose:
		level = slog.LevelDebug
	case *quiet:
		level = slog.LevelWarn
	}
	setupLogging(level)

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: filescan [-v|-q] [-m bytes] [-j N] FOLDER(S)...")
		os.Exit(1)
	}

	if *jobs < 1 {
		slog.Warn("invalid --jobs value; clamping to 1", "value", *jobs)
		*jobs = 1
	}

	roots := flag.Args()
	fmt.Printf("Scanning: %s\n", strings.Join(roots, ", "))

	paths, err := collectRoots(roots...)
	fmt.Printf("Found %d files\n", len(paths))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(paths) == 0 {
		fmt.Println("No files found.")
		return
	}

	cacheFile, err := cachePath()
	if err != nil {
		slog.Warn("cache disabled: cannot resolve path", "err", err)
	}
	cache, err := openCache(cacheFile)
	if err != nil {
		slog.Warn("cache disabled", "path", cacheFile, "err", err)
	}
	defer cache.Close()

	files := processFiles(paths, cache, *jobs)

	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		seen[p] = struct{}{}
	}
	absRoots := make([]string, 0, len(roots))
	for _, r := range roots {
		a, absErr := filepath.Abs(r)
		if absErr != nil {
			slog.Warn("cannot resolve absolute root for cache sweep", "root", r, "err", absErr)
			continue
		}
		absRoots = append(absRoots, a)
	}
	if err := cache.Sweep(seen, absRoots); err != nil {
		slog.Warn("cache sweep failed", "err", err)
	}

	analyse(files, *minSize)
}
