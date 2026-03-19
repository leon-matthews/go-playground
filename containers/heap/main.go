// Example usage of heap package
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"heap/heap"
)

var inFile = flag.String("input", "", "text file to read from")

func main() {
	// Parse input
	flag.Parse()
	if *inFile == "" {
		log.Fatal("input file required")
	}

	// Iterate over words
	in, err := os.Open(*inFile)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
	for w := range words(in) {
		fmt.Print(w, " ")
	}
	h := heap.NewHeap[int]()
	h.Push(42)
	fmt.Println(h.Len())
	fmt.Println(h)
}

type message struct {
	id       int
	contents string
}

// words sends the words from the given path to the returned channel
func words(in io.Reader) <-chan string {
	wordStream := make(chan string)
	go func() {
		defer close(wordStream)
		scanner := bufio.NewScanner(in)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			wordStream <- scanner.Text()
		}
	}()
	return wordStream
}
