package main

import (
    "testing"
    "sync"
)

func TestCounter(t *testing.T) {
    t.Run("increments three times", func(t *testing.T) {
        counter := NewCounter()

        counter.Increment()
        counter.Increment()
        counter.Increment()

        assertCounter(t, counter, 3)
    })

    t.Run("runs concurrently", func(t *testing.T) {
        counter := NewCounter()
        var want int64 = 1000

        var wg sync.WaitGroup
        for range want {
            wg.Go(func(){ counter.Increment() })
        }
        wg.Wait()

        assertCounter(t, counter, want)
    })
}

func BenchmarkCounter(b *testing.B) {
    c := NewCounter()
    for b.Loop() {
        c.Increment()
        c.Value()
    }
}

func assertCounter(t testing.TB, got *Counter, want int64) {
    t.Helper()
    if got.Value() != want {
        t.Errorf("got %d, want %d", got.Value(), want)
    }
}
