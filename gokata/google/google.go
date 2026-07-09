// Package google is a toy search engine using the Go concurrency patterns from
// the boring kata. We start with slow sequential (or serial) code and evolve it
// until it's fast and robust thanks to concurrency and replication.
// See https://youtu.be/f6kdp27TYZs?t=1702 for more.
package google

import (
	"fmt"
	"math/rand"
	"time"
)

// A Result is just a string
type Result string

// Search is a function that takes a query and returns its result
type Search func(query string) Result

// NewSearch creates a new Search function for a given type of search
// For example: imageSearch := NewSearch("image")
func NewSearch(kind string) Search {
	search := func(query string) Result {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
		return Result(fmt.Sprintf("%s search result for %q\n", kind, query))
	}
	return search
}
