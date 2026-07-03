package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	databasePath string
	verbose      bool
	quiet        bool

	// logs is set by the root command's PersistentPreRunE, before any
	// sub-command's RunE sees it
	logs logging
)

// newRootCmd builds the pwncache command tree.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pwncache",
		Short:         "Maintain and query a local mirror of the Have I Been Pwned password database",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			var err error
			logs, err = setupLogging(verbose, quiet)
			if err != nil {
				return fmt.Errorf("logging setup: %w", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&databasePath, "database", "d", "pwned.db", "path to the SQLite database")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "debug-level logging")
	cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "warnings and errors only")

	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newCheckCmd())
	return cmd
}
