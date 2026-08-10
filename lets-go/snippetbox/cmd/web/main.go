package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"local.dev/snippetbox/internal/models"
)

const defaultDSN = "web:web@/snippetbox?parseTime=true"

// application holds dependencies for the web app.
type application struct {
	logger    *slog.Logger
	snippets  *models.SnippetModel
	templates templateCache
}

func main() {
	// Flags
	addr := flag.String("addr", ":8000", "HTTP server address")
	dsn := flag.String("dsn", defaultDSN, "MariaDB data source name")
	flag.Parse()

	// Logging
	options := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, &options)
	logger := slog.New(handler)

	// Database
	db, err := models.OpenDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Initialize a new template cache...
	templates, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Global state
	app := application{
		logger:    logger,
		snippets:  &models.SnippetModel{DB: db},
		templates: templates,
	}

	// Let's go
	logger.Info("starting server", slog.String("addr", *addr))
	err = http.ListenAndServe(*addr, commonHeaders(app.routes()))
	logger.Error(err.Error())
	os.Exit(1)
}
