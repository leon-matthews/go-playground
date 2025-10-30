package main

import (
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	// Any handlers or routes would go here
	go startCPUHeavyRoutine()

	// Listen and serve on a dedicated port
	log.Println("Starting server on :6060")
	log.Println("pprof is available at http://localhost:6060/debug/pprof/")
	if err := http.ListenAndServe(":6060", nil); err != nil {
		log.Fatal(err)
	}
}

func startCPUHeavyRoutine() {
	for {
		doHeavyComputation()
		time.Sleep(500 * time.Millisecond)
	}
}

func doHeavyComputation() {
	sum := 0.0
	for i := 0; i < 1_000_000; i++ {
		sum += math.Sin(float64(i))
	}
}
