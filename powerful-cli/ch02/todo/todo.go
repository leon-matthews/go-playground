package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// item represents a single task to do
type item struct {
	Task      string
	Done      bool
	Created   *time.Time
	Completed *time.Time
}

// List represents a list of work items to do
type List []item

// Add creates a new todo item and add it to the list
func (l *List) Add(task string) {
	now := time.Now()
	t := item{
		Task:      task,
		Done:      false,
		Created:   &now,
		Completed: nil,
	}
	*l = append(*l, t)
}

// Complete marks a todo item as completed
func (l *List) Complete(i int) error {
	ls := *l
	if i <= 0 || i > len(ls) {
		return fmt.Errorf("Item %d does not exist", i)
	}

	// Adjust index for 0-based index
	now := time.Now()
	ls[i-1].Done = true
	ls[i-1].Completed = &now
	return nil
}

// Load todo list from JSON file
func (l *List) Load(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(file) == 0 {
		return nil
	}
	return json.Unmarshal(file, l)
}

// Delete completely removes a todo item
func (l *List) Delete(i int) error {
	ls := *l
	if i <= 0 || i > len(ls) {
		return fmt.Errorf("Item %d does not exist", i)
	}

	*l = append(ls[:i-1], ls[i:]...)
	return nil
}

// Save encodes the list as JSON and saves into the given filename
func (l *List) Save(filename string) error {
	js, err := json.Marshal(l)
	if err != nil {
		return nil
	}
	return os.WriteFile(filename, js, 0664)
}
