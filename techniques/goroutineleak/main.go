package main

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {
	fmt.Println("goroutines at start:", runtime.NumGoroutine())

	// Fetch a bunch of prices from our too-slow fake endpoint
	for range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		if _, err := fetchPrice(ctx); err != nil {
			fmt.Println("err:", err)
		}
		cancel()
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println("goroutines at end:", runtime.NumGoroutine())

	// Run goroutineleak profile
	var buf bytes.Buffer
	if err := pprof.Lookup("goroutineleak").WriteTo(&buf, 1); err != nil {
		fmt.Println("profile error:", err)
		return
	}
	fmt.Println(buf.String())
}

// fetchPrice will leave a goroutine running if ctx cancels before result comes back.
func fetchPrice(ctx context.Context) (int, error) {
	result := make(chan int)

	// Write value to result after some time
	go func() {
		time.Sleep(50 * time.Millisecond)
		result <- 42
	}()

	select {
	case price := <-result:
		return price, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}
