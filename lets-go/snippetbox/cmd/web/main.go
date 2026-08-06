package main

import (
	"flag"
	"log"
	"net/http"

	assets "local.dev/snippetbox/www"
)

func main() {
	// Flags
	addr := flag.String("addr", ":8000", "HTTP server address")
	flag.Parse()

	// Multiplex!
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", index)
	mux.HandleFunc("GET /snippet/view/{id}", snippetView)
	mux.HandleFunc("GET /snippet/create", snippetCreate)
	mux.HandleFunc("POST /snippet/create", snippetCreatePost)

	// Static files
	fileserver := http.FileServerFS(assets.StaticFiles)
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	// Let's go
	log.Printf("starting server on %s", *addr)
	err := http.ListenAndServe(*addr, mux)
	log.Fatal(err)
}
