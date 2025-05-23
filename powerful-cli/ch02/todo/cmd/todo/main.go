package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"powerful-cli/ch02/todo"
)

// Default file name
var todoPath = "todo.json"


func main() {
	// Use default file name?
	if os.Getenv("TODO_FILENAME") != "" {
		todoPath = os.Getenv("TODO_FILENAME")
	}

	// Change default usage
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(
			out,
			"%s tool. Chapter 2 of 'Powerful Command-line Applications in Go\n",
			os.Args[0],
		)
		fmt.Fprintf(out, "Copyright 2025\n")
		fmt.Fprintf(out, "Usage:\n")
		flag.PrintDefaults()
	}

	// Parse command-line flags
	taskFlag := flag.String("task", "", "Add task to TODO list")
	listFlag := flag.Bool("list", false, "List all tasks")
	completeFlag := flag.Int("complete", 0, "Item to be completed")
	flag.Parse()

	// Attempt to load list
	l := &todo.List{}
	if err := l.Load(todoPath); err != nil {
		log.Fatalf("Loading list: %s", err)
		os.Exit(1)
	}

	// Check args for action
	switch {

	case *listFlag:
		// List current TODO items
		fmt.Print(l)

	case *completeFlag > 0:
		// Mark task as done
		if err := l.Complete(*completeFlag); err != nil {
			log.Fatal(err)
		}
		if err := l.Save(todoPath); err != nil {
			log.Fatal(err)
		}

	case *taskFlag != "":
		// Add task to TODO list
		l.Add(*taskFlag)
		if err := l.Save(todoPath); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("Invalid option")
	}
}
