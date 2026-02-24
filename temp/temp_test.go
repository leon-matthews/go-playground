package main_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorIs(t *testing.T) {
	var err error
	assert.ErrorIs(t, err, os.ErrNotExist)
}
