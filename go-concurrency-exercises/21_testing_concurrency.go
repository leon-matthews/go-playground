package concurrency

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// EXERCISE 6.2: Testing Concurrent Code
// =============================================================================
//
// Testing concurrent code is challenging because:
// - Race conditions are non-deterministic
// - Deadlocks may only occur under specific timing
// - Goroutine leaks are easy to miss
// - Tests may pass locally but fail in CI
//
// KEY CONCEPTS:
// - go test -race: always run with race detector
// - go test -count=N: run multiple times to catch flaky tests
// - t.Parallel(): run tests concurrently
// - goleak: detect goroutine leaks
// - Timeouts: prevent hanging tests
//
// =============================================================================

// =============================================================================
// PART 1: Race Detector
// =============================================================================

// Example of code with a race condition.
// Run: go test -race -run TestRacyIncrement
type RacyCounter struct {
	value int
}

func (c *RacyCounter) Increment() {
	c.value++ // RACE: not thread-safe
}

func (c *RacyCounter) Value() int {
	return c.value // RACE: not thread-safe
}

// TODO: Implement a test that DETECTS the race condition.
// The test should pass normally but fail with -race flag.
func TestRacyIncrement(t *testing.T) {
	// YOUR CODE HERE
}

// SafeCounter is a fixed version.
type SafeCounter2 struct {
	value int64
}

func (c *SafeCounter2) Increment() {
	atomic.AddInt64(&c.value, 1)
}

func (c *SafeCounter2) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// TODO: Implement a test that verifies SafeCounter2 is race-free.
func TestSafeCounter(t *testing.T) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 2: Testing with t.Parallel()
// =============================================================================

// t.Parallel() allows tests to run concurrently.
// This can expose race conditions in test setup/teardown.

// SharedTestResource is used by multiple tests.
var sharedTestResource = make(map[string]int)
var resourceMu sync.Mutex

// TODO: Implement parallel tests that safely use shared resource.
func TestParallelSafe1(t *testing.T) {
	t.Parallel()
	// YOUR CODE HERE - use resourceMu to protect access
}

func TestParallelSafe2(t *testing.T) {
	t.Parallel()
	// YOUR CODE HERE - use resourceMu to protect access
}

// =============================================================================
// PART 3: Goroutine Leak Detection
// =============================================================================

// LeakingFunction creates goroutines that don't exit.
func LeakingFunction2() {
	ch := make(chan int)
	go func() {
		<-ch // Blocks forever - leak!
	}()
	// Returns without closing ch or sending
}

// NonLeakingFunction properly manages goroutines.
func NonLeakingFunction() {
	ch := make(chan int)
	done := make(chan struct{})

	go func() {
		select {
		case <-ch:
		case <-done:
		}
	}()

	close(done) // Signal goroutine to exit
}

// LeakChecker helps detect goroutine leaks in tests.
//
// TODO: Implement a test helper that:
// 1. Records goroutine count before test
// 2. Runs the test
// 3. Waits briefly for goroutines to exit
// 4. Checks goroutine count after
// 5. Fails test if count increased
type LeakChecker struct {
	t        *testing.T
	before   int
	maxWait  time.Duration
	checkInterval time.Duration
}

func NewLeakChecker(t *testing.T) *LeakChecker {
	// YOUR CODE HERE
	return nil
}

func (lc *LeakChecker) Check() {
	// YOUR CODE HERE
}

// TODO: Implement tests using LeakChecker.
func TestNoLeak(t *testing.T) {
	// YOUR CODE HERE
}

func TestDetectsLeak(t *testing.T) {
	// This test should demonstrate leak detection
	// (but we can't actually fail it without causing issues)
	t.Skip("Demonstrates leak detection - skipping to avoid test failure")
}

// =============================================================================
// PART 4: Testing Timeouts and Deadlines
// =============================================================================

// SlowOperation simulates a slow operation.
func SlowOperation(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TODO: Implement a test that verifies timeout behavior.
// The test should complete quickly (not wait 5 seconds).
func TestSlowOperationTimeout(t *testing.T) {
	// YOUR CODE HERE
}

// WithTestTimeout runs a test function with a timeout.
// If the function doesn't complete in time, it fails the test.
//
// TODO: Implement a test helper for timeouts.
func WithTestTimeout(t *testing.T, timeout time.Duration, fn func()) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 5: Testing Deadlock Scenarios
// =============================================================================

// PotentialDeadlock has a deadlock under certain conditions.
func PotentialDeadlock(ch1, ch2 chan int) {
	ch1 <- 1 // Blocks if ch1 is unbuffered and no receiver
	<-ch2    // Blocks if nothing sent to ch2
}

// TODO: Implement a test that detects the potential deadlock.
// Use timeout to prevent test from hanging forever.
func TestDeadlockDetection(t *testing.T) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 6: Stress Testing Concurrent Code
// =============================================================================

// ConcurrentMap is a thread-safe map wrapper.
type ConcurrentMap struct {
	mu   sync.RWMutex
	data map[string]int
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{data: make(map[string]int)}
}

func (m *ConcurrentMap) Set(key string, value int) {
	m.mu.Lock()
	m.data[key] = value
	m.mu.Unlock()
}

func (m *ConcurrentMap) Get(key string) (int, bool) {
	m.mu.RLock()
	v, ok := m.data[key]
	m.mu.RUnlock()
	return v, ok
}

// TODO: Implement stress tests for ConcurrentMap.
// Run many goroutines doing concurrent reads and writes.
func TestConcurrentMapStress(t *testing.T) {
	// YOUR CODE HERE
}

// StressTest runs a function under stress conditions.
//
// TODO: Implement a stress test helper that:
// 1. Runs fn concurrently from numGoroutines
// 2. Each goroutine runs fn numIterations times
// 3. Fails test if any panic or error occurs
// 4. Reports timing statistics
func StressTest(t *testing.T, numGoroutines, numIterations int, fn func() error) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 7: Testing Channel Behavior
// =============================================================================

// ChannelTestHelper provides utilities for testing channels.
type ChannelTestHelper[T any] struct {
	ch      chan T
	timeout time.Duration
}

func NewChannelTestHelper[T any](ch chan T, timeout time.Duration) *ChannelTestHelper[T] {
	return &ChannelTestHelper[T]{ch: ch, timeout: timeout}
}

// ExpectValue waits for a value and compares it.
func (h *ChannelTestHelper[T]) ExpectValue(t *testing.T, expected T, compare func(T, T) bool) {
	// YOUR CODE HERE
}

// ExpectClosed verifies the channel is closed.
func (h *ChannelTestHelper[T]) ExpectClosed(t *testing.T) {
	// YOUR CODE HERE
}

// ExpectNoValue verifies nothing is received within timeout.
func (h *ChannelTestHelper[T]) ExpectNoValue(t *testing.T) {
	// YOUR CODE HERE
}

// =============================================================================
// PART 8: Deterministic Testing with Controlled Scheduling
// =============================================================================

// Sometimes you need deterministic testing of concurrent code.
// This is hard because goroutine scheduling is non-deterministic.

// SequentialExecutor runs functions in a controlled order.
// Useful for testing specific interleavings.
//
// TODO: Implement an executor that:
// 1. Registers named operations
// 2. Runs operations in specified order
// 3. Blocks each operation until it's "their turn"
type SequentialExecutor struct {
	// YOUR FIELDS HERE
}

func NewSequentialExecutor() *SequentialExecutor {
	// YOUR CODE HERE
	return nil
}

// Register registers an operation that will run when allowed.
func (e *SequentialExecutor) Register(name string, op func()) {
	// YOUR CODE HERE
}

// Allow permits the named operation to proceed.
func (e *SequentialExecutor) Allow(name string) {
	// YOUR CODE HERE
}

// Wait waits for operation to complete.
func (e *SequentialExecutor) Wait(name string) {
	// YOUR CODE HERE
}

// =============================================================================
// CHALLENGE: Comprehensive Test Suite
// =============================================================================

// WorkerPool3 is a worker pool to test.
type WorkerPool3 struct {
	workers   int
	jobs      chan func()
	done      chan struct{}
	closeOnce sync.Once
}

func NewWorkerPool3(workers int) *WorkerPool3 {
	wp := &WorkerPool3{
		workers: workers,
		jobs:    make(chan func(), 100),
		done:    make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		go wp.worker()
	}
	return wp
}

func (wp *WorkerPool3) worker() {
	for {
		select {
		case job := <-wp.jobs:
			job()
		case <-wp.done:
			return
		}
	}
}

func (wp *WorkerPool3) Submit(job func()) {
	wp.jobs <- job
}

func (wp *WorkerPool3) Close() {
	wp.closeOnce.Do(func() {
		close(wp.done)
	})
}

// TODO: Implement a comprehensive test suite for WorkerPool3.
// Include tests for:
// 1. Basic functionality (jobs execute)
// 2. Concurrent job submission (race-free)
// 3. No goroutine leaks after Close()
// 4. Graceful shutdown (in-flight jobs complete)
// 5. Stress test with many jobs
// 6. Panic handling (one job panic doesn't kill pool)
func TestWorkerPool3Suite(t *testing.T) {
	// YOUR CODE HERE
}

// Ensure imports are used
var _ = context.Background
var _ = runtime.NumGoroutine
var _ = sync.Mutex{}
var _ = atomic.AddInt64
var _ = testing.T{}
var _ = time.Second
