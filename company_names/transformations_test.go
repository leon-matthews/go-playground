package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToAscii(t *testing.T) {
	t.Run("too easy", func(t *testing.T) {
		got, err := ToAscii("banana")
		assert.NoError(t, err)
		assert.Equal(t, "banana", got)
	})
}
