package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"sync"
	"time"
)

var mu sync.Mutex
var counter int

func main() {
	// Create a trace output file
	f, err := os.Create("out.bad")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()

	// Allocate the available number of CPUs
	fmt.Println("Starting workers...")
	workerCount := runtime.NumCPU()

	// Create multiple goroutines with lock contention
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Println("All workers done.")
}

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	mu.Lock() // Simulating lock contention
	time.Sleep(10 * time.Millisecond)
	counter++
	mu.Unlock()
}
