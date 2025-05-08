package main

import (
	"fmt"
	"time"
)

func main() {
	unbuffered()
	buffering()
	synchronisation()
	directions()
	selectChannel()
	selectTimeout()
	selectTimeoutTriggered()
	NonBlockingOperations()
	ClosingChannels()
	RangeOverChannels()
}

// By default sends and receives block until both the sender and
// receiver are ready.
func unbuffered() {
	messages := make(chan string)

	// Send string to channel in a goroutine
	go func() { messages <- "Hello" }()

	msg := <-messages
	fmt.Println(msg)
}

// Here we make a channel of strings buffering up to 2 values.
func buffering() {
	messages := make(chan string, 2)

	messages <- "buffered"
	messages <- "buffered2"
	// Too may writes cause a panic!
	// fatal error: all goroutines are asleep - deadlock!
	//~ messages <- "buffered3"

	fmt.Println(<-messages)
	fmt.Println(<-messages)

	// Too many reads cause panic
	// fatal error: all goroutines are asleep - deadlock!
	//~ fmt.Println(<- messages)

}

// We can use channels to synchronize execution across goroutines.
func synchronisation() {
	done := make(chan bool)
	go worker(done)

	// Block until we receive a notification from the worker on the channel
	<-done
}

func worker(done chan bool) {
	fmt.Print("working...")
	time.Sleep(300 * time.Millisecond)
	fmt.Println("done")

	// Send a value to notify that we’re done.
	done <- true
}

// You can specify channel direction in function parameters
func directions() {
	pings := make(chan string, 1)
	pongs := make(chan string, 1)
	ping(pings, "passed message")
	pong(pings, pongs)
	fmt.Println(<-pongs)
}

// Ping can only send to channel
func ping(pings chan<- string, msg string) {
	pings <- msg
}

// Pong can only receive over channel
func pong(pings <-chan string, pongs chan<- string) {
	msg := <-pings
	pongs <- msg
}

// Go's select lets you wait on multiple channel operations.
func selectChannel() {
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		time.Sleep(300 * time.Millisecond)
		c1 <- "one"
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		c2 <- "two"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			fmt.Println("received", msg1)
		case msg2 := <-c2:
			fmt.Println("received", msg2)
		}
	}
}

// Implement timeout in select, not triggered in this case
func selectTimeout() {
	ch1 := make(chan string, 1) // Buffered channel, non blocking
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "result 1"
	}()

	// Timeout not triggered
	select {
	case res := <-ch1:
		fmt.Println(res)
	case <-time.After(300 * time.Millisecond):
		fmt.Println("timeout1")
	}
}

// Timeout triggered
func selectTimeoutTriggered() {
	ch1 := make(chan string, 1) // Buffered channel, non blocking
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch1 <- "result 2"
	}()

	// Timeout triggered
	select {
	case res := <-ch1:
		fmt.Println(res)
	case <-time.After(200 * time.Millisecond):
		fmt.Println("timeout2")
	}
}

// Use select with a default clause to implement non-blocking sends, receives,
// and even non-blocking multi-way selects.
func NonBlockingOperations() {
	messages := make(chan string)
	signals := make(chan bool)

	// Non-blocking receive
	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	default:
		fmt.Println("no message received")
	}

	// Non-blocking send
	msg := "Hello"
	select {
	case messages <- msg:
		fmt.Println("sent message")
	default:
		fmt.Println("no message sent")
	}

	// Multi-way non-blocking select
	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	case sig := <-signals:
		fmt.Println("received signal", sig)
	default:
		fmt.Println("No activity")
	}
}

// Closing a channel indicates that no more values will be sent on it.
func ClosingChannels() {
	jobs := make(chan int, 2)
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

	for j := range 3 {
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
	fmt.Println("received more jobs", more)
}

// Range iterates over each element as it’s received from channel
func RangeOverChannels() {
	queue := make(chan string, 5)
	queue <- "one"
	queue <- "two"
	queue <- "three"

	// Without `close()`, program panics
	// fatal error: all goroutines are asleep - deadlock!
	close(queue)

	// Values can still be received from closed channel
	for word := range queue {
		fmt.Println(word)
	}
}
