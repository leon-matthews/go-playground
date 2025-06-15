package todo

import (
	"os"
	"testing"
)

func TestAdd(t *testing.T) {
	l := List{}
	name := "This is something I should do"
	l.Add(name)
	if l[0].Task != name {
		t.Errorf("Expected %q, got %q", name, l[0].Task)
	}
}

func TestComplete(t *testing.T) {
	l := List{}
	name := "Do I have to do all of of these?"
	l.Add(name)
	if l[0].Task != name {
		t.Errorf("Expected %q, got %q", name, l[0].Task)
	}
	if l[0].Done {
		t.Errorf("New task should not be completed")
	}

	l.Complete(1)
	if !l[0].Done {
		t.Errorf("New task should be completed")
	}
}

func TestDelete(t *testing.T) {
	l := List{}
	tasks := []string{
		"New Task 1",
		"New Task 2",
		"New Task 3",
	}
	for _, t := range tasks {
		l.Add(t)
	}
	if l[0].Task != tasks[0] {
		t.Errorf("Expected %q, got %q", tasks[0], l[0].Task)
	}

	l.Delete(2)
	if len(l) != 2 {
		t.Errorf("Expected list length %d, got %d", 2, len(l))
	}
}

func TestLoadSave(t *testing.T) {
	l1 := List{}
	l2 := List{}

	name := "New Task"
	l1.Add(name)
	if l1[0].Task != name {
		t.Errorf("Expected %q, got %q", name, l1[0].Task)
	}

	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	defer os.Remove(tf.Name())

	if err := l1.Save(tf.Name()); err != nil {
		t.Fatalf("Error saving list to file: %s", err)
	}
	if err := l2.Load(tf.Name()); err != nil {
		t.Fatalf("Error getting list from file: %s", err)
	}
	if l1[0].Task != l2[0].Task {
		t.Errorf("Task %q should match %q task", l1[0].Task, l2[0].Task)
	}
}
