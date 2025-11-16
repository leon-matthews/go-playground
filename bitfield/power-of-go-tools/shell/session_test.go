package shell_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"shell"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionCreatesExpectedSession(t *testing.T) {
	t.Parallel()
	want := shell.Session{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	got := *shell.NewSession(os.Stdin, os.Stdout, os.Stderr)
	if want != got {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func TestRunProducesExpectedOutput(t *testing.T) {
	t.Parallel()
	in := strings.NewReader("echo Hello World\n\n")
	want := "> echo Hello World\n> > \nSmell you later!\n"
	out := new(bytes.Buffer)
	session := shell.NewSession(in, out, io.Discard)
	session.DryRun = true
	session.Run()
	got := out.String()
	assert.Equal(t, want, got)
}
