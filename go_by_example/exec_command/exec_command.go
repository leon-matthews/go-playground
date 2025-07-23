package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os/exec"
	"strings"
)

// Run an external command and communicate with it using standard streams
func main() {
	// Bare-bones
	timestamp, err := UnixTimestamp()
	log.Print("Unix timestamp: ", timestamp)

	// No such command
	_, err = exec.Command("no-such-command").Output()
	if err != nil {
		slogExecError(err)
	}

	// Non-zero exit code (invalid option)
	_, err = exec.Command("date", "-x").Output()
	if err != nil {
		slogExecError(err)
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

// UnixTimestamp runs the external date command to get the current timestamp
// [cmd.Output()] waits for the command to exit then returns its stdout.
// Some of its stderr output may be found in the returned error, if the command
// finished with a non-zero exit code.
func UnixTimestamp() (string, error) {
	cmd := exec.Command("date", "+%s")
	out, err := cmd.Output()
	if err != nil {
		if err != nil {
			slog.Error(err.Error())
		}
		return "", err
	}
	return string(out), nil
}

// slogExecError writes an error record, specialising the message for [exec] errors
func slogExecError(err error) {
	if err == nil {
		return
	}
	e, ok := err.(*exec.ExitError)
	if ok {
		slog.Error(fmt.Sprintf("%s: %q", e, firstLine(string(e.Stderr))))
	} else {
		slog.Error(err.Error())
	}
}

// firstLine returns just the... first line from the given string
func firstLine(s string) string {
	i := strings.Index(s, "\n")
	if i == -1 {
		return s
	}
	return s[:i]
}
