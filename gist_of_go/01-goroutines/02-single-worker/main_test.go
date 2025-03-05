package main_test

import "testing"

func BenchmarkUnbufferedChannel(b *testing.B) {
	ch := make(chan int)
	go func() {
		for {
			<-ch
		}
	}()

	for i := 0; i < b.N; i++ {
		ch <- i
	}
}

func BenchmarkBufferedChannel(b *testing.B) {
	ch := make(chan int, 16)
	go func() {
		for {
			<-ch
		}
	}()

	for i := 0; i < b.N; i++ {
		ch <- i
	}
}
