package mediainfo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

const Binary = "mediainfo"

// Version runs the binary and returns its version string
// Intended to be used to check that the binary can be found and run successfully.
func Version() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, Binary, "--version")
	slog.Debug(cmd.String())
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("%s binary not found, please see installation instructions", Binary)
		}
		return "", fmt.Errorf("get version: %w: %s", err, out)
	}
	version, err := extractVersion(string(out))
	if err != nil {
		return "", err
	}
	return version, nil
}

// extractVersion pulls out just the version from command's output
func extractVersion(out string) (string, error) {
	start := strings.LastIndex(out, "v")
	if start < 0 {
		return "", fmt.Errorf("version not found: %q", out)
	}
	return out[start:], nil
}
