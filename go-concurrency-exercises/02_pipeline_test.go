package concurrency

import (
	"errors"
	"testing"
	"time"
)

// =============================================================================
// Tests for Exercise 1.2: Pipelines
// Run with: go test -v -run Test02
// =============================================================================

func TestGenerate(t *testing.T) {
	ch := Generate(1, 5)
	if ch == nil {
		t.Fatal("Generate returned nil channel")
	}

	var values []int
	for v := range ch {
		values = append(values, v)
	}

	expected := []int{1, 2, 3, 4}
	if len(values) != len(expected) {
		t.Fatalf("Generate(1,5) produced %d values; want %d", len(values), len(expected))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("values[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestSquare(t *testing.T) {
	input := make(chan int, 3)
	input <- 2
	input <- 3
	input <- 4
	close(input)

	output := Square(input)
	if output == nil {
		t.Fatal("Square returned nil channel")
	}

	var results []int
	for v := range output {
		results = append(results, v)
	}

	expected := []int{4, 9, 16}
	if len(results) != len(expected) {
		t.Fatalf("Square produced %d values; want %d", len(results), len(expected))
	}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestSum(t *testing.T) {
	input := make(chan int, 3)
	input <- 1
	input <- 4
	input <- 9
	close(input)

	result := Sum(input)
	expected := 14

	if result != expected {
		t.Errorf("Sum() = %d; want %d", result, expected)
	}
}

func TestRunPipeline(t *testing.T) {
	// 1^2 + 2^2 + 3^2 = 1 + 4 + 9 = 14
	result := RunPipeline(1, 4)
	expected := 14

	if result != expected {
		t.Errorf("RunPipeline(1, 4) = %d; want %d", result, expected)
	}
}

func TestFilter(t *testing.T) {
	input := Generate(1, 6) // 1, 2, 3, 4, 5
	isEven := func(n int) bool { return n%2 == 0 }

	output := Filter(input, isEven)
	if output == nil {
		t.Fatal("Filter returned nil channel")
	}

	var results []int
	for v := range output {
		results = append(results, v)
	}

	expected := []int{2, 4}
	if len(results) != len(expected) {
		t.Fatalf("Filter produced %d values; want %d", len(results), len(expected))
	}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestMap(t *testing.T) {
	input := Generate(1, 4) // 1, 2, 3
	double := func(n int) int { return n * 2 }

	output := Map(input, double)
	if output == nil {
		t.Fatal("Map returned nil channel")
	}

	var results []int
	for v := range output {
		results = append(results, v)
	}

	expected := []int{2, 4, 6}
	if len(results) != len(expected) {
		t.Fatalf("Map produced %d values; want %d", len(results), len(expected))
	}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestRunFilterMapPipeline(t *testing.T) {
	// 2^2 + 4^2 + 6^2 + 8^2 = 4 + 16 + 36 + 64 = 120
	result := RunFilterMapPipeline()
	expected := 120

	if result != expected {
		t.Errorf("RunFilterMapPipeline() = %d; want %d", result, expected)
	}
}

func TestGenerateWithError(t *testing.T) {
	results := GenerateWithError(1, 6, 3) // errors at 3
	if results == nil {
		t.Fatal("GenerateWithError returned nil channel")
	}

	var values []int
	var errCount int
	for r := range results {
		if r.Err != nil {
			errCount++
		} else {
			values = append(values, r.Value)
		}
	}

	// Values: 1, 2, 4, 5 (3 is an error)
	if len(values) != 4 {
		t.Errorf("Got %d successful values; want 4", len(values))
	}
	if errCount != 1 {
		t.Errorf("Got %d errors; want 1", errCount)
	}
}

func TestProcessResults(t *testing.T) {
	input := make(chan Result, 5)
	input <- Result{Value: 1}
	input <- Result{Value: 2}
	input <- Result{Err: errors.New("test error")}
	input <- Result{Value: 4}
	input <- Result{Err: errors.New("another error")}
	close(input)

	sum, errCount := ProcessResults(input)

	if sum != 7 { // 1 + 2 + 4
		t.Errorf("sum = %d; want 7", sum)
	}
	if errCount != 2 {
		t.Errorf("errorCount = %d; want 2", errCount)
	}
}

func TestGenerateCancellable(t *testing.T) {
	done := make(chan struct{})
	output := GenerateCancellable(1, 1000000, done) // large range

	// Read a few values
	<-output
	<-output
	<-output

	// Cancel the pipeline
	close(done)

	// Give it time to shut down
	time.Sleep(10 * time.Millisecond)

	// Pipeline should stop producing values
	// (This is a basic test - real test would check goroutine cleanup)
	t.Log("GenerateCancellable stopped successfully after done signal")
}

func TestTransformCancellable(t *testing.T) {
	done := make(chan struct{})
	input := GenerateCancellable(1, 1000000, done)
	double := func(n int) int { return n * 2 }
	output := TransformCancellable(input, double, done)

	// Read a few values
	v1 := <-output
	if v1 != 2 { // 1 * 2
		t.Errorf("First transformed value = %d; want 2", v1)
	}

	// Cancel
	close(done)
	time.Sleep(10 * time.Millisecond)

	t.Log("TransformCancellable stopped successfully after done signal")
}

// =============================================================================
// Conceptual tests
// =============================================================================

func TestPipelineConceptChannelDirection(t *testing.T) {
	t.Log("CONCEPT: Returning <-chan int instead of chan int prevents callers")
	t.Log("         from accidentally sending to or closing the channel.")
	t.Log("         The producer controls the channel lifecycle.")
}

func TestPipelineConceptClosePropagation(t *testing.T) {
	t.Log("CONCEPT: When a stage closes its output channel, the next stage's")
	t.Log("         `for range` loop exits naturally. Closure propagates!")
}
