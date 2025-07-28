package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// Not-very satisfiying worker pool, where we have to:
// Submit ALL jobs ahead of time using a large buffered channel
// Collect finished work into another large buffered channel
func main() {
	var numJobs = 10

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// Create worker pool
	var numWorkers = 3
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	// Send jobs
	for j := 0; j <= numJobs; j++ {
		fmt.Println("Queue job", j)
		jobs <- j
	}
	// Iteration over a closed channel finishes when its empty
	close(jobs)

	// Collect results from all workers, allowing them to finish
	for a := 1; a <= numJobs; a++ {
		<-results
	}
}

func worker(id int, jobs <-chan int, results chan<- int) {
	// Blocks until a job is available
	for j := range jobs {
		fmt.Println("worker", id, "started  job", j)
		time.Sleep(rand.N(1000 * time.Millisecond))
		fmt.Println("worker", id, "finished  job", j)
		results <- j * 2
	}
}
