package main

import "fmt"

func main() {
	// Start goroutine to do some work
	out := make(chan int)
	done := make(chan struct{})
	go work(done, out)

	// Close out once workers have all finished
	go func() {
		<-done
		close(out)
	}()

	// Read from out until closed
	for n := range out {
		fmt.Println(n)
	}
}

// work writes values to out, signals on done when finished.
func work(done chan<- struct{}, out chan<- int) {
	for i := 1; i <= 5; i++ {
		out <- i
	}
	done <- struct{}{}
}
