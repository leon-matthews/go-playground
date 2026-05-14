// List media files in a folder, sorted by bitrate descending
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"

	"go-playground/monarch/mediainfo"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

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
		return b.Bitrate - a.Bitrate
	})

	if len(media) > 0 {
		printHeader()
	}
	for _, m := range media {
		printLine(m)
	}
}

func printHeader() {
	fmt.Printf("%13s  %8s  %9s  %-6s %-6s  %s\n",
		"Bitrate", "Duration", "Size", "Video", "Audio", "Name")
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

func printLine(m *mediainfo.Media) {
	fmt.Printf("%8d kb/s  %8s  %4dx%-4d  %-6s %-6s  %s\n",
		m.Bitrate/1_000,
		m.Duration.Round(time.Second),
		m.Width, m.Height,
		m.VideoFormat, m.AudioFormat,
		filepath.Base(m.Name),
	)
}
