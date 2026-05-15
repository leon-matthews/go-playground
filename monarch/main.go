// List media files in a folder, sorted by bitrate descending
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"golang.org/x/term"

	"go-playground/monarch/mediainfo"
)

func main() {
	jobs := pflag.IntP("jobs", "j", runtime.NumCPU()/2, "number of parallel mediainfo processes")
	kbs := pflag.IntP("kbs", "k", 0, "minimum overall bitrate in kb/s (default: no filter)")
	verbose := pflag.BoolP("verbose", "v", false, "enable debug logging")
	pflag.Parse()

	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if pflag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <folder>\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(1)
	}
	folder := pflag.Arg(0)

	if _, err := mediainfo.Version(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	media, s, err := scan(folder, *jobs)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slices.SortFunc(media, func(a, b *mediainfo.Media) int {
		return b.OverallBitrate - a.OverallBitrate
	})

	if *kbs > 0 {
		media = slices.DeleteFunc(media, func(m *mediainfo.Media) bool {
			return m.OverallBitrate < *kbs*1_000
		})
	}

	width := terminalWidth()
	if len(media) > 0 {
		printHeader(width)
	}
	for _, m := range media {
		printLine(m, width)
	}
	printSummary(s, width)
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

type stats struct {
	Total    int
	NonMedia int
	Errors   int
}

type result struct {
	media *mediainfo.Media
	err   error
}

// scan reads the folder (non-recursive) and returns media info for each file
// that mediainfo can parse; unparseable files are skipped.
func scan(folder string, jobs int) ([]*mediainfo.Media, stats, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, stats{}, err
	}

	paths := make(chan string)
	results := make(chan result)

	var wg sync.WaitGroup
	for range jobs {
		wg.Go(func() {
			for path := range paths {
				info, err := mediainfo.Info(path)
				if err != nil {
					slog.Warn("skipping", "path", path, "err", err)
					results <- result{err: err}
					continue
				}
				results <- result{media: info}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var total int
	go func() {
		for _, e := range entries {
			if !e.IsDir() {
				total++
				paths <- filepath.Join(folder, e.Name())
			}
		}
		close(paths)
	}()

	var media []*mediainfo.Media
	var s stats
	for r := range results {
		if r.err != nil {
			if errors.Is(r.err, mediainfo.ErrTimeout) {
				s.Errors++
			} else {
				s.NonMedia++
			}
		} else {
			media = append(media, r.media)
		}
	}
	s.Total = total
	return media, s, nil
}

func printSummary(s stats, width int) {
	line := fmt.Sprintf("%d files  %d non-media  %d errors", s.Total, s.NonMedia, s.Errors)
	fmt.Println(truncate(line, width))
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
