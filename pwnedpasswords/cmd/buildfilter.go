package main

import (
	"context"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/filter"
)

// Log build progress after this many hashes have been scanned.
const buildProgressInterval = 100_000_000

// newBuildFilterCmd builds the "buildfilter" sub-command.
func newBuildFilterCmd() *cobra.Command {
	var sizeGiB float64
	var filterPath string
	cmd := &cobra.Command{
		Use:   "buildfilter",
		Short: "Build the membership filter from the pwnedcache hashes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := newLogger(verbose, quiet)
			return runBuildFilter(cmd.Context(), logger, pwnedcachePath, filterPath, sizeGiB)
		},
	}
	cmd.Flags().Float64VarP(&sizeGiB, "size", "s", 16,
		"target filter size in GiB (rounded down to a power-of-two block count)")
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter", "output filter file path")
	return cmd
}

// runBuildFilter scans every hash in the pwnedcache database into a split-block
// Bloom filter and writes it to disk.
func runBuildFilter(ctx context.Context, logger *charmlog.Logger, cachePath, filterPath string, sizeGiB float64) error {
	_, cacheDB, err := database.OpenCache(ctx, cachePath)
	if err != nil {
		return err
	}
	defer cacheDB.Close()

	numBlocks := filter.BlocksForBytes(uint64(sizeGiB * (1 << 30)))
	built, err := filter.New(numBlocks)
	if err != nil {
		return err
	}
	logger.Info("building filter", "blocks", numBlocks, "size_gib", float64(numBlocks*64)/(1<<30))

	rows, err := cacheDB.QueryContext(ctx, "SELECT hash FROM hashes")
	if err != nil {
		return err
	}
	defer rows.Close()

	var count uint64
	for rows.Next() {
		var hash []byte
		if err := rows.Scan(&hash); err != nil {
			return err
		}
		built.Add(hash)
		count++
		if count%buildProgressInterval == 0 {
			logger.Info("scanning hashes", "added", count)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	built.Elements = count

	logger.Info("writing filter",
		"elements", count,
		"bits_per_element", float64(numBlocks*512)/float64(count),
		"path", filterPath)
	if err := built.Write(filterPath, cachePath); err != nil {
		return err
	}
	logger.Info("filter complete", "elements", count)
	return nil
}
