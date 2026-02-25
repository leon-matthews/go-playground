package main

import "fmt"

func main() {
	out := make(chan int)
	done := make(chan struct{})

	// Write values to out, then write to done when finished
	go work(done, out)

	// Start goroutine to close out when done
	go func() {
		<-done
		close(out)
	}()

	// Keep iterating until out is closed
	for n := range out {
		fmt.Println(n)
	}
}

func work(done chan<- struct{}, out chan<- int) {
	for i := range 5 {
		out <- i
	}
	done <- struct{}{}
}
