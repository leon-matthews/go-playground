// Download and print the content found at the given URL
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	urls := os.Args[1:]
	for _, url := range urls {
		response, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
			os.Exit(1)
		}

		bytes, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", url, err)
		}
		fmt.Printf("%s", bytes)
	}
}
