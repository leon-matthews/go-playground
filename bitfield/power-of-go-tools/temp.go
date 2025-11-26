package main

import (
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("/bin/ls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
