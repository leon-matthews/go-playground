package main

import (
	"fmt"
)

// Closing a channel indicates that no more values will be sent on it.
func main() {
	jobs := make(chan int, 5)
	done := make(chan bool)

	go func() {
		for {
			j, more := <-jobs
			if more {
				fmt.Println("received job", j)
			} else {
				fmt.Println("received all jobs")
				done <- true
				return
			}
		}
	}()

	for j := range 10 {
		job := j + 1
		jobs <- job
		fmt.Println("sent job", job)
	}
	close(jobs)
	fmt.Println("sent all jobs")

	// Block until we receive a notification from the worker on the channel.
	<-done

	// Check that `jobs` is empty
	_, more := <-jobs
	fmt.Println("more jobs", more)
}
