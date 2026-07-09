// V2.1 times out googling after 80ms using a variation of the timeout pattern.
// Consequently it sometimes returns only partial results. Thus it is fast but
// not very robust.
package main

import (
	"fmt"
	"time"

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

	timeout := time.After(time.Millisecond * 80)
	for i := 0; i < 3; i++ {
		select {
		case result := <-c:
			results = append(results, result)
		case <-timeout:
			fmt.Println("timeout")
			return
		}
	}

	return
}

var (
	web   = google.NewSearch("web")
	image = google.NewSearch("image")
	video = google.NewSearch("video")
)
