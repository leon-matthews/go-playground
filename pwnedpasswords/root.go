package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	databasePath   string
	pwnedcachePath string
	profile        bool
	verbose        bool
)

// stopProfile stops the run's CPU profile, if one was started; nil otherwise.
var stopProfile func()

const profilePath = "cpu.pprof"

// newRootCmd builds the pwnedpasswords command tree.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "pwnedpasswords",
		Short:             "Build breach-frequency password denylists from word lists",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if !profile {
				return nil
			}
			stop, err := startProfile(profilePath)
			if err != nil {
				return fmt.Errorf("creating profile: %w", err)
			}
			stopProfile = stop
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&databasePath, "database", "d", "pwnedpasswords.db", "path to the output SQLite database")
	cmd.PersistentFlags().StringVarP(&pwnedcachePath, "pwnedcache", "c", "pwnedcache.db", "path to the read-only pwnedcache database")
	cmd.PersistentFlags().BoolVar(&profile, "profile", false, "write a CPU profile of the run to "+profilePath)
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "debug-level logging")

	// Replace 'help` built-in command with an unnamed stub to remove it from the command list
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.AddCommand(newWordlistCmd())
	cmd.AddCommand(newBruteforceCmd())
	cmd.AddCommand(newBuildFilterCmd())
	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newMergeCmd())
	return cmd
}
