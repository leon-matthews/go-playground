// V1.0 invokes (fake) web, image and video searches serially.
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
	results = append(results, web(query))
	results = append(results, image(query))
	results = append(results, video(query))
	return
}

var (
	web   = google.NewSearch("web")
	image = google.NewSearch("image")
	video = google.NewSearch("video")
)
