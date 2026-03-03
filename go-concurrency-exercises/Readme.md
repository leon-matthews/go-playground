# Go Concurrency Exercises

A comprehensive set of exercises to master Go concurrency patterns, from fundamentals to production-ready code.

## How to Use

1. Each file contains exercises with `// YOUR CODE HERE` markers
2. Read the comments carefully - they explain concepts and provide hints
3. Implement the functions to make tests pass
4. Run with race detector: `go test -race ./...`

## Exercise Overview

### Fundamentals (01-07)

| File | Topic | Concepts |
|------|-------|----------|
| `01_channel_semantics.go` | Channel Basics | Buffered vs unbuffered, closing, nil channels, directional channels |
| `02_pipeline.go` | Pipelines | Generator/transform/sink stages, channel direction, error handling |
| `03_select.go` | Select Statement | Multiplexing, timeouts, non-blocking operations, priority |
| `04_waitgroup.go` | WaitGroup | Coordinating goroutines, common pitfalls |
| `05_mutex.go` | Mutex & RWMutex | Critical sections, reader-writer locks, lock granularity |
| `06_once_cond.go` | Once & Cond | One-time initialization, condition variables, broadcast |
| `07_atomic.go` | Atomic Operations | Compare-and-swap, atomic counters, lock-free patterns |

### Patterns (08-12)

| File | Topic | Concepts |
|------|-------|----------|
| `08_worker_pool.go` | Worker Pool | Fixed workers, job queues, graceful shutdown, dynamic scaling |
| `09_fan_out_fan_in.go` | Fan-Out/Fan-In | Work distribution, result collection, ordered processing, map-reduce |
| `10_context.go` | Context | Cancellation, timeouts, deadlines, values, cascading cancel |
| `11_rate_limiting.go` | Rate Limiting | Token bucket, leaky bucket, per-client limits |
| `12_semaphore.go` | Semaphore | Bounded concurrency, weighted semaphore, resource pools |

### Debugging (13-15)

| File | Topic | Concepts |
|------|-------|----------|
| `13_race_conditions.go` | Race Conditions | Detection, fixing, subtle races (check-then-act, loop variable) |
| `14_deadlocks.go` | Deadlocks | Channel deadlocks, mutex ordering, WaitGroup misuse |
| `15_goroutine_leaks.go` | Goroutine Leaks | Detection, abandoned channels, forgotten timers, leak prevention |

### Advanced (16-21)

| File | Topic | Concepts |
|------|-------|----------|
| `16_sync_pool.go` | sync.Pool | Object reuse, buffer pools, reducing GC pressure |
| `17_errgroup.go` | errgroup | Goroutine groups, error propagation, limited concurrency |
| `18_singleflight.go` | singleflight | Request deduplication, thundering herd prevention, cache stampede |
| `19_memory_model.go` | Memory Model | Happens-before, visibility, safe publication, lock-free structures |
| `20_real_world.go` | Production Patterns | HTTP handlers, DB pools, graceful shutdown, background jobs |
| `21_testing_concurrency.go` | Testing | Race detector, leak detection, stress testing, deterministic tests |

## Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector (always do this!)
go test -race ./...

# Run specific exercise
go test -race -run TestChannel

# Run multiple times to catch flaky tests
go test -race -count=10 ./...

# Verbose output
go test -race -v ./...
```

## Key Concepts by Difficulty

### Beginner
- Goroutines and channels
- Buffered vs unbuffered channels
- WaitGroup for coordination
- Basic mutex usage

### Intermediate
- Select statement and timeouts
- Context for cancellation
- Worker pools
- Fan-out/fan-in patterns
- Race condition detection

### Advanced
- Memory model and happens-before
- Lock-free data structures
- sync.Pool for performance
- singleflight for deduplication
- Production patterns (graceful shutdown, connection pooling)

## Common Pitfalls Covered

1. **Goroutine leaks** - Always ensure goroutines can exit
2. **Race conditions** - Use `-race` flag religiously
3. **Deadlocks** - Consistent lock ordering, avoid nested locks
4. **Loop variable capture** - Copy loop variables in closures
5. **Closing channels** - Only sender should close, never close twice
6. **Nil channels** - Block forever in select (useful for disabling cases)
7. **WaitGroup misuse** - Call Add() before launching goroutine
8. **Context values** - Use sparingly, prefer explicit parameters

## Resources

- [Go Memory Model](https://go.dev/ref/mem)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Advanced Concurrency Patterns](https://go.dev/blog/io2013-talk-concurrency)
- [Context Package](https://go.dev/blog/context)
