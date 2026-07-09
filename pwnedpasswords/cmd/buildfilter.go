package main

import (
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/buildfilter"
	"pwnedpasswords/logging"
)

// newBuildFilterCmd builds the "buildfilter" sub-command.
func newBuildFilterCmd() *cobra.Command {
	var use4GB, use8GB, use16GB bool
	var filterPath string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "buildfilter",
		Short: "Build the membership filter from the pwnedcache hashes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var preset buildfilter.Preset
			switch {
			case use4GB:
				preset = buildfilter.Preset4GB
			case use8GB:
				preset = buildfilter.Preset8GB
			case use16GB:
				preset = buildfilter.Preset16GB
			}
			logs, err := logging.Setup(verbose)
			if err != nil {
				return err
			}
			defer logs.LogFile.Close()
			return buildfilter.Run(cmd.Context(), logs, pwnedcachePath, filterPath, preset, progressInterval)
		},
	}
	cmd.Flags().BoolVar(&use4GB, "4GB", false, "4 GiB filter (false positives ~1 in 1,500)")
	cmd.Flags().BoolVar(&use8GB, "8GB", false, "8 GiB filter, suggested (false positives ~1 in 250,000)")
	cmd.Flags().BoolVar(&use16GB, "16GB", false, "16 GiB filter (false positives ~1 in 150 million)")
	cmd.MarkFlagsMutuallyExclusive("4GB", "8GB", "16GB")
	cmd.MarkFlagsOneRequired("4GB", "8GB", "16GB")
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter", "output filter file path")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
	return cmd
}
