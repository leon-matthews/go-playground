package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	binName  = "todo"
	fileName = "todo.json"
)

func TestMain(m *testing.M) {
	// Build binary to run tests against
	log.Println("Building tool...")
	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		log.Fatalf("Cannot build tool %s: %s", binName, err)
	}

	log.Println("Running tests...")
	result := m.Run()

	log.Println("Cleaning up...")
	os.Remove(binName)
	os.Remove(fileName)
	os.Exit(result)
}

func TestCLI(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmdPath := filepath.Join(dir, binName)

	heading := "test task number 1"
	t.Run("AddNewTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "--add", heading)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Println(string(out))
			t.Fatal(err)
		}
	})

	heading2 := "test task number 2"
	t.Run("AddNewTaskStdin", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "--add")
		cmdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(cmdIn, heading2)
		cmdIn.Close()
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Println(string(out))
			t.Fatal(err)
		}
	})

	t.Run("ListTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "--list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(out))
			t.Fatal(err)
		}

		expected := fmt.Sprintf("  1: %s\n  2: %s\n", heading, heading2)
		if expected != string(out) {
			t.Errorf("Expected %q, got %q", expected, string(out))
		}
	})
}
