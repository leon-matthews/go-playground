package concurrency

import (
	"time"
)

// =============================================================================
// EXERCISE 1.1: Channel Semantics
// =============================================================================
//
// This exercise explores the fundamental differences between buffered and
// unbuffered channels, and how channels behave when closed.
//
// KEY CONCEPTS:
// - Unbuffered channels block until both sender AND receiver are ready
// - Buffered channels block only when the buffer is full (send) or empty (receive)
// - Receiving from a closed channel returns the zero value immediately
// - Sending to a closed channel panics
// - You can detect closure with the "comma ok" idiom: val, ok := <-ch
//
// =============================================================================

// UnbufferedDemo demonstrates unbuffered channel behavior.
// An unbuffered channel blocks the sender until a receiver is ready.
//
// TODO: Implement this function to:
// 1. Create an unbuffered channel of int
// 2. Launch a goroutine that sends the value 42 to the channel
// 3. Sleep for 100ms in the main goroutine (simulating work)
// 4. Receive from the channel and return the value
//
// QUESTION: Where does the goroutine block? Before or after the sleep?
func UnbufferedDemo() int {
	unbuffered := make(chan int)
	go func() {
		unbuffered <- 42
	}()
	time.Sleep(100 * time.Millisecond)
	return <-unbuffered
}

// BufferedDemo demonstrates buffered channel behavior.
// A buffered channel allows sends to complete without blocking (until full).
func BufferedDemo() int {
	buffered := make(chan int, 1)
	go func() {
		buffered <- 42
	}()
	time.Sleep(100 * time.Millisecond)
	return <-buffered
}

// BufferFullDemo demonstrates what happens when a buffer fills up.
//
// TODO: Implement this function to:
// 1. Create a buffered channel of int with capacity 2
// 2. Send values 1, 2 to the channel (fills the buffer)
// 3. Launch a goroutine that will receive one value after 50ms
// 4. Send value 3 (this should block until the goroutine receives!)
// 5. Return true if all sends completed successfully
//
// QUESTION: What would happen if you didn't have the goroutine?
func BufferFullDemo() bool {
	stream := make(chan int, 2)
	stream <- 1
	stream <- 2
	go func() {
		time.Sleep(50 * time.Millisecond)
		<-stream
	}()
	stream <- 3
	return true
}

// ClosedChannelReceive demonstrates receiving from a closed channel.
//
// TODO: Implement this function to:
// 1. Create a buffered channel of string with capacity 2
// 2. Send "first" and "second" to the channel
// 3. Close the channel
// 4. Receive ALL values (including after close) and return them as a slice
//
// HINT: Use the "comma ok" idiom to detect when channel is exhausted
// QUESTION: How many receives can you do? What do you get after the buffered values?
func ClosedChannelReceive() []string {
	s := make(chan string, 2)
	s <- "first"
	s <- "second"
	close(s)

	l := make([]string, 0, 2)
	for {
		v, ok := <-s
		if !ok {
			break
		}
		l = append(l, v)
	}
	return l
}

// RangeOverChannel demonstrates using range to receive until close.
//
// TODO: Implement this function to:
// 1. Create an unbuffered channel of int
// 2. Launch a goroutine that sends values 1, 2, 3 then closes the channel
// 3. Use `for val := range ch` to collect all values
// 4. Return the sum of all values
//
// NOTE: range automatically stops when channel is closed
func RangeOverChannel() int {
	stream := make(chan int)
	go func() {
		stream <- 1
		stream <- 2
		stream <- 3
		close(stream)
	}()

	sum := 0
	for v := range stream {
		sum += v
	}
	return sum
}

// NilChannelBehavior demonstrates that nil channels block forever.
//
// TODO: Implement this function to:
// 1. Create a nil channel (var ch chan int, NOT make(chan int))
// 2. Use select with the nil channel and a timeout of 100ms
// 3. Return "timeout" if the select hit the timeout case
// 4. Return "received" if somehow a value was received (should never happen!)
//
// QUESTION: Why would you ever want a nil channel? (Hint: dynamic select cases)
func NilChannelBehavior() string {
	var stream chan int
	select {
	case <-stream:
		return "received"
	case <-time.After(100 * time.Millisecond):
		return "timeout"
	}
}

// ChannelDirection demonstrates send-only and receive-only channel types.
// This is a compile-time safety feature.
//
// TODO: Complete these three functions:

// generator creates values and sends them on a send-only channel
func generator(out chan<- int, count int) {
	go func() {
		for n := range count {
			out <- n
		}
		close(out)
	}()
}

// squarer receives from one channel, squares, sends to another
func squarer(in <-chan int, out chan<- int) {
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
}

// ChannelDirectionDemo ties it together
// TODO: Create channels, wire up generator -> squarer, return sum of squares
func ChannelDirectionDemo(count int) int {
	sum := 0
	numbers := make(chan int)
	generator(numbers, count)
	squares := make(chan int)
	squarer(numbers, squares)
	for square := range squares {
		sum += square
	}
	return sum
}

// =============================================================================
// CHALLENGE: Implement a timeout pattern without time.After
// =============================================================================

// SendWithTimeout attempts to send a value with a timeout.
// Returns true if send succeeded, false if timeout occurred.
//
// TODO: Implement WITHOUT using time.After (use time.NewTimer instead)
// Note: Go 1.23+ garbage collects unreferenced timers even if they haven't
// fired, but using time.NewTimer with Stop() is still good practice.
//
// HINT: time.NewTimer returns a *Timer with a channel C and method Stop()
func SendWithTimeout(ch chan<- int, value int, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case ch <- 42:
			return true
		case <-timer.C:
			return false
		}
	}
}
