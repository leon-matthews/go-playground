package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShortAndTall(t *testing.T) {
	lines := []string{
		"The quick brown fox jumps over the lazy dog",
		"Sphinx of black quartz judge my vow",
		"The five boxing wizards jump quickly",
		"Jackdaws love my big sphinx of quartz",
	}

	short, long := ShortAndTall(lines)
	assert.Equal(t, 35, short)
	assert.Equal(t, 43, long)
}
