package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/database/sqlite"
)

// newImportCmd builds the "import" sub-command.
func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <wordlist>...",
		Short: "Import word lists, recording passwords found in the breach corpus",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := newLogger(verbose, quiet)
			return runImport(cmd.Context(), logger, databasePath, pwnedcachePath, args)
		},
	}
}

// runImport reads each word list, hashes every candidate, and records the ones
// present in the pwnedcache database along with their breach counts.
func runImport(ctx context.Context, logger *charmlog.Logger, dbPath, cachePath string, files []string) error {
	writeQueries, writeDB, err := database.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer writeDB.Close()

	cacheQueries, cacheDB, err := database.OpenCache(ctx, cachePath)
	if err != nil {
		return err
	}
	defer cacheDB.Close()

	var totalRead, totalMatched int64
	for _, file := range files {
		read, matched, err := importFile(ctx, writeQueries, cacheQueries, file)
		if err != nil {
			return fmt.Errorf("importing %q: %w", file, err)
		}
		totalRead += read
		totalMatched += matched
		logger.Info("imported word list", "file", file, "read", read, "matched", matched)
	}
	logger.Info("import complete", "read", totalRead, "matched", totalMatched)
	return nil
}

// importFile streams one word list, looking up each candidate password's hash
// and upserting any match into the passwords table.
func importFile(ctx context.Context, write, cache *sqlite.Queries, path string) (read, matched int64, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		password := scanner.Text()
		if password == "" {
			continue
		}
		read++

		found, err := recordCandidate(ctx, write, cache, password)
		if err != nil {
			return read, matched, err
		}
		if found {
			matched++
		}
	}
	if err := scanner.Err(); err != nil {
		return read, matched, err
	}
	return read, matched, nil
}
