package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwnedpasswords/bruteforce"
)

// bruteforceSubCmd returns the parsed bruteforce sub-command for the given
// command line, so tests can inspect flag resolution without executing the run.
func bruteforceSubCmd(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	root := newRootCmd()
	var sub *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "bruteforce" {
			sub = c
		}
	}
	require.NotNil(t, sub)
	require.NoError(t, sub.ParseFlags(args))
	return sub
}

// writeCheckpoint drops a resume file under a temporary XDG_STATE_HOME.
func writeCheckpoint(t *testing.T, cp bruteforce.Checkpoint) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	path := filepath.Join(dir, "pwnedpasswords", "resume.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	data, err := json.Marshal(cp)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

func TestResolveBruteforceOptions(t *testing.T) {
	stored := bruteforce.Checkpoint{
		Version: 1, Alphabet: 4, Pattern: "abc",
		Database: "stored.db", Cache: "stored-cache.db", Filter: "stored.filter",
		Workers: 8, ProgressInterval: "1m0s", CheckpointInterval: "1h0m0s",
	}

	t.Run("resume restores stored settings and an explicit flag overrides them", func(t *testing.T) {
		writeCheckpoint(t, stored)
		cmd := bruteforceSubCmd(t, "--resume", "-d", "explicit.db")

		opts, err := resolveBruteforceOptions(cmd, bruteforceFlags{
			resume: true, filterPath: "pwnedpasswords.filter",
			progressInterval: 10 * time.Second, checkpointInterval: time.Hour,
		})
		require.NoError(t, err)

		assert.Equal(t, 4, opts.Level)
		assert.Equal(t, "abc", opts.Resume)
		assert.Equal(t, "explicit.db", opts.DBPath, "explicit -d wins over the file")
		assert.Equal(t, "stored.filter", opts.FilterPath, "unset flag comes from the file")
		assert.Equal(t, 8, opts.Workers)
		assert.Equal(t, time.Minute, opts.Progress)
		wantAlphabet, _ := bruteforce.AlphabetForLevel(4)
		assert.Equal(t, wantAlphabet, opts.Alphabet)
	})

	t.Run("missing checkpoint is a friendly error", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		cmd := bruteforceSubCmd(t, "--resume")

		_, err := resolveBruteforceOptions(cmd, bruteforceFlags{resume: true, checkpointInterval: time.Hour})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no resume state found")
	})

	t.Run("alphabet is required without resume", func(t *testing.T) {
		cmd := bruteforceSubCmd(t)
		_, err := resolveBruteforceOptions(cmd, bruteforceFlags{checkpointInterval: time.Hour})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "alphabet")
	})

	t.Run("a fresh run takes the alphabet from the flag", func(t *testing.T) {
		cmd := bruteforceSubCmd(t, "-a", "1")
		opts, err := resolveBruteforceOptions(cmd, bruteforceFlags{level: 1, checkpointInterval: time.Hour})
		require.NoError(t, err)
		assert.Equal(t, 1, opts.Level)
		wantAlphabet, _ := bruteforce.AlphabetForLevel(1)
		assert.Equal(t, wantAlphabet, opts.Alphabet)
	})

	t.Run("a non-positive checkpoint interval is rejected", func(t *testing.T) {
		cmd := bruteforceSubCmd(t, "-a", "0", "--checkpoint", "0s")
		_, err := resolveBruteforceOptions(cmd, bruteforceFlags{level: 0, checkpointInterval: 0})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checkpoint")
	})
}
