package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"sync"
	"time"
)

func main() {
	f, err := os.Create("out.good")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()

	jobs := make(chan int, 5) // Buffered channel to avoid lock contention
	var wg sync.WaitGroup

	// Allocate the available number of CPUs
	workerCount := runtime.NumCPU()

	// Create workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	// Send jobs
	for j := 0; j < 10; j++ {
		jobs <- j
	}
	close(jobs)

	wg.Wait()
}

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		time.Sleep(10 * time.Millisecond) // Simulated work
		fmt.Printf("Worker %d processed job %d\n", id, job)
	}
}
