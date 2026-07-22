package main

import (
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/logging"
	"pwnedpasswords/wordlist"
)

// newWordlistCmd builds the "wordlist" sub-command.
func newWordlistCmd() *cobra.Command {
	var filterPath string
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "wordlist <file>...",
		Short: "Read candidates from given word-list",
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
