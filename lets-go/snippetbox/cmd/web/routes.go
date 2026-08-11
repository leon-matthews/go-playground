package main

import (
	"net/http"

	"github.com/justinas/alice"

	assets "local.dev/snippetbox/www"
)

func (app *application) routes() http.Handler {
	// Static files
	mux := http.NewServeMux()
	fileserver := http.FileServerFS(assets.StaticFiles)
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	// Middleware chain for just dynamic pages
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// Snippets
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.index))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	mux.Handle("GET /snippet/create", dynamic.ThenFunc(app.snippetCreate))
	mux.Handle("POST /snippet/create", dynamic.ThenFunc(app.snippetCreatePost))

	// Users
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("POST /user/logout", dynamic.ThenFunc(app.userLogoutPost))

	// Wrap all handlers in standard middleware
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}
