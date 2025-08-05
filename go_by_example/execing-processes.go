// Replace current process with another
package main

import (
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// Find path to ls
	path, err := exec.LookPath("ls")
	if err != nil {
		panic(err)
	}

	// First arg is command's name
	args := []string{"ls", "-a", "-s", "-h"}
	env := os.Environ()

	err = syscall.Exec(path, args, env)
	if err != nil {
		panic(err)
	}
}
