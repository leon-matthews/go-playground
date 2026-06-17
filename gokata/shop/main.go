// Shop is a web shop selling shoes and socks. It serves an HTML item listing at /
// and individual prices at /price, registered on an explicit http.ServeMux. Adapted
// from https://github.com/adonovan/gopl.io/tree/master/ch7/http4.
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const address = "localhost:8000"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := database{"shoes": 50_00, "socks": 5_00}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", db.list)
	mux.HandleFunc("GET /list", db.list) // alias for the root listing
	mux.HandleFunc("GET /price", db.price)

	const writeTimeout = 10 * time.Second

	srv := &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: writeTimeout,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting server at http://%s", address)
	srvErr := make(chan error, 1)
	go func() { srvErr <- srv.ListenAndServe() }()

	// Serve until the server fails or a signal arrives.
	select {
	case err := <-srvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	// Drain in-flight requests, allowing longer than WriteTimeout so the drain wins.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), writeTimeout+5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// cents is a USD monetary amount in whole cents.
type cents int64

// String formats the amount as dollars and cents, e.g. "$50.00" or "-$1.50".
func (c cents) String() string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s$%d.%02d", sign, c/100, c%100)
}

type database map[string]cents

// listTemplate renders the item listing; ranging the map visits keys in sorted order.
var listTemplate = template.Must(template.New("list").Parse(`<!DOCTYPE html>
<title>Shop</title>
<ul>
{{range $item, $price := .}}<li><a href="/price?item={{$item}}">{{$item}}</a>: {{$price}}</li>
{{end}}</ul>
`))

func (db database) list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := listTemplate.Execute(w, db); err != nil {
		log.Printf("list: %v", err)
	}
}

func (db database) price(w http.ResponseWriter, r *http.Request) {
	item := r.URL.Query().Get("item")
	price, ok := db[item]
	if !ok {
		w.WriteHeader(http.StatusNotFound) // 404
		fmt.Fprintf(w, "no such item: %q", item)
		return
	}
	fmt.Fprintf(w, "%s", price)
}
