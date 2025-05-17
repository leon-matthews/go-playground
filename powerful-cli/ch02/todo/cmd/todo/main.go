package main

import (
	"fmt"
	"os"
	"strings"

	"powerful-cli/ch02/todo"
)

const todoPath = "todo.json"

func main() {
	// Attempt to load list
	l := &todo.List{}
	if err := l.Load(todoPath); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}

	// Check args for action
	switch {

	// List current todo items
	case len(os.Args) == 1:
		for _, item := range *l {
			fmt.Printf("\u2713 %v\n", item.Task)
		}

	// Treat multiple arguments as a single string
	default:
		task := strings.Join(os.Args[1:], " ")
		l.Add(task)
		if err := l.Save(todoPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
