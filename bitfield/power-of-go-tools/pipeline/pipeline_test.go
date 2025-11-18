package pipeline_test

import (
	"bytes"
	"errors"
	"io"
	"pipeline"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdoutPrintsMessageToOutput(t *testing.T) {
	t.Parallel()
	want := "Hello, world\n"
	p := pipeline.FromString(want)
	buf := new(bytes.Buffer)
	p.Writer = buf
	p.Stdout()
	require.NoError(t, p.Error)
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestStdoutPrintsNothingOnError(t *testing.T) {
	// By design, pipeline stages don't do anything at all if in error
	t.Parallel()
	p := pipeline.FromString("Hello, world\n")
	p.Error = errors.New("oh no")
	buf := new(bytes.Buffer)
	p.Writer = buf
	p.Stdout()
	got := buf.String()
	if got != "" {
		t.Errorf("want no output from Stdout after error, but got %q", got)
	}
}

func TestFromFile_ReadsAllDataFromFile(t *testing.T) {
	t.Parallel()
	want := []byte("Hello, world!\n")
	p := pipeline.FromFile("testdata/hello.txt")
	require.NoError(t, p.Error)
	got, err := io.ReadAll(p.Reader)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestFromFile_SetsErrorGivenNonexistentFile(t *testing.T) {
	t.Parallel()
	p := pipeline.FromFile("no-such-file.txt")
	assert.Error(t, p.Error, "error expected but was nil")
}
