package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/form/v4"
)

// decodePostForm populates destination with values from posted form.
func (app *application) decodePostForm(r *http.Request, destination any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(destination, r.PostForm)
	if err != nil {
		// Panic, as we've passed in the wrong type. Programmer error!
		if _, ok := errors.AsType[*form.InvalidDecoderError](err); ok {
			panic(err)
		}
		return err
	}
	return nil
}

// clientError sends given status code and generic message
func (app *application) clientError(w http.ResponseWriter, status int) {
	msg := http.StatusText(status)
	http.Error(w, msg, status)
}

// Return true if the current request is from an authenticated user
// Uses value in context put there by the authenticate middleware
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

// serverError logs an error, then sends generic 500 response to user
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), slog.String("method", r.Method), slog.String("uri", r.URL.RequestURI()))
	msg := http.StatusText(http.StatusInternalServerError)
	http.Error(w, msg, http.StatusInternalServerError)
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
