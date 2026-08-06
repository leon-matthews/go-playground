package main

import (
	"database/sql"
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
	logger   *slog.Logger
	snippets *models.SnippetModel
}

func main() {
	// Flags
	addr := flag.String("addr", ":8000", "HTTP server address")
	dsn := flag.String("dsn", defaultDSN, "MariaDB data source name")
	flag.Parse()

	// Logging
	options := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stdout, &options)
	logger := slog.New(handler)

	// Database
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Global state
	app := application{
		logger:   logger,
		snippets: &models.SnippetModel{DB: db},
	}

	// Let's go
	logger.Info("starting server", slog.String("addr", *addr))
	err = http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB wraps sql.Open() and returns an sql.DB connection poll
func openDB(dsn string) (*sql.DB, error) {
	// Create pool
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Actually make connection
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
