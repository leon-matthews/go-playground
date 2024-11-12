package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

	shortest, longest := ShortAndTall(lines)
	fmt.Println("shortest", shortest, "longest", longest)

	counts := CountLengths(lines)
	binSize := 5
	printHistogram(counts, binSize, longest)
}

// usage add positional argument details
func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s PATH\n", os.Args[0])
	flag.PrintDefaults()
}
