package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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
	addFlag := flag.Bool("add", false, "Add task to TODO list")
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

	case *addFlag:
		// Add task to TODO list
		t, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			log.Fatal(err)
		}
		l.Add(t)

		if err := l.Save(todoPath); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("Invalid option")
	}
}

// getTask fetches a task title from flag or stdin
func getTask(r io.Reader, args ...string) (string, error) {
	// Prefer args if present
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	// Try first line from stdio instead
	s := bufio.NewScanner(r)
	s.Scan()
	if err := s.Err(); err != nil {
		return "", err
	}
	if len(s.Text()) == 0 {
		return "", fmt.Errorf("Task cannot be blank")
	}
	return s.Text(), nil
}
