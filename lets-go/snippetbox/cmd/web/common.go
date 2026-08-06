package main

import (
	"log/slog"
	"net/http"
)

// serverError logs an error, then sends generic 500 response to user
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), slog.String("method", r.Method), slog.String("uri", r.URL.RequestURI()))
	msg := http.StatusText(http.StatusInternalServerError)
	http.Error(w, msg, http.StatusInternalServerError)
}

// clientError sends given status code and generic message
func (app *application) clientError(w http.ResponseWriter, status int) {
	msg := http.StatusText(status)
	http.Error(w, msg, status)
}
