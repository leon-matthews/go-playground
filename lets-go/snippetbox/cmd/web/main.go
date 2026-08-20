// Package main implements the entire web application
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"

	"local.dev/snippetbox/internal/models"
)

const (
	defaultDSN       = "web:web@/snippetbox?parseTime=true"
	userIDSessionKey = "authenticatedUserId"
)

// application holds dependencies for the web app.
type application struct {
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templates      templateCache
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// Flags
	addr := flag.String("addr", ":8000", "HTTP server address")
	dsn := flag.String("dsn", defaultDSN, "MariaDB data source name")
	flag.Parse()

	// Logging
	logOptions := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logHandler := slog.NewTextHandler(os.Stdout, &logOptions)
	logger := slog.New(logHandler)

	// Database
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := models.OpenDB(ctx, *dsn)
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

	// Sessions
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 8 * 24 * time.Hour // 8 days

	// Global state
	app := application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templates:      templates,
		formDecoder:    form.NewDecoder(),
		sessionManager: sessionManager,
	}

	// TLS Setup
	// Recommended configurations from:
	// https://docs.tlsref.org/server-side-tls.html
	tlsConfig := &tls.Config{
		// Avoid slow elliptic curve implementations. Despite the name this
		// value is just set of supported key-exchange mechanisms
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},

		// Note that when TLS 1.3 is in use, `tls.Config.CipherSuites` is ignored.
		MinVersion: tls.VersionTLS13,
	}

	// Server
	srv := &http.Server{
		Addr:    *addr,
		Handler: app.routes(),
		// Requires a *log.Logger which can create using our existing slog handler
		ErrorLog:  slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		// Timeouts
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Let's go
	logger.Info("starting server", slog.String("addr", *addr))
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}
