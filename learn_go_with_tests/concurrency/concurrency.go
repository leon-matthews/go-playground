package main

import (
	"fmt"
)

func main() {
	fmt.Println("concurrency")
}

type WebsiteChecker func(string) bool

type result struct {
	string
	bool
}

func CheckWebsites(wc WebsiteChecker, urls []string) map[string]bool {
	results := make(map[string]bool)
	ch := make(chan result)

	// Start one goroutine for each URL
	// Each one sending a `result` back through `ch`
	for _, url := range urls {
		go func(u string) {
			ch <- result{u, wc(u)}
		}(url)
	}

	// Receive results from `ch`
	for range len(urls) {
		r := <-ch
		results[r.string] = r.bool
	}

	return results
}
