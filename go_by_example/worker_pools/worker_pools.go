package main

import (
	"fmt"
	"time"
)

func main() {
	var numJobs = 5

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// Create worker pool
	var numWorkers = 3
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	// Send jobs
	for j := 0; j <= numJobs; j++ {
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
		time.Sleep(1 * time.Second)
		fmt.Println("worker", id, "finished  job", j)
		results <- j * 2
	}
}
