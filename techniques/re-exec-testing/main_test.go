package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// TestEchoHelper runs when re-exec'd with GO_WANT_HELPER_PROCESS=1
// Prints arguments and exits, emulating the "echo" binary.
func TestEchoHelper(t *testing.T) {
	if os.Getenv("BANANA") != "1" {
		return
	}
	fmt.Print(os.Args[len(os.Args)-1])
	os.Exit(0)
}

func TestRunEcho(t *testing.T) {
	// Create and re-exec the same binary (os.Args[0]) with additional args
	want := "Hello, world!"
	cmd := exec.Command(os.Args[0], "--test.run=TestEchoHelper", "--", want)
	t.Log(cmd)
	cmd.Env = append(os.Environ(), "BANANA=1")

	out, err := handleRunEcho(cmd)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
