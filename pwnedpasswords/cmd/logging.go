package main

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

// newLogger builds a stderr logger whose level follows the -v and -q flags,
// with -v winning if both are set.
func newLogger(verbose, quiet bool) *log.Logger {
	level := log.InfoLevel
	switch {
	case verbose:
		level = log.DebugLevel
	case quiet:
		level = log.WarnLevel
	}
	return log.NewWithOptions(os.Stderr, log.Options{
		Level:           level,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
}
