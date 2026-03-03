package concurrency

import (
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 1.3: Select Statement
// Run with: go test -v -run Test03
// =============================================================================

func TestFirstResponse(t *testing.T) {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	ch1 <- "first"
	// ch2 is empty

	result := FirstResponse(ch1, ch2)
	if result != "first" {
		t.Errorf("FirstResponse() = %q; want \"first\"", result)
	}
}

func TestFirstResponse_Ch2First(t *testing.T) {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	ch2 <- "second"
	// ch1 is empty

	result := FirstResponse(ch1, ch2)
	if result != "second" {
		t.Errorf("FirstResponse() = %q; want \"second\"", result)
	}
}

func TestMergeChannels(t *testing.T) {
	ch1 := make(chan int, 3)
	ch2 := make(chan int, 3)

	ch1 <- 1
	ch1 <- 3
	ch2 <- 2
	ch2 <- 4
	close(ch1)
	close(ch2)

	output := MergeChannels(ch1, ch2)
	if output == nil {
		t.Fatal("MergeChannels returned nil")
	}

	var values []int
	for v := range output {
		values = append(values, v)
	}

	if len(values) != 4 {
		t.Errorf("MergeChannels produced %d values; want 4", len(values))
	}

	// Values should be 1, 2, 3, 4 in some order
	sum := 0
	for _, v := range values {
		sum += v
	}
	if sum != 10 {
		t.Errorf("Sum of merged values = %d; want 10", sum)
	}
}

func TestReceiveWithTimeout_Success(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42

	value, ok := ReceiveWithTimeout(ch, 100*time.Millisecond)

	if !ok {
		t.Error("ReceiveWithTimeout returned false; expected true")
	}
	if value != 42 {
		t.Errorf("value = %d; want 42", value)
	}
}

func TestReceiveWithTimeout_Timeout(t *testing.T) {
	ch := make(chan int) // unbuffered, no sender

	start := time.Now()
	value, ok := ReceiveWithTimeout(ch, 50*time.Millisecond)
	elapsed := time.Since(start)

	if ok {
		t.Errorf("ReceiveWithTimeout returned true with value %d; expected timeout", value)
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("Returned too fast (%v); expected ~50ms timeout", elapsed)
	}
}

func TestReceiveWithDeadline(t *testing.T) {
	ch := make(chan int, 10)
	for i := 1; i <= 5; i++ {
		ch <- i
	}

	deadline := time.Now().Add(50 * time.Millisecond)
	values := ReceiveWithDeadline(ch, deadline)

	// Should get at least the 5 buffered values
	if len(values) < 5 {
		t.Errorf("Got %d values; want at least 5", len(values))
	}
}

func TestPeriodicTask(t *testing.T) {
	done := make(chan struct{})
	count := 0
	mu := sync.Mutex{}

	fn := func() {
		mu.Lock()
		count++
		mu.Unlock()
	}

	go func() {
		time.Sleep(75 * time.Millisecond)
		close(done)
	}()

	result := PeriodicTask(fn, 20*time.Millisecond, done)

	// With 20ms interval and 75ms runtime, should run ~3-4 times
	if result < 2 || result > 5 {
		t.Errorf("PeriodicTask ran %d times; expected 2-5", result)
	}
}

func TestTrySend_Success(t *testing.T) {
	ch := make(chan int, 1)

	result := TrySend(ch, 42)

	if !result {
		t.Error("TrySend returned false; expected true (buffer has space)")
	}

	// Verify value was sent
	if v := <-ch; v != 42 {
		t.Errorf("Channel received %d; want 42", v)
	}
}

func TestTrySend_Full(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1 // fill buffer

	result := TrySend(ch, 42)

	if result {
		t.Error("TrySend returned true; expected false (buffer full)")
	}
}

func TestTryReceive_Success(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42

	value, ok := TryReceive(ch)

	if !ok {
		t.Error("TryReceive returned false; expected true")
	}
	if value != 42 {
		t.Errorf("value = %d; want 42", value)
	}
}

func TestTryReceive_Empty(t *testing.T) {
	ch := make(chan int, 1) // empty buffer

	value, ok := TryReceive(ch)

	if ok {
		t.Errorf("TryReceive returned true with value %d; expected false", value)
	}
}

func TestDrainChannel(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3

	values := DrainChannel(ch)

	if len(values) != 3 {
		t.Errorf("DrainChannel returned %d values; want 3", len(values))
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("values[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestPriorityReceive_HighReady(t *testing.T) {
	high := make(chan int, 1)
	low := make(chan int, 1)

	high <- 100
	low <- 1

	value, priority := PriorityReceive(high, low)

	if priority != "high" {
		t.Errorf("priority = %q; want \"high\"", priority)
	}
	if value != 100 {
		t.Errorf("value = %d; want 100", value)
	}
}

func TestPriorityReceive_OnlyLow(t *testing.T) {
	high := make(chan int, 1)
	low := make(chan int, 1)

	low <- 1
	// high is empty

	// Need a goroutine because we'll block waiting
	done := make(chan struct{})
	var value int
	var priority string

	go func() {
		value, priority = PriorityReceive(high, low)
		close(done)
	}()

	select {
	case <-done:
		if priority != "low" {
			t.Errorf("priority = %q; want \"low\"", priority)
		}
		if value != 1 {
			t.Errorf("value = %d; want 1", value)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("PriorityReceive blocked; should have received from low")
	}
}

func TestRelay(t *testing.T) {
	input := make(chan int, 5)
	output := make(chan int, 5)

	for i := 1; i <= 3; i++ {
		input <- i
	}
	close(input)

	Relay(input, output)
	close(output)

	var values []int
	for v := range output {
		values = append(values, v)
	}

	if len(values) != 3 {
		t.Errorf("Relay forwarded %d values; want 3", len(values))
	}
}

func TestMultiplex(t *testing.T) {
	input := make(chan int, 6)
	output0 := make(chan int, 3)
	output1 := make(chan int, 3)

	outputs := []chan<- int{output0, output1}

	// Route even to 0, odd to 1
	routeFn := func(v int) int { return v % 2 }

	for i := 0; i < 6; i++ {
		input <- i
	}
	close(input)

	Multiplex(input, outputs, routeFn)

	// Check outputs
	close(output0)
	close(output1)

	var evens, odds []int
	for v := range output0 {
		evens = append(evens, v)
	}
	for v := range output1 {
		odds = append(odds, v)
	}

	if len(evens) != 3 {
		t.Errorf("output0 got %d values; want 3 evens", len(evens))
	}
	if len(odds) != 3 {
		t.Errorf("output1 got %d values; want 3 odds", len(odds))
	}
}

func TestFairMerge(t *testing.T) {
	ch1 := make(chan int, 10)
	ch2 := make(chan int, 10)

	// Both channels have values
	for i := 0; i < 5; i++ {
		ch1 <- 1 // all 1s from ch1
		ch2 <- 2 // all 2s from ch2
	}
	close(ch1)
	close(ch2)

	output := FairMerge([]<-chan int{ch1, ch2})
	if output == nil {
		t.Fatal("FairMerge returned nil")
	}

	var values []int
	for v := range output {
		values = append(values, v)
	}

	if len(values) != 10 {
		t.Errorf("FairMerge produced %d values; want 10", len(values))
	}

	// With fair merging, we should see interleaved values
	// Count transitions between 1 and 2
	transitions := 0
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1] {
			transitions++
		}
	}

	// Fair merge should have many transitions (round-robin)
	// Unfair merge might have 1 transition (all 1s then all 2s)
	if transitions < 5 {
		t.Logf("Values: %v", values)
		t.Errorf("FairMerge had %d transitions; expected at least 5 for fair scheduling", transitions)
	}
}

// =============================================================================
// Conceptual tests
// =============================================================================

func TestSelectRandomness(t *testing.T) {
	t.Log("CONCEPT: When multiple cases are ready, select chooses randomly")
	t.Log("         This prevents starvation but means you can't assume order")

	ch1 := make(chan int, 100)
	ch2 := make(chan int, 100)

	for i := 0; i < 100; i++ {
		ch1 <- 1
		ch2 <- 2
	}

	var from1, from2 int
	for i := 0; i < 100; i++ {
		select {
		case <-ch1:
			from1++
		case <-ch2:
			from2++
		}
	}

	t.Logf("From ch1: %d, from ch2: %d (should be roughly equal)", from1, from2)

	// With random selection, we'd expect roughly 50/50 split
	// Allow 30-70 range for statistical variance
	if from1 < 30 || from1 > 70 {
		t.Error("Select doesn't appear to be random between ready cases")
	}
}

func TestNilChannelInSelect(t *testing.T) {
	t.Log("CONCEPT: A nil channel in select is NEVER ready")
	t.Log("         Use this to dynamically enable/disable cases")

	var ch1 chan int = nil
	ch2 := make(chan int, 1)
	ch2 <- 42

	// With ch1 nil, only ch2 case can fire
	select {
	case <-ch1:
		t.Error("nil channel should never be ready")
	case v := <-ch2:
		t.Logf("Received %d from ch2 (as expected, ch1 is nil)", v)
	}
}
