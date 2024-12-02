package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestTape_Write(t *testing.T) {
	file, clean := createTempFile(t, "12345")
	defer clean()

	tape := &tape{file}
	tape.Write([]byte("abc"))

	file.Seek(0, io.SeekStart)
	newContents, _ := io.ReadAll(file)

	got := string(newContents)
	want := "abc"
	assert.Equal(t, want, got)
}
