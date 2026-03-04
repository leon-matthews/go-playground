package concurrency

import (
	"time"
)

// =============================================================================
// EXERCISE 1.3: Select Statement Mastery
// =============================================================================
//
// The select statement is Go's way of handling multiple channel operations
// concurrently. It's similar to switch but for channels.
//
// KEY CONCEPTS:
// - select blocks until ONE case can proceed, then executes that case
// - If multiple cases are ready, one is chosen at RANDOM (fair selection)
// - default case makes select non-blocking
// - Closed channels are always ready to receive (return zero value)
// - time.After returns a channel that receives after a duration
//
// NOTE: Before Go 1.23, time.After created timers that weren't garbage
// collected until they fired, causing leaks in loops. Since Go 1.23,
// unreferenced timers are collected even if they haven't fired.
//
// =============================================================================

// =============================================================================
// PART 1: Basic Select Patterns
// =============================================================================

// FirstResponse returns the first value received from either channel.
// This is the "first wins" pattern used in redundant systems.
// If both are ready simultaneously, either is acceptable.
func FirstResponse(ch1, ch2 <-chan string) string {
	var result string
	select {
	case result = <-ch1:
	case result = <-ch2:
	}
	return result
}

// MergeChannels combines two channels into one output channel.
// Values from both inputs appear on the output in arrival order.
//
// 1. Create an output channel
// 2. Launch a goroutine that:
//   - Uses select in a loop to receive from either ch1 or ch2
//   - Sends received values to output
//   - Handles closure of both channels properly
//   - Closes output when BOTH inputs are closed
//
// 3. Return the output channel
//
// HINT: You need to track which channels are still open. A nil channel
// in select is never ready - use this to "disable" closed channels.
func MergeChannels(ch1, ch2 <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for numClosed := 0; numClosed < 2; {
			select {
			case v, ok := <-ch1:
				if ok == false {
					ch1 = nil
					numClosed++
				} else {
					out <- v
				}
			case v, ok := <-ch2:
				if ok == false {
					ch2 = nil
					numClosed++
				} else {
					out <- v
				}
			}
		}
	}()
	return out
}

// =============================================================================
// PART 2: Timeouts and Deadlines
// =============================================================================

// ReceiveWithTimeout receives from a channel with a timeout.
// Returns (value, true) if received, (zero, false) if timeout.
func ReceiveWithTimeout(ch <-chan int, timeout time.Duration) (int, bool) {
	var result int
	var ok bool
	select {
	case result = <-ch:
		ok = true
	case <-time.After(timeout):
		ok = false
	}
	return result, ok
}

// ReceiveWithDeadline receives until a specific time.
// Returns all values received before the deadline.
func ReceiveWithDeadline(ch <-chan int, deadline time.Time) []int {
	out := make([]int, 0)
	timeout := time.Until(deadline)
	for {
		select {
		case v, ok := <-ch:
			if ok == false {
				return out
			}
			out = append(out, v)
		case <-time.After(timeout):
			return out
		}
	}
}

// PeriodicTask runs a function periodically until done is closed.
// Call fn() every interval, stop when done is closed
// Return the number of times fn was called
func PeriodicTask(fn func(), interval time.Duration, done <-chan struct{}) int {
	count := 0
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			fn()
			count++
		case <-done:
			return count
		}
	}
}

// =============================================================================
// PART 3: Non-Blocking Operations with Default
// =============================================================================

// TrySend attempts to send without blocking.
// Returns true if send succeeded, false if channel is full/blocked.
// TODO: Use select with default to make non-blocking send
func TrySend(ch chan<- int, value int) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}

// TryReceive attempts to receive without blocking.
// Returns (value, true) if received, (zero, false) if channel is empty.
func TryReceive(ch <-chan int) (int, bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		return 0, false
	}
}

// DrainChannel empties a channel without blocking.
// Returns all values that were buffered.
func DrainChannel(ch <-chan int) []int {
	values := make([]int, 0)
	for {
		v, ok := TryReceive(ch)
		if !ok {
			break
		}
		values = append(values, v)
	}
	return values
}

// =============================================================================
// PART 4: Priority Select (Trick Question!)
// =============================================================================

// PriorityReceive should receive from highPriority if available,
// otherwise from lowPriority.
//
// QUESTION: Why doesn't this simple implementation work correctly?
//
//	select {
//	case v := <-highPriority:
//	    return v, "high"
//	case v := <-lowPriority:
//	    return v, "low"
//	}
//
// ANSWER: If both are ready to read, the Go runtime will select one randomly.
func PriorityReceive(highPriority, lowPriority <-chan int) (int, string) {
	select {
	case v := <-highPriority:
		return v, "high"
	default:
		select {
		case v := <-lowPriority:
			return v, "low"
		case v := <-highPriority:
			return v, "high"
		}
	}
	return 0, ""
}

// =============================================================================
// PART 5: Select with Send and Receive
// =============================================================================

// Relay forwards values from input to output, with buffering.
// Stops when input is closed AND buffer is empty.
//
// TODO: This is tricky! You need to:
// 1. Use an internal buffer (slice)
// 2. select should try to:
//   - Receive from input (if not closed) -> add to buffer
//   - Send to output (if buffer not empty) -> remove from buffer
//
// 3. Handle input closure and drain buffer before returning
//
// HINT: You can conditionally enable select cases using nil channels
// If buffer is empty, set the "send" channel to nil to disable that case
func Relay(input <-chan int, output chan<- int) {
	// YOUR CODE HERE
}

// =============================================================================
// CHALLENGE: Implement a multiplexer
// =============================================================================

// Multiplex routes values from input to one of N output channels.
// The routeFn determines which output (0 to n-1) each value goes to.
// Stops when input is closed (close all outputs).
//
// TODO: Implement multiplexing with select
// HINT: You can't use select with a dynamic number of cases directly.
// One approach: try each output in order using non-blocking sends.
func Multiplex(input <-chan int, outputs []chan<- int, routeFn func(int) int) {
	// YOUR CODE HERE
}

// =============================================================================
// CHALLENGE: Implement fair merge
// =============================================================================

// FairMerge merges N channels with fair scheduling.
// No single channel can starve others even if it's always ready.
//
// TODO: Implement round-robin selection from channels
// Return values in round-robin order (not arrival order)
// Stop when all channels are closed
//
// HINT: Track current index, try each channel in order
func FairMerge(channels []<-chan int) <-chan int {
	// YOUR CODE HERE
	return nil
}
