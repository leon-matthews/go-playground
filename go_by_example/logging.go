package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
)

func main() {
	log.Println("Standard logger defined in package")

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("With microseconds")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("With file/line")

	// Create new logger with custom prefix
	mylog := log.New(os.Stdout, "mylog:", log.LstdFlags)
	mylog.Println("Using custom logger")

	// Loggers can write to any io.Writer
	var buf bytes.Buffer
	buflog := log.New(&buf, "buflog:", log.LstdFlags)
	buflog.Println("This is going to the buffer")
	fmt.Println(buf.String())

	// Structured logging using the [log/slog] package
	handler := slog.NewJSONHandler(os.Stdout, nil)
	jlog := slog.New(handler)
	jlog.Info("Structured logging")

	// Alternate key/value pairs as shortcut...
	jlog.Info("Alternating", "key", "value", "number", 42)

	// ...but best practice is to use [slog.Attr] values...
	jlog.Info(
		"You can call them 'strongly-typed contextual attributes' if you like",
		slog.Bool("pretentious", true),
		slog.String("trying", "too hard"),
		slog.Any("error", errors.New("error printing error")),
	)

	// ...which can be grouped
	jlog.Info(
		"Grouped attributes might be useful one day",
		slog.Group(
			"properties",
			slog.Int("width", 1024),
			slog.Int("height", 768),
			slog.String("format", "png"),
		),
	)

	// A logger can be wrapped to carry default attributes
	// This is also a performance optimisation
	url := "https://example.com/login"
	urllog := jlog.With(slog.String("url", url))
	urllog.Warn("Could not connect: no such host")
}
