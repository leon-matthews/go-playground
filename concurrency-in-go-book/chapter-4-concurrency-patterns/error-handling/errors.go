package main

import (
	"fmt"
	"net/http"
)

func main() {
	urls := []string{"https://google.com/", "https://lost.co.nz/", "a", "b", "c"}

	done := make(chan any)
	defer close(done)
	var numErrors int
	for r := range checkStatus(done, urls...) {
		if r.Error != nil {
			fmt.Printf("Error: %v\n", r.Error)
			numErrors++
			if numErrors >= 3 {
				fmt.Println("Too many errors, breaking")
				break
			}
			continue
		}
		fmt.Printf("%v: %v\n", r.URL, r.Response.Status)
	}
}

// Bundle error with result so that upper layer can handle it intelligently
type Result struct {
	URL      string
	Response *http.Response
	Error    error
}

// checkStatus attempts to fetch, via HTTP, all of the given URLs
func checkStatus(done <-chan any, urls ...string) <-chan Result {
	results := make(chan Result)
	go func() {
		defer close(results)
		for _, url := range urls {
			// Fetch URL and build result
			response, err := http.Get(url)
			r := Result{url, response, err}

			// Send response back - or handle cancellation signal
			select {
			case results <- r:
			case <-done:
				return
			}
		}
	}()
	return results
}
