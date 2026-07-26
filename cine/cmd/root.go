package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// profile requests a CPU profile of the run.
var profile bool

// stopProfile stops the run's CPU profile, if one was started; nil otherwise.
var stopProfile func()

const profilePath = "cpu.pprof"

// newRootCmd builds the cine command tree.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cine",
		Short:             "Build and query a local SQLite database of the IMDb datasets",
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

	cmd.PersistentFlags().BoolVar(&profile, "profile", false, "write a CPU profile of the run to "+profilePath)

	// Replace the 'help' built-in command with an unnamed stub to remove it from the command list
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.AddCommand(newBuildDatabaseCmd())
	cmd.AddCommand(newReaderBenchmarkCmd())
	return cmd
}

// newLogger builds the console logger the commands report progress to.
func newLogger() *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
	})
}

// requireFolder checks that path is an existing directory, as every command
// takes a folder of dataset files rather than the files themselves.
func requireFolder(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}
