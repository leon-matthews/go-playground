package concurrency

import (
	"fmt"
)

// =============================================================================
// EXERCISE 1.2: Pipelines with Directional Channels
// =============================================================================
//
// Pipelines are a powerful pattern where data flows through stages connected
// by channels. Each stage:
// 1. Receives values from upstream via inbound channels
// 2. Performs some function on that data
// 3. Sends values downstream via outbound channels
//
// KEY CONCEPTS:
// - chan<- T is a send-only channel (can only send to it)
// - <-chan T is a receive-only channel (can only receive from it)
// - These are COMPILE-TIME safety features
// - The stage that creates values should close the channel when done
// - Closing propagates through the pipeline naturally with range
//
// =============================================================================

// =============================================================================
// PART 1: Basic Three-Stage Pipeline
// =============================================================================

// Generate sends integers from start to end (exclusive) on the returned channel.
// This is the SOURCE stage of a pipeline.
//
// 1. Create a channel
// 2. Launch a goroutine that sends values start, start+1, ... end-1
// 3. Close the channel when done (inside the goroutine!)
// 4. Return the channel immediately (don't wait for goroutine)
//
// QUESTION: Why do we return <-chan int instead of chan int?
// ANSWER: To restrict what the consumer can do with the channel, eg. prevent closing
func Generate(start, end int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i < end; i++ {
			out <- i
		}
		close(out)
	}()
	return out
}

// Square receives integers, squares them, and sends results.
// This is a TRANSFORM stage of a pipeline.
//
// 1. Create an output channel
// 2. Launch a goroutine that:
//   - Ranges over the input channel
//   - Sends the square of each value to output
//   - Closes output when input is exhausted
//
// 3. Return the output channel immediately
//
// QUESTION: What happens if you forget to close the output channel?
// ANSWER: The consumer won't be informed that we have reached end of data
func Square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

// Sum receives integers and returns their sum.
// This is a SINK stage of a pipeline.
//
// 1. Range over the input channel
// 2. Accumulate the sum
// 3. Return the total
//
// NOTE: This function blocks until the channel is closed!
func Sum(in <-chan int) int {
	sum := 0
	for n := range in {
		sum += n
	}
	return sum
}

// RunPipeline connects the stages: Generate -> Square -> Sum
//
// Example: RunPipeline(1, 4) should compute 1^2 + 2^2 + 3^2 = 1 + 4 + 9 = 14
func RunPipeline(start, end int) int {
	nums := Generate(start, end)
	squares := Square(nums)
	sum := Sum(squares)
	return sum
}

// =============================================================================
// PART 2: Pipeline with Multiple Transforms
// =============================================================================

// Filter returns only values that pass the predicate function.
// 1. Creates an output channel
// 2. Launches a goroutine that only forwards values where pred(v) is true
// 3. Closes output when input is exhausted
func Filter(in <-chan int, pred func(int) bool) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			if pred(n) {
				out <- n
			}
		}
		close(out)
	}()
	return out
}

// Map applies a function to each value.
//
// 1. Creates an output channel
// 2. Launches a goroutine that sends fn(v) for each v
// 3. Closes output when input is exhausted
func Map(in <-chan int, fn func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- fn(n)
		}
		close(out)
	}()
	return out
}

// RunFilterMapPipeline demonstrates chaining multiple transforms.
//
// 1. Generates numbers from 1 to 10
// 2. Filters to keep only even numbers
// 3. Maps each to its square
// 4. Returns the sum
//
// Expected: 2^2 + 4^2 + 6^2 + 8^2 = 4 + 16 + 36 + 64 = 120
func RunFilterMapPipeline() int {
	nums := Generate(1, 10)
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	squares := Map(evens, func(n int) int { return n * n })
	sum := Sum(squares)
	return sum
}

// =============================================================================
// PART 3: Pipeline with Error Handling
// =============================================================================

// Result holds either a value or an error
type Result struct {
	Value int
	Err   error
}

// GenerateWithError sends Results, occasionally producing errors.
//
// 1. Create a channel of Result
// 2. For each value from start to end-1:
//   - If the value is divisible by errEvery, send a Result with Err set
//   - Otherwise send a Result with Value set
//
// 3. Close the channel when done
func GenerateWithError(start, end, errEvery int) <-chan Result {
	out := make(chan Result)
	go func() {
		for i := start; i < end; i++ {
			if i%errEvery == 0 {
				out <- Result{i, fmt.Errorf("%d is divisible by %d", i, errEvery)}
			} else {
				out <- Result{i, nil}
			}
		}
		close(out)
	}()
	return out
}

// ProcessResults demonstrates handling errors in a pipeline.
//
// 1. Range over the input channel
// 2. Skip values that have errors (count them)
// 3. Sum the successful values
// 4. Return (sum, errorCount)
func ProcessResults(in <-chan Result) (sum int, errorCount int) {
	for r := range in {
		if r.Err != nil {
			errorCount++
			continue
		}
		sum += r.Value
	}
	return sum, errorCount
}

// =============================================================================
// CHALLENGE: Implement a pipeline that can be cancelled
// =============================================================================

// GenerateCancellable sends integers but stops if done channel is closed.
//
// 1. Create an output channel
// 2. Launch a goroutine that:
//   - Uses select to either send the next value OR detect done
//   - Stops immediately if done is closed
//   - Closes output when finished (either done or reached end)
//
// 3. Return the output channel
//
// HINT: This is a preview of context.Context cancellation patterns!
func GenerateCancellable(start, end int, done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := start; i < end; i++ {
			select {
			case out <- i:
			case <-done:
				return
			}
		}
	}()
	return out
}

// TransformCancellable applies a transform but respects cancellation.
func TransformCancellable(in <-chan int, fn func(int) int, done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- fn(n):
			case <-done:
				return
			}
		}
	}()
	return out
}
