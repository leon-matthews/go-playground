package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"pwnedpasswords/database"
	"pwnedpasswords/filter"
)

// Log build progress after this many hashes have been scanned.
const buildProgressInterval = 100_000_000

// filterPreset pairs a filter size with the probe count that minimises its
// false-positive rate for the ~2 billion hash pwnedcache corpus. If that corpus
// grows substantially, retune the probe counts against the new element count.
type filterPreset struct {
	blocks uint64
	probes int
}

var (
	preset4GB  = filterPreset{filter.BlocksForBytes(4 << 30), 10}
	preset8GB  = filterPreset{filter.BlocksForBytes(8 << 30), 16}
	preset16GB = filterPreset{filter.BlocksForBytes(16 << 30), 21}
)

// newBuildFilterCmd builds the "buildfilter" sub-command.
func newBuildFilterCmd() *cobra.Command {
	var use4GB, use8GB, use16GB bool
	var filterPath string
	cmd := &cobra.Command{
		Use:   "buildfilter",
		Short: "Build the membership filter from the pwnedcache hashes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			preset := preset8GB
			switch {
			case use4GB:
				preset = preset4GB
			case use8GB:
				preset = preset8GB
			case use16GB:
				preset = preset16GB
			}
			logger := newLogger(verbose)
			return runBuildFilter(cmd.Context(), logger, pwnedcachePath, filterPath, preset)
		},
	}
	cmd.Flags().BoolVar(&use4GB, "4GB", false, "4 GiB filter (false positives ~1 in 1,500)")
	cmd.Flags().BoolVar(&use8GB, "8GB", false, "8 GiB filter, the default (false positives ~1 in 270,000)")
	cmd.Flags().BoolVar(&use16GB, "16GB", false, "16 GiB filter (false positives ~1 in 175 million)")
	cmd.MarkFlagsMutuallyExclusive("4GB", "8GB", "16GB")
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter", "output filter file path")
	return cmd
}

// runBuildFilter scans every hash in the pwnedcache database into a split-block
// Bloom filter sized by preset and writes it to disk. It refuses to overwrite an
// existing filter file.
func runBuildFilter(ctx context.Context, logger *charmlog.Logger, cachePath, filterPath string, preset filterPreset) error {
	if _, err := os.Stat(filterPath); err == nil {
		return fmt.Errorf("filter %q already exists; remove it to rebuild", filterPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	_, cacheDB, err := database.OpenCache(ctx, cachePath)
	if err != nil {
		return err
	}
	defer cacheDB.Close()

	built, err := filter.New(preset.blocks, preset.probes)
	if err != nil {
		return err
	}
	logger.Info("building filter",
		"blocks", preset.blocks,
		"probes", preset.probes,
		"size_gib", float64(preset.blocks*64)/(1<<30))

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
		"bits_per_element", float64(preset.blocks*512)/float64(count),
		"path", filterPath)
	if err := built.Write(filterPath, cachePath); err != nil {
		return err
	}
	logger.Info("filter complete", "elements", count)
	return nil
}
