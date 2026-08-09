package main

import (
	"bytes"
	"fmt"
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

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	// Retrieve the appropriate template set from the page name, eg. "index.html"
	ts, ok := app.templates[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// Write the template to the buffer, so we can catch runtime errors
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Write out status and body
	w.WriteHeader(status)
	buf.WriteTo(w)
}
