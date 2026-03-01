// Canceling goroutines.
package main

import (
	"fmt"
	"time"
)

// solution start

// count sends numbers from start to infinity to the out channel.
func count(cancel <-chan struct{}, start int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; ; i++ {
			select {
			case <-cancel:
				fmt.Println("count: close output")
				close(out)

				fmt.Println("count: exit goroutine")
				return
			case out <- i:
			}
		}
	}()
	return out
}

// take selects the first n numbers from in and sends them to the out channel.
func take(cancel <-chan struct{}, in <-chan int, n int) <-chan int {
	out := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			select {
			case <-cancel:
				fmt.Println("take: close output")
				close(out)

				fmt.Println("take: exit goroutine")
				return
			case out <- <-in:
			}
		}

		close(out)
	}()
	return out
}

// solution end

func main() {
	cancel := make(chan struct{})
	// defer close(cancel)

	stream := take(cancel, count(cancel, 10), 5)
	first := <-stream
	second := <-stream
	third := <-stream

	fmt.Println(first, second, third)
	close(cancel)

	time.Sleep(1 * time.Second)
}
