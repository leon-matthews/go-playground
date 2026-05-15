// List media files in a folder, sorted by bitrate descending
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"

	"golang.org/x/term"

	"go-playground/monarch/mediainfo"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <folder>\n", os.Args[0])
		os.Exit(1)
	}
	folder := os.Args[1]

	if _, err := mediainfo.Version(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	media, err := scan(folder)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slices.SortFunc(media, func(a, b *mediainfo.Media) int {
		return b.OverallBitrate - a.OverallBitrate
	})

	width := terminalWidth()
	if len(media) > 0 {
		printHeader(width)
	}
	for _, m := range media {
		printLine(m, width)
	}
}

// terminalWidth returns the current terminal width, or 80 as a fallback.
func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// truncate cuts s to at most width characters, appending "..." if truncated.
func truncate(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s
}

func printHeader(width int) {
	line := fmt.Sprintf("%13s  %8s  %9s  %-6s %-6s  %4s  %s",
		"Bitrate", "Duration", "Size", "Video", "Audio", "Text", "Name")
	fmt.Println(truncate(line, width))
}

// scan reads the folder (non-recursive) and returns media info for each file
// that mediainfo can parse; unparseable files are skipped.
func scan(folder string) ([]*mediainfo.Media, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	var media []*mediainfo.Media
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(folder, e.Name())
		info, err := mediainfo.Info(path)
		if err != nil {
			slog.Debug("skipping", "path", path, "err", err)
			continue
		}
		media = append(media, info)
	}
	return media, nil
}

func printLine(m *mediainfo.Media, width int) {
	var dimensions, videoFormat, audioFormat string
	if len(m.Video) > 0 {
		v := m.Video[0]
		dimensions = fmt.Sprintf("%dx%d", v.Width, v.Height)
		videoFormat = v.Format
		if len(m.Video) > 1 {
			videoFormat = fmt.Sprintf("%s+%d", v.Format, len(m.Video)-1)
		}
	}
	if len(m.Audio) > 0 {
		a := m.Audio[0]
		audioFormat = a.Format
		if len(m.Audio) > 1 {
			audioFormat = fmt.Sprintf("%s+%d", a.Format, len(m.Audio)-1)
		}
	}

	textCount := ""
	if len(m.Text) > 0 {
		textCount = fmt.Sprintf("x%d", len(m.Text))
	}

	line := fmt.Sprintf("%8d kb/s  %8s  %9s  %-6s %-6s  %4s  %s",
		m.OverallBitrate/1_000,
		m.Duration.Round(time.Second),
		dimensions,
		videoFormat, audioFormat,
		textCount,
		filepath.Base(m.Name),
	)
	fmt.Println(truncate(line, width))
}
