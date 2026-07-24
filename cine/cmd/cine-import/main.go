// Command cine-import builds a cine SQLite database from the IMDb dataset files.
//
// Usage:
//
//	cine-import <imdb-data-folder> <output.db>
//
// It builds into a temporary file and renames it into place once the whole
// import succeeds, logging per-file progress and the total build time.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"

	"local.dev/cine/importer"
)

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
	})
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <imdb-data-folder> <output.db>\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2], logger); err != nil {
		logger.Fatal(err)
	}
}

// run imports the dataset files in dir into a new database at out.
func run(dir, out string, logger *log.Logger) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return importer.Import(context.Background(), out, dir, logger)
}
