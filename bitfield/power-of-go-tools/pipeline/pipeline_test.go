package pipeline_test

import (
	"bytes"
	"errors"
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
	p.Output = buf
	p.Stdout()
	require.NoError(t, p.Error)
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestStdoutPrintsNothingOnError(t *testing.T) {
	t.Parallel()
	p := pipeline.FromString("Hello, world\n")
	p.Error = errors.New("oh no")
	buf := new(bytes.Buffer)
	p.Output = buf
	p.Stdout()
	got := buf.String()
	if got != "" {
		t.Errorf("want no output from Stdout after error, but got %q", got)
	}
}
