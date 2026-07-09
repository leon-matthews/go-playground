// V2.0 runs the three kinds of search concurrently using the fan-in pattern.
// Concurrency makes the program faster.
package main

import (
	"fmt"

	"github.com/gokatas/google"
)

func main() {
	results := googleIt("golang")
	fmt.Println(results)
}

func googleIt(query string) (results []google.Result) {
	c := make(chan google.Result)

	go func() { c <- web(query) }()
	go func() { c <- image(query) }()
	go func() { c <- video(query) }()

	for i := 0; i < 3; i++ {
		result := <-c
		results = append(results, result)
	}

	return
}

var (
	web   = google.NewSearch("web")
	image = google.NewSearch("image")
	video = google.NewSearch("video")
)
