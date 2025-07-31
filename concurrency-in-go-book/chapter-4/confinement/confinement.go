package main

import (
	"fmt"
)

// Confinement is the idea of restricting ownership of data to ONE goroutine
// only, avoiding the need for synchonisation primitives like mutexes.
func main() {
	adHocConfinement()
	lexicalConfinement()
}

// adHocConfinement is when confinement is by convention only.
// Here we have a rule when the consumer is only 'allowed' to access data
// via the channel - but this could be circumvented if somebody was in a rush.
func adHocConfinement() {
	// Data owner
	data := make([]int, 5)
	loopData := func(out chan<- int) {
		defer close(out)
		for i := range data {
			out <- data[i]
		}
	}

	// Data consumer
	consumer := make(chan int)
	go loopData(consumer)
	for num := range(consumer) {
		fmt.Println(num)
	}
}

// lexicalConfinement is superior, using lexical scope to enforce confinement
func lexicalConfinement() {
	results := loopData()
	consumer(results)
}

// loopData 'owns' the data
func loopData() <-chan int {
	data := make([]int, 5)
	results := make(chan int)
	go func () {
		defer close(results)
		defer fmt.Println("Sending goroutine finished")
		for _, n := range data {
			results <-n
		}
	}()
	return results
}

func consumer(results <-chan int) {
	for r := range results {
		fmt.Println(r)
	}
	fmt.Println("Finished receiving")
}
