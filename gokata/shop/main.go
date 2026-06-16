// Shop is a web shop selling shoes and socks. It registers the /list and /price
// endpoints on an explicit http.ServeMux. Adapted from
// https://github.com/adonovan/gopl.io/tree/master/ch7/http4.
package main

import (
	"context"
	"fmt"
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

	db := database{"shoes": 50, "socks": 5}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /list", db.list)
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

type dollars float32

func (d dollars) String() string {
	return fmt.Sprintf("$%.2f", d)
}

type database map[string]dollars

func (db database) list(w http.ResponseWriter, r *http.Request) {
	for item, price := range db {
		fmt.Fprintf(w, "%s: %s\n", item, price)
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
