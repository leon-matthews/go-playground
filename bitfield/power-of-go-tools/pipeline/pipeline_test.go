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

func TestColumn_ProducesNothingWhenPipeErrorSet(t *testing.T) {
	t.Parallel()
	p := pipeline.FromString("1 2 3\n")
	p.Error = errors.New("oh no")
	// We can't just call String() here, as that already returns nothing on error
	data, err := io.ReadAll(p.Column(1).Reader)
	require.NoError(t, err)
	if len(data) > 0 {
		t.Errorf("want no output from Column after error, but got %q", data)
	}
}

func TestColumn_SelectsColumn2Of3(t *testing.T) {
	t.Parallel()
	input := "1 2 3\n1 2 3\n1 2 3\n"
	p := pipeline.FromString(input)
	want := "2\n2\n2\n"
	got, err := p.Column(2).String()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestColumn_SetsErrorAndProducesNothingGivenInvalidArg(t *testing.T) {
	t.Parallel()
	p := pipeline.FromString("1 2 3\n1 2 3\n1 2 3\n")
	p.Column(-1)
	if p.Error == nil {
		t.Error("want error on non-positive Column, but got nil")
	}
	data, err := io.ReadAll(p.Column(1).Reader)
	require.NoError(t, err)
	if len(data) > 0 {
		t.Errorf("want no output from Column with invalid col, but got %q", data)
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

func TestStringReturnsPipeContents(t *testing.T) {
	t.Parallel()
	want := "Hello, world\n"
	p := pipeline.FromString(want)
	got, err := p.String()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestStringReturnsErrorWhenPipeErrorSet(t *testing.T) {
	t.Parallel()
	p := pipeline.FromString("Hello, world\n")
	p.Error = errors.New("oh no")
	_, err := p.String()
	assert.EqualError(t, err, "oh no")
}
