package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
)

// newRootCmd builds the mimicry command tree.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "mimicry",
		Short:         "Scan directory trees for duplicate files",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (debug-level logging)")
	cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output (warnings and errors only)")
	cmd.AddCommand(newScanCmd())
	cmd.AddCommand(newReportCmd())
	return cmd
}

// cachePath returns the absolute path of the persistent hash cache.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mimicry", "cache.db"), nil
}

// logLevel maps the persistent verbose/quiet flags onto an slog.Level.
func logLevel() slog.Level {
	switch {
	case verbose:
		return slog.LevelDebug
	case quiet:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
