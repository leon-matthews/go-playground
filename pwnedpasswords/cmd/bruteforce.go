package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"pwnedpasswords/bruteforce"
	"pwnedpasswords/logging"
)

const bruteforceUsage = "Generate candidate passwords in order and record breach matches"

// bruteforceFlags holds the raw flag values before the checkpoint is merged in.
type bruteforceFlags struct {
	level              int
	resume             bool
	resumeFrom         string
	filterPath         string
	workers            int
	progressInterval   time.Duration
	checkpointInterval time.Duration
}

// newBruteforceCmd builds the "bruteforce" sub-command.
func newBruteforceCmd() *cobra.Command {
	var f bruteforceFlags
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
			opts, err := resolveBruteforceOptions(cmd, f)
			if err != nil {
				return err
			}
			return bruteforce.Run(cmd.Context(), logs, opts)
		},
	}
	cmd.Flags().IntVarP(&f.level, "alphabet", "a", 0,
		"character set: 0=digits, 1=lowercase, 2=+space+digits, 3=+uppercase, 4=+symbols")
	cmd.Flags().BoolVar(&f.resume, "resume", false,
		"continue the previous run from its saved checkpoint")
	cmd.Flags().StringVar(&f.resumeFrom, "resume-from", "",
		"resume from this pattern (as logged when interrupted)")
	cmd.Flags().StringVar(&f.filterPath, "filter", "pwnedpasswords.filter",
		"membership filter used to skip database lookups")
	cmd.Flags().IntVarP(&f.workers, "workers", "w", 0,
		"number of parallel workers (default: number of CPUs)")
	cmd.Flags().DurationVarP(&f.progressInterval, "progress", "p", 10*time.Second,
		"interval between progress reports")
	cmd.Flags().DurationVar(&f.checkpointInterval, "checkpoint", time.Hour,
		"interval between resume-checkpoint writes")
	cmd.MarkFlagsMutuallyExclusive("resume", "resume-from")
	cmd.MarkFlagsMutuallyExclusive("resume", "alphabet")
	return cmd
}

// resolveBruteforceOptions builds the run options. With --resume it merges the
// saved checkpoint, where an explicit flag overrides the file, which in turn
// overrides the built-in default.
func resolveBruteforceOptions(cmd *cobra.Command, f bruteforceFlags) (bruteforce.Options, error) {
	if f.checkpointInterval <= 0 {
		return bruteforce.Options{}, fmt.Errorf("--checkpoint must be positive")
	}
	if !f.resume && !cmd.Flags().Changed("alphabet") {
		return bruteforce.Options{}, fmt.Errorf(`required flag "alphabet" not set`)
	}

	opts := bruteforce.Options{
		DBPath:     databasePath,
		CachePath:  pwnedcachePath,
		FilterPath: f.filterPath,
		Level:      f.level,
		Resume:     f.resumeFrom,
		Workers:    f.workers,
		Progress:   f.progressInterval,
		Checkpoint: f.checkpointInterval,
	}

	if f.resume {
		cp, path, err := bruteforce.LoadCheckpoint()
		switch {
		case errors.Is(err, os.ErrNotExist):
			return bruteforce.Options{}, fmt.Errorf("no resume state found at %s; start a run first", path)
		case err != nil:
			return bruteforce.Options{}, err
		}
		applyCheckpoint(cmd, &opts, cp)
	}

	alphabet, err := bruteforce.AlphabetForLevel(opts.Level)
	if err != nil {
		return bruteforce.Options{}, err
	}
	opts.Alphabet = alphabet
	return opts, nil
}

// applyCheckpoint fills opts from cp, leaving any explicitly-set flag untouched,
// and warns when the run targets a different database than the checkpoint.
func applyCheckpoint(cmd *cobra.Command, opts *bruteforce.Options, cp *bruteforce.Checkpoint) {
	opts.Level = cp.Alphabet
	opts.Resume = cp.Pattern
	if cp.Pattern != "" {
		slog.Info(fmt.Sprintf("resuming from %q", cp.Pattern))
	}

	if !cmd.Flags().Changed("database") {
		opts.DBPath = cp.Database
	}
	if !cmd.Flags().Changed("pwnedcache") {
		opts.CachePath = cp.Cache
	}
	if !cmd.Flags().Changed("filter") {
		opts.FilterPath = cp.Filter
	}
	if !cmd.Flags().Changed("workers") {
		opts.Workers = cp.Workers
	}
	if !cmd.Flags().Changed("progress") {
		if d, err := time.ParseDuration(cp.ProgressInterval); err == nil {
			opts.Progress = d
		}
	}
	if !cmd.Flags().Changed("checkpoint") {
		if d, err := time.ParseDuration(cp.CheckpointInterval); err == nil {
			opts.Checkpoint = d
		}
	}

	if opts.DBPath != cp.Database {
		slog.Warn("resuming against a different database than the checkpoint",
			"checkpoint", cp.Database, "using", opts.DBPath)
	}
}
