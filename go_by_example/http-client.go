package main

import (
	"bufio"
	"fmt"
	"net/http"
)

func main() {
	// Get is a wrapper around DefaultClient.Get; create new client to customise
	resp, err := http.Get("https://gobyexample.com") // Returns *http.Response
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Request.URL, resp.Status)

	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 10; i++ {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
