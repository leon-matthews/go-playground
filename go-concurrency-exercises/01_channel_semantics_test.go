package concurrency

import (
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 1.1: Channel Semantics
// Run with: go test -v -run Test01
// =============================================================================

func TestUnbufferedDemo(t *testing.T) {
	result := UnbufferedDemo()
	if result != 42 {
		t.Errorf("UnbufferedDemo() = %d; want 42", result)
	}
}

func TestBufferedDemo(t *testing.T) {
	result := BufferedDemo()
	if result != 42 {
		t.Errorf("BufferedDemo() = %d; want 42", result)
	}
}

func TestBufferFullDemo(t *testing.T) {
	start := time.Now()
	result := BufferFullDemo()
	elapsed := time.Since(start)

	if !result {
		t.Error("BufferFullDemo() returned false; expected true")
	}

	// Should take at least 50ms because third send blocks until receiver is ready
	if elapsed < 40*time.Millisecond {
		t.Errorf("BufferFullDemo() completed too fast (%v); expected blocking behavior", elapsed)
	}
}

func TestClosedChannelReceive(t *testing.T) {
	result := ClosedChannelReceive()

	if len(result) < 2 {
		t.Errorf("ClosedChannelReceive() returned %d values; want at least 2", len(result))
		return
	}

	if result[0] != "first" {
		t.Errorf("First value = %q; want \"first\"", result[0])
	}
	if result[1] != "second" {
		t.Errorf("Second value = %q; want \"second\"", result[1])
	}
}

func TestRangeOverChannel(t *testing.T) {
	result := RangeOverChannel()
	expected := 1 + 2 + 3 // = 6

	if result != expected {
		t.Errorf("RangeOverChannel() = %d; want %d", result, expected)
	}
}

func TestNilChannelBehavior(t *testing.T) {
	start := time.Now()
	result := NilChannelBehavior()
	elapsed := time.Since(start)

	if result != "timeout" {
		t.Errorf("NilChannelBehavior() = %q; want \"timeout\"", result)
	}

	// Should take about 100ms (the timeout)
	if elapsed < 90*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("NilChannelBehavior() took %v; expected ~100ms", elapsed)
	}
}

func TestChannelDirectionDemo(t *testing.T) {
	// Sum of squares of 0, 1, 2, 3, 4 = 0 + 1 + 4 + 9 + 16 = 30
	result := ChannelDirectionDemo(5)
	expected := 30

	if result != expected {
		t.Errorf("ChannelDirectionDemo(5) = %d; want %d", result, expected)
	}
}

func TestSendWithTimeout_Success(t *testing.T) {
	ch := make(chan int, 1) // buffered, so send won't block

	result := SendWithTimeout(ch, 42, 100*time.Millisecond)

	if !result {
		t.Error("SendWithTimeout() returned false; expected true (send should succeed)")
	}

	// Verify the value was sent
	select {
	case v := <-ch:
		if v != 42 {
			t.Errorf("Channel received %d; want 42", v)
		}
	default:
		t.Error("Channel is empty; expected value 42")
	}
}

func TestSendWithTimeout_Timeout(t *testing.T) {
	ch := make(chan int) // unbuffered, no receiver = will block

	start := time.Now()
	result := SendWithTimeout(ch, 42, 50*time.Millisecond)
	elapsed := time.Since(start)

	if result {
		t.Error("SendWithTimeout() returned true; expected false (should timeout)")
	}

	// Should take about 50ms (the timeout)
	if elapsed < 40*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("SendWithTimeout() took %v; expected ~50ms", elapsed)
	}
}

// =============================================================================
// BONUS: Conceptual understanding tests
// =============================================================================

// TestUnbufferedBlocking verifies you understand WHERE blocking happens
func TestUnbufferedBlocking(t *testing.T) {
	t.Log("CONCEPT: In UnbufferedDemo, the goroutine blocks on send until main receives")
	t.Log("         The 100ms sleep in main delays the receive, so goroutine waits")
}

// TestBufferedNonBlocking verifies you understand buffered behavior
func TestBufferedNonBlocking(t *testing.T) {
	t.Log("CONCEPT: In BufferedDemo, send completes immediately because buffer has space")
	t.Log("         This is why it works without a goroutine!")
}

// TestClosedChannelConcept explains the comma-ok idiom
func TestClosedChannelConcept(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	close(ch)

	// First receive: gets the buffered value
	v1, ok1 := <-ch
	t.Logf("First receive: value=%d, ok=%v (buffered value)", v1, ok1)

	// Second receive: channel is closed AND empty
	v2, ok2 := <-ch
	t.Logf("Second receive: value=%d, ok=%v (zero value, closed)", v2, ok2)

	// This demonstrates: closed channel returns zero value + false
	if ok1 != true || ok2 != false {
		t.Error("Unexpected comma-ok behavior")
	}
}
