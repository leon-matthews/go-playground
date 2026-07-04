package main

import (
	"github.com/spf13/cobra"
)

var (
	databasePath   string
	pwnedcachePath string
	verbose        bool
	quiet          bool
)

// newRootCmd builds the pwnedpasswords command tree.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pwnedpasswords",
		Short:         "Build breach-frequency password denylists from word lists",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&databasePath, "database", "d", "pwnedpasswords.db", "path to the output SQLite database")
	cmd.PersistentFlags().StringVarP(&pwnedcachePath, "pwnedcache", "c", "pwnedcache.db", "path to the read-only pwnedcache database")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "debug-level logging")
	cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "warnings and errors only")

	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newBruteforceCmd())
	cmd.AddCommand(newBuildFilterCmd())
	cmd.AddCommand(newExportCmd())
	return cmd
}
