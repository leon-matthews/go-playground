package concurrency

// =============================================================================
// EXERCISE 3.2: Fan-Out / Fan-In Pattern
// =============================================================================
//
// Fan-out: Multiple goroutines read from the same channel (distribute work)
// Fan-in: Multiple channels are merged into a single channel (collect results)
//
// KEY CONCEPTS:
// - Fan-out distributes work across workers for parallelism
// - Fan-in collects results from multiple sources
// - Order may not be preserved (unless you track it)
// - Need to handle closure of all channels properly
//
// =============================================================================

// =============================================================================
// PART 1: Basic Fan-Out
// =============================================================================

// FanOut distributes values from input to n output channels.
// Values are sent round-robin to outputs.
//
// TODO: Implement to:
// 1. Create n output channels
// 2. Launch a goroutine that:
//    - Reads from input
//    - Sends to outputs in round-robin fashion
//    - Closes all outputs when input is closed
// 3. Return the output channels
//
// QUESTION: What's the difference between round-robin and "first available"?
func FanOut(input <-chan int, n int) []<-chan int {
	// YOUR CODE HERE
	return nil
}

// FanOutFirstAvailable distributes values to whichever worker is free.
// This maximizes throughput when workers have varying processing times.
//
// TODO: Implement using select to send to first available channel
// HINT: You can't select on a dynamic number of channels directly.
// Consider having workers pull from a shared channel instead.
func FanOutFirstAvailable(input <-chan int, n int) []<-chan int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 2: Basic Fan-In
// =============================================================================

// FanIn merges multiple input channels into a single output channel.
//
// TODO: Implement to:
// 1. Create output channel
// 2. For each input, launch a goroutine that forwards values to output
// 3. Use WaitGroup to close output only after ALL inputs are closed
// 4. Return output channel
func FanIn(inputs ...<-chan int) <-chan int {
	// YOUR CODE HERE
	return nil
}

// FanInOrdered merges inputs but preserves which channel each value came from.
// Returns channel of (value, sourceIndex) pairs.
//
// TODO: Implement to track source of each value
type TaggedValue struct {
	Value       int
	SourceIndex int
}

func FanInOrdered(inputs ...<-chan int) <-chan TaggedValue {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 3: Fan-Out/Fan-In Pipeline
// =============================================================================

// ParallelProcess distributes work across n workers and collects results.
// This is the classic "scatter-gather" pattern.
//
// TODO: Implement to:
// 1. Fan out input to n workers
// 2. Each worker applies processFn to its values
// 3. Fan in all worker outputs to single result channel
// 4. Return result channel
//
// QUESTION: Are results in the same order as inputs? Why or why not?
func ParallelProcess(input <-chan int, n int, processFn func(int) int) <-chan int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 4: Order-Preserving Parallel Process
// =============================================================================

// OrderedResult pairs a result with its original sequence number.
type OrderedResult struct {
	SeqNum int
	Value  int
}

// ParallelProcessOrdered processes in parallel but returns results in order.
//
// TODO: Implement to:
// 1. Tag each input with a sequence number
// 2. Process in parallel (order of processing doesn't matter)
// 3. Reorder results by sequence number before outputting
//
// HINT: You might need a buffer to hold out-of-order results while
// waiting for earlier sequence numbers.
//
// QUESTION: What's the memory implication of this approach?
func ParallelProcessOrdered(input <-chan int, n int, processFn func(int) int) <-chan int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// PART 5: Bounded Fan-In
// =============================================================================

// BoundedFanIn merges inputs but limits output buffer.
// If consumer is slow, it applies backpressure to producers.
//
// TODO: Implement with bounded output channel
// QUESTION: How does this affect the input goroutines?
func BoundedFanIn(bufferSize int, inputs ...<-chan int) <-chan int {
	// YOUR CODE HERE
	return nil
}

// =============================================================================
// CHALLENGE: Implement a map-reduce pattern
// =============================================================================

// MapReduce applies map to inputs in parallel, then reduces results.
//
// TODO: Implement classic map-reduce:
// 1. Fan out inputs to n mappers
// 2. Each mapper applies mapFn
// 3. Fan in mapper outputs
// 4. Apply reduceFn to combine all values into single result
//
// Example: Sum of squares
//   MapReduce(input, 4, func(x int) int { return x*x }, func(a,b int) int { return a+b })
func MapReduce(input <-chan int, n int, mapFn func(int) int, reduceFn func(int, int) int, initial int) int {
	// YOUR CODE HERE
	return 0
}

// =============================================================================
// CHALLENGE: Implement batched fan-in
// =============================================================================

// BatchedFanIn collects values and emits them in batches.
//
// TODO: Implement to:
// - Collect values from all inputs
// - Emit a batch when batchSize values collected OR timeout elapsed
// - Handle closure of all inputs
//
// HINT: Use select with time.After for timeout
func BatchedFanIn(batchSize int, timeout int, inputs ...<-chan int) <-chan []int {
	// YOUR CODE HERE
	return nil
}
