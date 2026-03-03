package concurrency

// =============================================================================
// EXERCISE 3.1: Worker Pool Pattern
// =============================================================================
//
// A worker pool manages a fixed number of goroutines that process jobs.
// This is essential for controlling resource usage and preventing goroutine explosion.
//
// KEY CONCEPTS:
// - Fixed number of workers prevents unbounded goroutine creation
// - Jobs are sent to a channel, workers receive and process them
// - Results can be collected via a separate channel
// - Graceful shutdown requires coordinating job submission and worker termination
//
// =============================================================================

// =============================================================================
// PART 1: Basic Worker Pool
// =============================================================================

// Job represents a unit of work.
type Job struct {
	ID      int
	Payload int
}

// Result represents the output of processing a job.
type JobResult struct {
	JobID  int
	Output int
	Err    error
}

// WorkerPool manages a pool of worker goroutines.
//
// TODO: Implement a worker pool that:
// - Has a configurable number of workers
// - Accepts jobs via a channel
// - Returns results via a channel
// - Supports graceful shutdown
type WorkerPool struct {
	// YOUR FIELDS HERE
}

// NewWorkerPool creates a pool with n workers.
// The process function is called for each job.
//
// TODO: Implement to:
// 1. Create job and result channels
// 2. Start n worker goroutines
// 3. Each worker loops: receive job, process, send result
// 4. Workers should exit when jobs channel is closed
func NewWorkerPool(n int, process func(Job) JobResult) *WorkerPool {
	// YOUR CODE HERE
	return nil
}

// Submit sends a job to the pool.
// Returns false if pool is shut down.
func (wp *WorkerPool) Submit(job Job) bool {
	// YOUR CODE HERE
	return false
}

// Results returns the channel to receive results from.
func (wp *WorkerPool) Results() <-chan JobResult {
	// YOUR CODE HERE
	return nil
}

// Shutdown gracefully stops the pool.
// Waits for all in-flight jobs to complete.
//
// TODO: Implement to:
// 1. Close the jobs channel (no more submissions)
// 2. Wait for all workers to finish
// 3. Close the results channel
func (wp *WorkerPool) Shutdown() {
	// YOUR CODE HERE
}

// =============================================================================
// PART 2: Worker Pool with Job Queue
// =============================================================================

// BufferedPool has an internal job queue for backpressure handling.
//
// TODO: Implement a pool where:
// - Submit never blocks (adds to internal queue)
// - Workers pull from the queue
// - Queue has maximum size (reject jobs when full)
//
// QUESTION: How would you handle the case when the queue is full?
type BufferedPool struct {
	// YOUR FIELDS HERE
}

// NewBufferedPool creates a pool with n workers and queue of given size.
// func NewBufferedPool(workers, queueSize int, process func(Job) JobResult) *BufferedPool {
// 	// YOUR CODE HERE
// 	return nil
// }

// =============================================================================
// PART 3: Dynamic Worker Pool
// =============================================================================

// DynamicPool adjusts worker count based on load.
//
// TODO: Implement a pool that:
// - Starts with minWorkers
// - Scales up to maxWorkers when jobs are queued
// - Scales down when idle for a period
//
// HINT: This is quite advanced! Consider:
// - How to detect load (queue depth)
// - How to spawn new workers safely
// - How to signal workers to exit
type DynamicPool struct {
	// YOUR FIELDS HERE
}

// =============================================================================
// CHALLENGE: Implement a pool with priorities
// =============================================================================

// PriorityJob includes a priority level.
type PriorityJob struct {
	Job
	Priority int // Higher = more urgent
}

// PriorityPool processes higher priority jobs first.
//
// TODO: Implement a pool that:
// - Maintains separate queues per priority level
// - Workers always take from highest priority non-empty queue
// - Lower priority jobs don't starve (optional: implement aging)
//
// HINT: You might need a heap or multiple channels
type PriorityPool struct {
	// YOUR FIELDS HERE
}
