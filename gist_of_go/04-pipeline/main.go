package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
)

/*
A 5-stage pipeline:

1. rangeGen() generates numbers within a given range (reader).
2. takeLucky() selects "lucky" numbers (processor).
3. merge() combines independent channels (processor).
4. sum() sums the numbers (processor).
5. printTotal() prints the result (writer).

┌─────────────┐
│   rangeGen  │
└─────────────┘
       │
  workerIn ─┬────────┬──────────────┬──────────────┐
┌─────────────┐┌─────────────┐┌─────────────┐┌─────────────┐
│  takeLucky  ││  takeLucky  ││  takeLucky  ││  takeLucky  │
└─────────────┘└─────────────┘└─────────────┘└─────────────┘
       │               │              │              │
 luckyChans[0]   luckyChans[1]  luckyChans[2]  luckyChans[3]
       │               │              │              │
┌──────────────────────────────────────────────────────────┐
│                           merge                          │
└──────────────────────────────────────────────────────────┘
       │
   mergedChan
       │
┌─────────────┐
│     sum     │
└─────────────┘
       │
   totalChan
       │
┌─────────────┐
│ printTotal  │
└─────────────┘
*/

// Total represents the count
// and the sum of the lucky numbers.
type Total struct {
	count  int
	amount int
}

func main() {
	numWorkers := getNumWorkers()
	readerChan := rangeGen(1, 1_000_000)
	luckyChans := make([]<-chan int, numWorkers)
	for i := range numWorkers {
		luckyChans[i] = takeLucky(readerChan)
	}
	mergedChan := merge(luckyChans)
	totalChan := sum(mergedChan)
	printTotal(totalChan)
}

// rangeGen generates numbers within a given range.
func rangeGen(start, stop int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := start; i < stop; i++ {
			out <- i
		}
	}()
	return out
}

// isPrimes performs primality check using naive trial-division
func isPrime(num int) bool {
	for i := 2; i < num; i++ {
		if num%i == 0 {
			return false
		}
	}
	return true
}

// takeLucky selects lucky numbers.
func takeLucky(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for num := range in {
			if isPrime(num) {
				out <- num
			}
		}
	}()
	return out
}

func merge(inputs []<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, in := range inputs {
		go func(in <-chan int) {
			for v := range in {
				out <- v
			}
			wg.Done()
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// sum sums the numbers.
func sum(in <-chan int) <-chan Total {
	out := make(chan Total)
	go func() {
		defer close(out)
		total := Total{}
		for num := range in {
			total.amount += num
			total.count++
		}
		out <- total
	}()
	return out
}

// printTotal prints the result.
func printTotal(in <-chan Total) {
	total := <-in
	fmt.Printf("Total of %d lucky numbers = %d\n", total.count, total.amount)
}

// getNumWorkers decides how many CPU cores to use, respecting GOMAXPROCS
func getNumWorkers() int {
	numCPU := runtime.NumCPU()
	numProcs := runtime.GOMAXPROCS(0)
	if numProcs != numCPU {
		log.Printf("Found %d CPU cores, but GOMAXPROCS=%d", numCPU, numProcs)
		numCPU = numProcs
	} else {
		log.Printf("Found %d CPU cores", numCPU)
	}
	return numCPU
}
