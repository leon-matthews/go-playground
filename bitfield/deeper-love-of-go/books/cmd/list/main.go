// Package main implements a light-weight CLI to list all books available
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"local.dev/books"
)

func main() {
	// Fetch JSON
	response, err := http.Get("http://localhost:8000/")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Read JSON
	if response.StatusCode != http.StatusOK {
		log.Fatal(response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal JSON
	books := &[]books.Book{}
	err = json.Unmarshal(body, &books)
	if err != nil {
		log.Fatalf("unmarshalling JSON: %v", err)
	}

	// Print list of books
	for _, book := range *books {
		fmt.Println(book)
	}
}
