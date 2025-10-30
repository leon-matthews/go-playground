package main

import (
	"bytes"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var globalStore [][]byte

func main() {
	go leakMemory()

	// pprof server
	log.Println("Server running on :6060")
	if err := http.ListenAndServe(":6060", nil); err != nil {
		log.Fatal(err)
	}
}

func leakMemory() {
	for {
		// Simulate memory allocations
		data := bytes.Repeat([]byte("x"), 1_000_000)
		globalStore = append(globalStore, data)

		time.Sleep(500 * time.Millisecond)
	}
}
