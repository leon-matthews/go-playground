// V3.0 introduces replication. It means we have multiple search services
// (replicas) for each kind and we take the first result returned by the fastest
// replica. This way we dramatically lower the likelihood of discarding results.
// This is a fast and robust program.
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

	go func() { c <- firstResult(query, web1, web2) }()
	go func() { c <- firstResult(query, image1, image2) }()
	go func() { c <- firstResult(query, video1, video2) }()

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
	web1   = google.NewSearch("web")
	web2   = google.NewSearch("web")
	image1 = google.NewSearch("image")
	image2 = google.NewSearch("image")
	video1 = google.NewSearch("video")
	video2 = google.NewSearch("video")
)

func firstResult(query string, replicas ...google.Search) google.Result {
	c := make(chan google.Result)
	for i := range replicas {
		go func(i int) { c <- replicas[i](query) }(i)
	}
	result := <-c
	return result
}
