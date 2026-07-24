// Command cine-import builds a cine SQLite database from the IMDb dataset files.
//
// Usage:
//
//	cine-import <imdb-data-folder> <output.db>
//
// It builds into a temporary file and renames it into place once the whole
// import succeeds, then prints the elapsed build time.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"local.dev/cine/importer"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <imdb-data-folder> <output.db>\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}

// run imports the dataset files in dir into a new database at out.
func run(dir, out string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	start := time.Now()
	if err := importer.Import(context.Background(), out, dir); err != nil {
		return err
	}
	log.Printf("built %s in %s", out, time.Since(start).Round(time.Millisecond))
	return nil
}
