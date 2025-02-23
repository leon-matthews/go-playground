// === Goroutines ===
// 1. Launch a new goroutine that prints "Hello from Goroutine!".
// 2. Modify the goroutine to run in a loop and print numbers from 1 to 10.
// 3. Use `time.Sleep` to wait for a goroutine to complete (not the best practice).

// === WaitGroups (`sync.WaitGroup`) ===
// 4. Use `sync.WaitGroup` to wait for multiple goroutines to finish.
// 5. Launch three goroutines that each print a message, and use `WaitGroup` to wait for them.

// === Mutexes (`sync.Mutex`) ===
// 6. Use a `sync.Mutex` to safely increment a shared counter in multiple goroutines.
// 7. Convert the above mutex example into a `sync.RWMutex` for read-write efficiency.

// === Unbuffered and Buffered Channels ===
// 8. Create an unbuffered channel and send a value from one goroutine to another.
// 9. Create a buffered channel with capacity 3 and send three values before receiving.
// 10. Use `close()` on a channel and demonstrate how to receive from a closed channel.

// === Channel Direction (`chan<-` & `<-chan`) ===
// 11. Write a function that only sends values to a channel (`chan<-`).
// 12. Write a function that only receives values from a channel (`<-chan`).

// === Select Statements with Channels ===
// 13. Use a `select` statement to read from two channels simultaneously.
// 14. Modify the `select` statement to include a `default` case.
// 15. Implement a timeout mechanism using `select` and `time.After`.

// === Worker Pools ===
// 16. Implement a worker pool where multiple workers process jobs from a channel.
// 17. Modify the worker pool to return results via another channel.

// === Concurrency Patterns ===
// 18. Implement the **Fan-In** pattern: multiple goroutines send data to a single channel.
// 19. Implement the **Fan-Out** pattern: one channel distributes work to multiple goroutines.
// 20. Implement a **Pipeline** where one stage processes data and passes it to the next.

// === Avoiding Data Races ===
// 21. Write a function that encounters a **race condition** (without a mutex).
// 22. Fix the race condition using a `sync.Mutex`.
// 23. Fix the race condition using a **channel instead of a mutex**.

// === Advanced Synchronization ===
// 24. Use `sync.Once` to ensure a function runs only once in multiple goroutines.
// 25. Use `sync.Pool` to manage reusable objects efficiently.
// 26. Use `sync.Cond` to synchronize multiple goroutines waiting for a condition.

// === Performance Considerations ===
// 27. Compare performance between **goroutines + channels** vs. **mutexes**.
// 28. Explain when **using a worker pool** improves efficiency.
// 29. Demonstrate how **excessive goroutines** can lead to high memory usage.

// === Context Package ===
// 30. Use `context.WithCancel()` to cancel a long-running goroutine.
// 31. Use `context.WithTimeout()` to set a time limit on a goroutine.
// 32. Use `context.WithDeadline()` to cancel a goroutine at a specific time.
// 33. Pass `context.Context` into a function to propagate cancellation signals.

// === Worker pool with dynamic scaling ===
// 34. Implement a worker pool that scales dynamically based on job queue size.
// 35. Modify the worker pool to **gracefully shut down** when jobs are complete.
// 36. Use a channel to **signal when to scale up/down** workers.

// === Rate Limiting with Goroutines ===
// 37. Implement a rate limiter using `time.Ticker`.
// 38. Use `golang.org/x/time/rate.Limiter` to control request flow.
// 39. Apply rate limiting to an HTTP API endpoint.

// === Goroutine Leaks ===
// 40. Create a goroutine leak by spawning goroutines without an exit condition.
// 41. Detect goroutine leaks using `pprof`.
// 42. Fix a goroutine leak by ensuring proper cancellation with `context.Context`.

// === Orchestration with error handling with Goroutines ===
// 43. Run multiple goroutines using `errgroup.Group` and return the first error.
// 44. Use `errgroup.WithContext()` to cancel all goroutines if one fails.
// 45. Compare `errgroup` with `sync.WaitGroup` for error handling.

// === Actor Model with Goroutines ===
// 46. Implement an actor system where each actor handles messages via a channel.
// 47. Implement an actor that processes different message types (start, stop, status).
// 48. Modify the actor to allow graceful shutdown.

// === Atomic Operations ===
// 49. Use `sync/atomic.AddInt64()` to increment a counter safely without a mutex.
// 50. Implement a read-mostly counter using `sync/atomic.LoadInt64()`.
// 51. Compare the performance of `sync/atomic` vs `sync.Mutex` in concurrent writes.

// === Select with timeouts, hearbeats and fan-in
// 52. Implement a `select` statement with a timeout (`time.After`).
// 53. Implement a periodic heartbeat using `time.Ticker` inside `select`.
// 54. Implement a **Fan-In** pattern: merging multiple channels into one.

// === Advanced Concurrency Patterns ===
// 55. Implement a **bounded work queue** where jobs are dropped if the queue is full.
// 56. Implement **priority scheduling** using multiple channels (high, medium, low priority).
