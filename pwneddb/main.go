package main

import (
	"fmt"
	"io"
	"net/http"
	"slices"
)

const etag = `"0x8DD55EF4D1039F4"`
const url = "https://api.pwnedpasswords.com/range/cafe5"

func main() {
	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("If-None-Match", etag)

	r, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	fmt.Println("Response status:", r.Status)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Response bytes:", len(body))
	fmt.Println()
	fmt.Println("REQUEST")
	printHeaders(r.Request.Header)
	fmt.Println()
	fmt.Println("RESPONSE")
	printHeaders(r.Header)

	//fmt.Println(string(body))
}

func printHeaders(header http.Header) {
	// Sort keys alphanumerically
	names := make([]string, 0, len(header))
	for key := range header {
		names = append(names, key)
	}
	slices.Sort(names)

	// Print header
	for _, name := range names {
		fmt.Printf("%s: %s\n", name, header.Get(name))
	}
}
