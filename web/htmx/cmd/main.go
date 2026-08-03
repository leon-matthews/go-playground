package main

import (
	"log/slog"
	"net/http"
	"os"

	"local.dev/htmx/assets"
)

// The application struct holds the dependencies needed for our handlers,
// including a htmlRenderer type.
type application struct {
	logger *slog.Logger
	html   *htmlRenderer
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Initialize a new htmlRenderer, parsing the base template and all partial
	// templates from assets/html into the shared template set.
	htmlRenderer, err := newHTMLRenderer(assets.HTMLFiles, "base.tmpl", "snippets/*.tmpl")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Include the htmlRenderer in the application struct.
	app := &application{
		logger: logger,
		html:   htmlRenderer,
	}

	// Create a file server that serves the files from assets/static.
	fileserver := http.FileServerFS(assets.StaticFiles)

	// Register the application routes.
	mux := http.NewServeMux()
    mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))
    mux.HandleFunc("GET /{$}", app.home)
    mux.HandleFunc("GET /gopher", app.gopher)

	// Start the HTTP server.
	logger.Info("starting server", "port", 5051)
	err = http.ListenAndServe(":5051", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
