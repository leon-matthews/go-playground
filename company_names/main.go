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

	names, err := ReadNames(path)
	if err != nil {
		log.Fatal(err)
	}

	shortest, longest := ShortestAndLongest(names)
	fmt.Print("Unicode strings, lengths in bytes. ")
	fmt.Println("shortest", shortest, "longest", longest)

	binSize := 5
	counts := CountLengths(names)
	PrintHistogram(counts, binSize)
}

// usage add positional argument details
func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s PATH\n", os.Args[0])
	flag.PrintDefaults()
}
