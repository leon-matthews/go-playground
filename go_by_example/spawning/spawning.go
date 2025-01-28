package main

import (
	"fmt"
	"io"
	"os/exec"
)

func main() {
	dateCmd := exec.Command("date")
	fmt.Println(dateCmd)

	// Success
	dateOut, err := dateCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("date>", string(dateOut))

	// Error
	_, err = exec.Command("date", "-x").Output()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			// Command not found
			fmt.Println("failed executing:", err)
		case *exec.ExitError:
			// Command exited with non-zero exit code
			fmt.Println("command exited with exit code", e.ExitCode())
		default:
			panic(err)
		}
	}

	// Pipe data into stdin
	fmt.Println("> grep hello")
	grepCmd := exec.Command("grep", "hello")
	grepOut, _ := grepCmd.StdoutPipe()
	grepIn, _ := grepCmd.StdinPipe()

	grepCmd.Start()
	grepIn.Write([]byte("starting grep\nhello grep\ngoodbye grep"))
	grepIn.Close()
	b, _ := io.ReadAll(grepOut)
	fmt.Println(string(b))
	grepCmd.Wait()
}
