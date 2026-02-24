package main

import (
	"fmt"
	"log"
	"os/exec"
)

// RunEcho builds a *Cmd, but doesn't run it
func RunEcho(msg string) (string, error) {
	cmd := exec.Command("echo", msg)
	return handleRunEcho(cmd)
}


// This function is tested by the re-exec technique
func handleRunEcho(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	return string(out), err
}

func main() {
	output, err := RunEcho("Hello, world!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}
