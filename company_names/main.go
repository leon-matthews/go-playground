package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	path := flag.Arg(0)
	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	lines, err := ReadLines(path)
	if err != nil {
		log.Fatal(err)
	}

	counts := make(map[int]int)
	for _, line := range lines {
		counts[len(line)]++
	}

	numBins := 14
	binSize := 5
	for bin := 0; bin < numBins; bin++ {
		count := 0
		for i := 0; i < binSize; i++ {
			index := binSize*bin + i
			c, ok := counts[index]
			if ok {
				count += c
			}
		}
		fmt.Printf("%2d: %v\n", bin*binSize, strings.Repeat("#", count))
	}
}

// usage add positional argument details
func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s PATH\n", os.Args[0])
	flag.PrintDefaults()
}
