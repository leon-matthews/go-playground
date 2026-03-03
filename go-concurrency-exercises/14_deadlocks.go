package concurrency

import (
	"sync"
)

// =============================================================================
// EXERCISE 4.2: Deadlock Scenarios
// =============================================================================
//
// A deadlock occurs when goroutines are waiting for each other forever.
// Go's runtime detects some deadlocks and panics with "all goroutines are asleep".
//
// KEY CONCEPTS:
// - Channel deadlock: sender waits for receiver (or vice versa) that will never come
// - Mutex deadlock: goroutines hold locks and wait for each other's locks
// - WaitGroup deadlock: Wait() called but Done() will never be called enough times
// - Go only detects deadlock when ALL goroutines are blocked
//
// =============================================================================

// =============================================================================
// PART 1: Channel Deadlocks
// =============================================================================

// DeadlockUnbufferedSend demonstrates sending without a receiver.
// QUESTION: Why does this deadlock?
func DeadlockUnbufferedSend() int {
	ch := make(chan int)
	ch <- 42 // DEADLOCK: no receiver
	return <-ch
}

// FixUnbufferedSend should work correctly.
//
// TODO: Fix the deadlock (multiple valid solutions)
func FixUnbufferedSend() int {
	// YOUR CODE HERE
	return 0
}

// DeadlockReceiveNoSend demonstrates receiving without a sender.
func DeadlockReceiveNoSend() int {
	ch := make(chan int)
	return <-ch // DEADLOCK: no sender
}

// FixReceiveNoSend should work correctly.
//
// TODO: Fix the deadlock
func FixReceiveNoSend() int {
	// YOUR CODE HERE
	return 0
}

// DeadlockBufferedFull demonstrates a full buffer blocking.
func DeadlockBufferedFull() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	ch <- 3 // DEADLOCK: buffer full, no receiver
}

// FixBufferedFull should send all three values.
//
// TODO: Fix the deadlock
func FixBufferedFull() []int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 2: WaitGroup Deadlocks
// =============================================================================

// DeadlockWaitGroupNotDone demonstrates Done() never called.
// QUESTION: What happens if a goroutine panics before calling Done()?
func DeadlockWaitGroupNotDone() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Forgot to call wg.Done()!
		_ = 1 + 1
	}()
	wg.Wait() // DEADLOCK: Done() never called
}

// FixWaitGroupNotDone should complete successfully.
//
// TODO: Fix using defer
func FixWaitGroupNotDone() {
	// YOUR CODE HERE
}

// DeadlockWaitGroupAddInGoroutine demonstrates Add() called too late.
func DeadlockWaitGroupAddInGoroutine() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		go func() {
			wg.Add(1) // RACE/DEADLOCK: Add() inside goroutine
			defer wg.Done()
			_ = 1 + 1
		}()
	}
	wg.Wait() // Might return too early or deadlock
}

// FixWaitGroupAddInGoroutine should complete successfully.
//
// TODO: Fix by moving Add() to correct location
func FixWaitGroupAddInGoroutine() {
	// YOUR CODE HERE
}

// =============================================================================
// PART 3: Mutex Deadlocks
// =============================================================================

// DeadlockMutexRecursive demonstrates trying to lock twice.
func DeadlockMutexRecursive() {
	var mu sync.Mutex
	mu.Lock()
	mu.Lock() // DEADLOCK: same goroutine, can't lock twice
	mu.Unlock()
	mu.Unlock()
}

// FixMutexRecursive should work correctly.
//
// TODO: Restructure to avoid recursive locking
// HINT: Go's Mutex is not reentrant by design
func FixMutexRecursive() {
	// YOUR CODE HERE
}

// DeadlockMutexOrdering demonstrates AB-BA deadlock.
// Two goroutines lock mutexes in opposite order.
func DeadlockMutexOrdering() {
	var muA, muB sync.Mutex
	done := make(chan bool, 2)

	// Goroutine 1: locks A then B
	go func() {
		muA.Lock()
		// time.Sleep(time.Millisecond) // Increase chance of deadlock
		muB.Lock()
		muB.Unlock()
		muA.Unlock()
		done <- true
	}()

	// Goroutine 2: locks B then A (opposite order!)
	go func() {
		muB.Lock()
		// time.Sleep(time.Millisecond)
		muA.Lock() // DEADLOCK: G1 has A, waiting for B; G2 has B, waiting for A
		muA.Unlock()
		muB.Unlock()
		done <- true
	}()

	<-done
	<-done
}

// FixMutexOrdering should work correctly.
//
// TODO: Fix by ensuring consistent lock ordering
// RULE: Always acquire multiple locks in the same order
func FixMutexOrdering() {
	// YOUR CODE HERE
}

// =============================================================================
// PART 4: Channel + WaitGroup Interaction
// =============================================================================

// DeadlockChannelWaitGroup demonstrates a subtle deadlock.
// The channel fills up, blocking sends, so Done() is never called.
func DeadlockChannelWaitGroup() []int {
	var wg sync.WaitGroup
	ch := make(chan int, 2) // Small buffer

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ch <- n // DEADLOCK: blocks when buffer full
		}(i)
	}

	wg.Wait()   // Waits for all goroutines
	close(ch)   // Close after all sends done

	var results []int
	for v := range ch {
		results = append(results, v)
	}
	return results
}

// FixChannelWaitGroup should collect all 10 values.
//
// TODO: Fix by restructuring the wait/receive logic
// HINT: Can't wait for senders if receiver is blocked too!
func FixChannelWaitGroup() []int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 5: Select Deadlock
// =============================================================================

// DeadlockEmptySelect demonstrates select with no cases.
func DeadlockEmptySelect() {
	select {} // DEADLOCK: blocks forever
}

// DeadlockSelectNilChannels demonstrates all-nil channels.
func DeadlockSelectNilChannels() {
	var ch1, ch2 chan int // Both nil
	select {
	case <-ch1: // Never ready (nil channel)
	case <-ch2: // Never ready (nil channel)
	}
	// DEADLOCK: no case can proceed
}

// FixSelectNilChannels should not deadlock.
//
// TODO: Fix using default or proper channel initialization
func FixSelectNilChannels() string {
	// YOUR CODE HERE
	return ""
}

// =============================================================================
// CHALLENGE: Debug this production-like code
// =============================================================================

// ProcessItems has a deadlock under certain conditions.
// Find and fix it!
//
// HINT: Think about what happens when len(items) > buffer size
func ProcessItems(items []int, processor func(int) int) []int {
	results := make(chan int, 5) // Fixed buffer
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results <- processor(n)
		}(item)
	}

	// Wait for all processing to complete, then collect results
	wg.Wait()
	close(results)

	var output []int
	for r := range results {
		output = append(output, r)
	}
	return output
}

// ProcessItemsFix should work for any number of items.
//
// TODO: Fix the deadlock
func ProcessItemsFix(items []int, processor func(int) int) []int {
	// YOUR CODE HERE
	return nil
}

// Ensure sync import is used
var _ = sync.Mutex{}
