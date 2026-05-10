// Package main implements a File Duplicate Scanner.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

var (
	verbose = flag.BoolP("verbose", "v", false, "verbose output (debug-level logging)")
	quiet   = flag.BoolP("quiet", "q", false, "quiet output (warnings and errors only)")
	minSize = flag.Int64P("min-size", "m", 1024, "ignore duplicates smaller than this many bytes")
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
		fmt.Fprintln(os.Stderr, "Usage: filescan [-v|-q] [-m bytes] FOLDER(S)...")
		os.Exit(1)
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

	cache := loadCache()
	files := processFiles(paths, cache)
	updateCache(cache, files, paths, roots)
	if err := saveCache(cache); err != nil {
		slog.Warn("failed to save cache", "err", err)
	}

	analyse(files, *minSize)
}
