package main

import (
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/export"
	"pwnedpasswords/logging"
	"pwnedpasswords/progress"
)

// newExportCmd builds the "export" sub-command.
func newExportCmd() *cobra.Command {
	var top int
	var format string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write the most-breached passwords as a denylist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return export.Run(cmd.Context(), cmd.OutOrStdout(), logging.NewConsoleLogger(verbose),
				export.Options{
					DBPath:   databasePath,
					Top:      top,
					Format:   format,
					Interval: progressInterval,
				})
		},
	}
	cmd.Flags().IntVarP(&top, "top", "n", 1000, "number of passwords to write (ignored by csv)")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "output format: text, json, or csv (csv is a full dump for 'merge')")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", progress.DefaultInterval,
		"interval between progress reports (csv only)")
	return cmd
}
