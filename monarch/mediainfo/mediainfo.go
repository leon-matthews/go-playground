package mediainfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

const Binary = "mediainfo"

type MediaInfo struct {
	Media struct {
		Track []Track
	}
}

type Track struct {
    Type           string `json:"@type"`
    BitRate        string
    OverallBitRate string
    Duration       string
    Format         string
    Channels       string
    Width          string
    Height         string
}


// Info attepmts to read metadata for the given media file
func Info(name string) ([]Track, error) {
	// Check path
	if _, err := os.Stat(name); err != nil {
		return nil, err
	}

	// Run command
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, Binary, "--Output=JSON", name)
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("%s binary not found, please see installation instructions", Binary)
		}
		return nil, fmt.Errorf("mediainfo: %w: %s", err, out)
	}

	// Parse output
	info := new(MediaInfo)
	if err := json.Unmarshal(out, info); err != nil {
		panic(err)
	}

	return info.Media.Track, nil
}

// Version runs the binary and returns its version string
// Intended to be used to check that the binary can be found and run successfully.
func Version() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, Binary, "--version")
	log.Debug(cmd.String())
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("%s binary not found, please see installation instructions", Binary)
		}
		return "", fmt.Errorf("mediainfo version: %w: %s", err, out)
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
