package main

import (
	"fmt"
	"log"
	"net/http"
)

const addr = ":8080"

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello SaaS")
}

func main() {
	log.Println("Starting server on", addr)
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handler)))
}
