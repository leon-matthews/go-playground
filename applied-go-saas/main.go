package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const addr = ":8080"

//go:embed www/static/*
var staticAssets embed.FS

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello SaaS")
}

func main() {
	// Static assets
	staticFS, err := fs.Sub(staticAssets, "www/static")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(staticFS))

	// Use timeouts
	server := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start servin'!
	go func() {
		log.Printf("Starting server on http://localhost:%v/\n", addr)
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Handle SIGTERM (ctrl+c or kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutdown signal received. Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown due to error: %v", err)
	}
	log.Println("Server exited gracefully")
	defer cancel()
}
