package main

import (
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/export"
	"pwnedpasswords/logging"
	"pwnedpasswords/progress"
)

// newMergeCmd builds the "merge" sub-command.
func newMergeCmd() *cobra.Command {
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "merge <file.csv>...",
		Short: "Merge CSVs of passwords and counts, skipping any already present",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logs, err := logging.Setup(verbose)
			if err != nil {
				return err
			}
			defer logs.LogFile.Close()
			return export.Merge(cmd.Context(), logs, databasePath, args, progressInterval)
		},
	}
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", progress.DefaultInterval,
		"interval between progress reports")
	return cmd
}
