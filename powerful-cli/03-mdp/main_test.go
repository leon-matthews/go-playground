package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseContent(t *testing.T) {
	input := readTestFile(t, "test.md")

	got := parseContent(input)
	want := readTestFile(t, "test.md.html")

	assert.Equal(t, string(want), string(got))
}

// readTestFile reads string from file from `testdata` folder
func readTestFile(t testing.TB, filename string) []byte {
	t.Helper()
	relpath := fmt.Sprintf("./testdata/%s", filename)
	contents, err := os.ReadFile(relpath)
	if err != nil {
		t.Fatal("could not read file:", err)
	}
	return contents
}
