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
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Binary is the name of the CLI program used
const Binary = "mediainfo"

const timeout = 10 * time.Second

// ErrTimeout is returned when mediainfo exceeds the per-file timeout.
var ErrTimeout = errors.New("timed out")

// Duration wraps time.Duration with JSON serialization as fractional seconds.
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).Seconds())
}

func (d *Duration) UnmarshalJSON(raw []byte) error {
	var seconds float64
	if err := json.Unmarshal(raw, &seconds); err != nil {
		return err
	}
	*d = Duration(time.Duration(seconds * float64(time.Second)))
	return nil
}

// Media collects the most interesting fields from mediainfo's extensive output
type Media struct {
	Name           string       `json:"name"`
	Size           int          `json:"size"`
	Format         string       `json:"format"`
	OverallBitrate int          `json:"overall_bitrate"`
	Duration       Duration     `json:"duration"`
	Video          []VideoTrack `json:"video"`
	Audio          []AudioTrack `json:"audio"`
	Text           []TextTrack  `json:"text"`
}

// VideoTrack describes a single video stream
type VideoTrack struct {
	Format  string `json:"format"`
	Bitrate int    `json:"bitrate"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// AudioTrack describes a single audio stream
type AudioTrack struct {
	Format   string `json:"format"`
	Bitrate  int    `json:"bitrate"`
	Channels int    `json:"channels"`
}

// TextTrack describes a single subtitle or caption stream
type TextTrack struct {
	Format string `json:"format"`
}

// Info attempts to read metadata for the given media file
func Info(name string) (*Media, error) {
	// Check path
	if _, err := os.Stat(name); err != nil {
		return nil, err
	}

	// Run command
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
	return json.Marshal(float64(f))
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

	// Start output struct
	info := Media{
		Name:  filepath.Base(name),
		Video: []VideoTrack{},
		Audio: []AudioTrack{},
		Text:  []TextTrack{},
	}

	// Collect tracks by type; Text and Menu tracks are ignored.
	var numGeneral int
	for _, t := range data.Media.Track {
		switch t.Type {
		case "General":
			info.OverallBitrate = int(t.OverallBitRate)
			info.Duration = Duration(float64(t.Duration) * float64(time.Second))
			info.Format = t.Format
			info.Size = int(t.FileSize)
			numGeneral++
		case "Video":
			info.Video = append(info.Video, VideoTrack{
				Format:  t.Format,
				Bitrate: int(t.BitRate),
				Width:   int(t.Width),
				Height:  int(t.Height),
			})
		case "Audio":
			info.Audio = append(info.Audio, AudioTrack{
				Format:   t.Format,
				Bitrate:  int(t.BitRate),
				Channels: int(t.Channels),
			})
		case "Text":
			info.Text = append(info.Text, TextTrack{Format: t.Format})
		case "Menu":
			// ignore chapter/menu tracks
		}
	}
	if numGeneral != 1 {
		return nil, fmt.Errorf("expected one General track in %q, found %v", name, numGeneral)
	}
	if len(info.Video) == 0 && len(info.Audio) == 0 {
		return nil, fmt.Errorf("no audio or video tracks in %q", name)
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

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("mediainfo: %w after %v", ErrTimeout, time.Since(start))
	}

	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("%s binary not found, please see installation instructions", Binary)
		}

		return nil, fmt.Errorf("mediainfo: %w: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}
	log.Debug("Command completed", "elapsed", time.Since(start))
	return stdout.Bytes(), nil
}
