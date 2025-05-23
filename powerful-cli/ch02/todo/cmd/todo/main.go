package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"powerful-cli/ch02/todo"
)

const todoPath = "todo.json"

func main() {
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
		for _, item := range *l {
			if item.Done {
				fmt.Printf("\u2713 %v\n", item.Task)
			} else {
				fmt.Printf("  %v\n", item.Task)
			}
		}

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
