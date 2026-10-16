package bruteforce

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckpoint(t *testing.T) {
	t.Run("resumeFilePath honours XDG_STATE_HOME", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_STATE_HOME", dir)

		path, err := resumeFilePath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "pwnedpasswords", "resume.json"), path)
		assert.DirExists(t, filepath.Join(dir, "pwnedpasswords"))
	})

	t.Run("save and LoadCheckpoint round-trip", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		path, err := resumeFilePath()
		require.NoError(t, err)

		want := &Checkpoint{
			Version:            checkpointVersion,
			Alphabet:           4,
			Pattern:            "aPEOX?it",
			Database:           "pwnedpasswords.db",
			Cache:              "pwnedcache.db",
			Filter:             "16GB.filter",
			Workers:            8,
			ProgressInterval:   "1m0s",
			CheckpointInterval: "1h0m0s",
			Updated:            "2026-07-23T10:00:00Z",
		}
		require.NoError(t, want.save(path))

		got, gotPath, err := LoadCheckpoint()
		require.NoError(t, err)
		assert.Equal(t, path, gotPath)
		assert.Equal(t, want, got)
	})

	t.Run("save leaves no temp file behind", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		path, err := resumeFilePath()
		require.NoError(t, err)

		require.NoError(t, (&Checkpoint{Version: checkpointVersion}).save(path))
		_, err = os.Stat(path + ".tmp")
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("LoadCheckpoint reports a missing file as os.ErrNotExist", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		_, _, err := LoadCheckpoint()
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}
