package main

import (
	"net/http"

	"github.com/justinas/alice"

	assets "local.dev/snippetbox/www"
)

func (app *application) routes() http.Handler {
	// Index
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.index)

	// Snippets
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

	// Static files
	fileserver := http.FileServerFS(assets.StaticFiles)
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	// Wrap mux in our middleware
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}
