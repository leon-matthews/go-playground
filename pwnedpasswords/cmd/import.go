package main

import (
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/logging"
	"pwnedpasswords/wordlist"
)

// newImportCmd builds the "import" sub-command.
func newImportCmd() *cobra.Command {
	var filterPath string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "import <wordlist>...",
		Short: "Import word lists, recording passwords found in the breach corpus",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logs, err := logging.Setup(verbose)
			if err != nil {
				return err
			}
			defer logs.LogFile.Close()
			return wordlist.Run(cmd.Context(), logs, wordlist.Options{
				DBPath:     databasePath,
				CachePath:  pwnedcachePath,
				FilterPath: filterPath,
				Progress:   progressInterval,
				Files:      args,
			})
		},
	}
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter",
		"membership filter used to skip database lookups")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
	return cmd
}
