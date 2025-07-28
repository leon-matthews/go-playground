package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

func main() {
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
	for i := range 21 {
		count := int(math.Pow(2, float64(i)))
		fmt.Fprintf(
			tw,
			"2**%d\t%v\t%v\n",
			i,
			test(count, &m, m.RLocker()),
			test(count, &m, &m),
		)
	}
}

func test(count int, mutex, rwMutex sync.Locker) time.Duration {
	var wg sync.WaitGroup
	wg.Add(count + 1)
	start := time.Now()
	go producer(&wg, mutex)
	for i := count; i > 0; i-- {
		go observer(&wg, rwMutex)
	}
	wg.Wait()
	return time.Since(start)
}

func observer(wg *sync.WaitGroup, l sync.Locker) {
	defer wg.Done()
	l.Lock()
	defer l.Unlock()
}

func producer(wg *sync.WaitGroup, l sync.Locker) {
	defer wg.Done()
	for range 5 {
		l.Lock()
		l.Unlock()
		time.Sleep(1)
	}
}
