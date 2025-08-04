package main

import (
	"flag"
	"fmt"
	"os"
)

const commandMissing = "expected 'first' or 'second' subcommands"

// Multi-command CLIs using stdlib are possible, but a little awkward.
func main() {
	// First command
	first := flag.NewFlagSet("first", flag.ExitOnError)
	firstEnable := first.Bool("enable", true, "enable feature")
	firstName := first.String("name", "default", "name to use")

	// Second command
	second := flag.NewFlagSet("second", flag.ExitOnError)
	secondLevel := second.Int("level", 0, "level to start on")

	// Ensure command arg
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, commandMissing)
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "first":
		first.Parse(os.Args[2:])
		fmt.Printf("Running command %q\n", command)
		fmt.Println("enabled:", *firstEnable)
		fmt.Println("name:", *firstName)
	case "second":
		second.Parse(os.Args[2:])
		fmt.Printf("Running command %q\n", command)
		fmt.Println("level:", *secondLevel)
	default:
		fmt.Fprintln(os.Stderr, commandMissing)
		os.Exit(1)
	}
}
