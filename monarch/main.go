package main

import (
	"fmt"
	"log/slog"
	"os"

	"go-playground/monarch/mediainfo"
)

func main() {
	// Logger
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// Check tool installation
	version, err := mediainfo.Version()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("Using %s %s", mediainfo.Binary, version))

	// Cute cow!
	name := "mediainfo/testdata/cow.mp4"
	tracks, err := mediainfo.Info(name)
    if err != nil {
        slog.Error(err.Error())
        os.Exit(1)
    }
    for _, track := range tracks {
        fmt.Printf("[%T]%+[1]v\n", track)
    }
}
