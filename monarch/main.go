package main

import (
	"fmt"
	"log/slog"
	"os"

	"go-playground/monarch/mediainfo"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	version, err := mediainfo.Version()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("Using %s %s", mediainfo.Binary, version))
}
