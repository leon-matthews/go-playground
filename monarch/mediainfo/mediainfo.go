// Package mediainfo runs CLI command and parses its output
package mediainfo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Binary is the name of the CLI program used
const Binary = "mediainfo"

// Media collects the most interesting fields from mediainfo's extensive output
type Media struct {
	Name          string
	Size          int
	Format        string
	Bitrate       int
	Duration      time.Duration
	Height        int
	Width         int
	AudioBitrate  int
	AudioChannels int
	AudioFormat   string
	VideoBitrate  int
	VideoFormat   string
}

// Info attempts to read metadata for the given media file
func Info(name string) (*Media, error) {
	// Check path
	if _, err := os.Stat(name); err != nil {
		return nil, err
	}

	// Run command
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	out, err := run(ctx, "--Output=JSON", name)
	if err != nil {
		return nil, err
	}

	info, err := extractInfo(name, out)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// Version runs the binary and returns its version string
// Intended to be used to check that the binary can be found and run successfully.
func Version() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	out, err := run(ctx, "--version")
	if err != nil {
		return "", err
	}

	version, err := extractVersion(out)
	if err != nil {
		return "", err
	}
	return version, nil
}

// output wraps the structure from mediainfo's JSON output
type output struct {
	Media struct {
		Track []track
	}
}

// jsonFloat handles converting quoted floating-point numbers
type jsonFloat float64

func (f jsonFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

func (f *jsonFloat) UnmarshalJSON(raw []byte) error {
	str := string(raw)
	str = strings.Trim(str, `"`)
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	*f = jsonFloat(value)
	return nil
}

// jsonInt handles converting quoted integers
type jsonInt int

func (i jsonInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(i))
}

func (i *jsonInt) UnmarshalJSON(raw []byte) error {
	str := string(raw)
	str = strings.Trim(str, `"`)
	value, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*i = jsonInt(value)
	return nil
}

// track is where the details are found in the JSON output
// We expect a least three tracks per video file: General, Video, and Audio
type track struct {
	Type           string `json:"@type"`
	BitRate        jsonInt
	Channels       jsonInt
	Duration       jsonFloat
	FileSize       jsonInt
	Format         string
	OverallBitRate jsonInt
	Width          jsonInt
	Height         jsonInt
}

// extractInfo unmarshalls command's JSON output and extracts fields of interest
func extractInfo(name string, raw []byte) (*Media, error) {
	// Parse output
	data := &output{}
	if err := json.Unmarshal(raw, data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	info := Media{
		Name: name,
	}

	// Combine data from different track types
	var numVideo, numAudio int
	for idx, t := range data.Media.Track {
		switch t.Type {
		case "Audio":
			info.AudioBitrate = int(t.BitRate)
			info.AudioChannels = int(t.Channels)
			info.AudioFormat = t.Format
			numAudio++
		case "General":
			info.Bitrate = int(t.OverallBitRate)
			info.Duration = time.Duration(float64(t.Duration) * float64(time.Second))
			info.Format = t.Format
			info.Size = int(t.FileSize)
		case "Video":
			info.Height = int(t.Height)
			info.Width = int(t.Width)
			info.VideoBitrate = int(t.BitRate)
			info.VideoFormat = t.Format
			numVideo++
		default:
			return nil, fmt.Errorf("unexpected track #%v %q in %s", idx, t.Type, name)
		}
	}
	if numVideo != 1 {
		return nil, fmt.Errorf("expected one video track in %q, found %v", name, numVideo)
	}
	if numAudio != 1 {
		return nil, fmt.Errorf("expected one audio track in %q, found %v", name, numAudio)
	}

	return &info, nil
}

// extractVersion pulls out just the version from command's output
func extractVersion(out []byte) (string, error) {
	start := bytes.LastIndex(out, []byte("v"))
	if start < 0 {
		return "", fmt.Errorf("version not found: %q", out)
	}
	version := string(out[start:])
	return version, nil
}

// run binary with given args and return its stdout
func run(ctx context.Context, args ...string) ([]byte, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, Binary, args...)
	log.Debug(cmd.String())

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("command timed out after %v", time.Since(start))
	}

	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("%s binary not found, please see installation instructions", Binary)
		}

		return nil, fmt.Errorf("mediainfo: %w: %s", err, out)
	}
	log.Debug("Command completed", "elapsed", time.Since(start))
	return out, nil
}
