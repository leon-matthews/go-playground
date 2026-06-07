// Fetch ranges over the CLI arguments (that should be URLs) and fetches each of
// them. It reports the size of the resource at the URL and the time it took to
// download it. It does so concurrently.
//
// Adapted from https://github.com/adonovan/gopl.io/tree/master/ch1/fetchall.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	start := time.Now()
	urls := os.Args[1:]

	// Results from fetch() will be written here
	ch := make(chan string)

	// Start one goroutine for each URL
	for _, u := range urls {
		go fetch(u, ch)
	}

	// Read results back as they come in
	for range urls {
		fmt.Println(<-ch)
	}

	fmt.Printf("%.3fs elapsed\n", time.Since(start).Seconds())
}

func fetch(url string, ch chan<- string) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	// Count bytes in body
	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		ch <- fmt.Sprintf("while reading %s: %v", url, err)
		return
	}

	// Report: time size url
	ch <- fmt.Sprintf("%.3fs %6dkB %s", time.Since(start).Seconds(), n/1_000, url)
}
