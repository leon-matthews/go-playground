package main

import (
    "net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    err := app.html.render(w, 200, nil, "base", "pages/home.tmpl")
    if err != nil {
        app.logger.Error(err.Error())
        http.Error(w, http.StatusText(500), 500)
    }
}

func (app *application) gopher(w http.ResponseWriter, r *http.Request) {
    width := 100
    err := app.html.render(w, http.StatusOK, width, "partial:image:gopher")
    if err != nil {
        app.logger.Error(err.Error())
        http.Error(w, http.StatusText(500), 500)
    }
}
