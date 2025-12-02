package shell

import (
	"bufio"
	"fmt"
	"io"
)

type Session struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	DryRun         bool
}

func NewSession(in io.Reader, out, err io.Writer) *Session {
	s := Session{
		Stdin:  in,
		Stdout: out,
		Stderr: err,
	}
	return &s
}

func (s *Session) Run() {
	fmt.Fprintf(s.Stdout, "> ")
	input := bufio.NewScanner(s.Stdin)
	for input.Scan() {
		line := input.Text()
		cmd, err := CmdFromString(line)

		// Ignore empty lines
		if err != nil {
			fmt.Fprintf(s.Stdout, "> ")
			continue
		}

		// Just echo command in dry-run mode
		if s.DryRun {
			fmt.Fprintf(s.Stdout, "%s\n> ", line)
			continue
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintln(s.Stderr, "error:", err)
		}
		fmt.Fprintf(s.Stdout, "%s> ", out)
	}

	fmt.Fprintln(s.Stdout, "\nSmell you later!")
}
