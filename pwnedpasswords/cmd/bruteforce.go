package main

import (
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/bruteforce"
	"pwnedpasswords/logging"
)

const bruteforceUsage = "Generate candidate passwords in order and record breach matches"

// newBruteforceCmd builds the "bruteforce" sub-command.
func newBruteforceCmd() *cobra.Command {
	var level int
	var resume string
	var filterPath string
	var workers int
	var progressInterval time.Duration
	cmd := &cobra.Command{
		Use:   "bruteforce",
		Short: bruteforceUsage,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logs, err := logging.Setup(verbose)
			if err != nil {
				return err
			}
			defer logs.LogFile.Close()
			alphabet, err := bruteforce.AlphabetForLevel(level)
			if err != nil {
				return err
			}
			return bruteforce.Run(cmd.Context(), logs, bruteforce.Options{
				DBPath:     databasePath,
				CachePath:  pwnedcachePath,
				FilterPath: filterPath,
				Alphabet:   alphabet,
				Resume:     resume,
				Workers:    workers,
				Progress:   progressInterval,
			})
		},
	}
	cmd.Flags().IntVarP(&level, "alphabet", "a", 4,
		"character set: 1=lowercase, 2=+space+digits, 3=+uppercase, 4=+symbols")
	cmd.Flags().StringVar(&resume, "resume", "",
		"resume from this pattern (as logged when interrupted)")
	cmd.Flags().StringVar(&filterPath, "filter", "pwnedpasswords.filter",
		"membership filter used to skip database lookups")
	cmd.Flags().IntVarP(&workers, "workers", "w", 0,
		"number of parallel workers (default: number of CPUs)")
	cmd.Flags().DurationVarP(&progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
	return cmd
}
