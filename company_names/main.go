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
	fmt.Println(names)
	/*
		shortest, longest := ShortAndTall(lines)
		fmt.Print("Unicode strings, lengths in bytes. ")
		fmt.Println("shortest", shortest, "longest", longest)

		counts := CountLengths(lines)
		binSize := 5
		printHistogram(counts, binSize, longest)

		for _, line := range lines {
			line, _ = ToAscii(line)
			var b = smaz.Compress([]byte(line))
			fmt.Printf("%d -> %d\n", len(line), len(b))
		}
	*/
}

// usage add positional argument details
func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s PATH\n", os.Args[0])
	flag.PrintDefaults()
}
