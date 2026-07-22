package bruteforce

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"pwnedpasswords/progress"
)

// checkpointVersion is the schema version stamped into the resume file.
const checkpointVersion = 1

// resumeFileName is the state file recording where a run can be resumed.
const resumeFileName = "resume.json"

// Checkpoint is the persisted state a run needs to continue where it left off.
type Checkpoint struct {
	Version            int    `json:"version"`
	Alphabet           int    `json:"alphabet"`
	Pattern            string `json:"pattern"`
	Database           string `json:"database"`
	Cache              string `json:"cache"`
	Filter             string `json:"filter"`
	Workers            int    `json:"workers"`
	ProgressInterval   string `json:"progress_interval"`
	CheckpointInterval string `json:"checkpoint_interval"`
	Updated            string `json:"updated"`
}

// resumeFilePath returns the resume-state path under $XDG_STATE_HOME (falling
// back to ~/.local/state), creating the enclosing directory if needed.
func resumeFilePath() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	dir := filepath.Join(base, "pwnedpasswords")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, resumeFileName), nil
}

// LoadCheckpoint reads the resume state left by a previous run.
//
// The returned path is reported even on error so callers can name it. A missing
// file surfaces as a wrapped os.ErrNotExist.
func LoadCheckpoint() (cp *Checkpoint, path string, err error) {
	path, err = resumeFilePath()
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}
	cp = &Checkpoint{}
	if err := json.Unmarshal(data, cp); err != nil {
		return nil, path, fmt.Errorf("reading resume file %q: %w", path, err)
	}
	return cp, path, nil
}

// save atomically writes the checkpoint via a temp file and rename, so a crash
// mid-write cannot corrupt the resume state.
func (c *Checkpoint) save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// snapshotFunc returns the current resume pattern, and false when there is no
// safe position to record right now (e.g. between candidate lengths).
type snapshotFunc func() (pattern string, ok bool)

// checkpointer periodically persists a run's resume position to disk. Its
// template carries the static run settings; each write stamps in the latest
// pattern and time.
type checkpointer struct {
	path     string
	interval time.Duration
	template Checkpoint
}

// start writes an initial checkpoint and then keeps writing on every interval
// tick. The returned reporter writes a final checkpoint when stopped.
func (c *checkpointer) start(snapshot snapshotFunc) *progress.Reporter {
	c.write(snapshot)
	return progress.StartReporter(c.interval, func(string) { c.write(snapshot) })
}

// write records the current snapshot, skipping the write when no safe position
// is available.
func (c *checkpointer) write(snapshot snapshotFunc) {
	if pattern, ok := snapshot(); ok {
		c.writePattern(pattern)
	}
}

// writePattern persists pattern as the resume position. A write failure is
// logged rather than fatal, so a full disk cannot abort a long run.
func (c *checkpointer) writePattern(pattern string) {
	cp := c.template
	cp.Pattern = pattern
	cp.Updated = time.Now().UTC().Format(time.RFC3339)
	if err := cp.save(c.path); err != nil {
		slog.Warn("could not write resume checkpoint", "path", c.path, "err", err)
	}
}
