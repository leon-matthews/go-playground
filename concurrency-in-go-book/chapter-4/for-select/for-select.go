package main

import (
	"fmt"
	"time"
)

func main() {
	// Close after a second
	done := make(chan any)
	go func() {
		time.Sleep(1 * time.Second)
		close(done)
	}()

	// Send values from iterable to channel
	for i := range iterationToChannel(done) {
		fmt.Print(i)
	}
	fmt.Println()

	// Loop 'forever'
	loopForever(done)
}

// iterationToChannel uses for/select to send something that can be iterated
// over into a channel.
func iterationToChannel(done <-chan any) <-chan string {
	data := []string{"a", "b", "c", "d", "e", "f"}
	out := make(chan string)
	go func() {
		defer close(out)
		for _, s := range data {
			select {
			case out <- s:
			case <-done:
				break
			}
		}
	}()
	return out
}

// loopForever uses for/select to do exactly that - until it is cancelled
func loopForever(done <-chan any) {
	for {
		// Check for cancellation
		select {
		case <-done:
			return
		default:
		}

		// Do something that shouldn't be interrepted, ie. non-preempable work
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Looping!")
	}
}
