package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MyNumber struct{ int }

// Lots of assert functions!
// https://pkg.go.dev/github.com/stretchr/testify/assert
func TestAssert(t *testing.T) {
	t.Run("compare", func(t *testing.T) {
		assert.Equal(t, 123, 123, "they should be equal")
		assert.NotEqual(t, 33, 44, "they should not be equal")
	})

	t.Run("nil", func(t *testing.T) {
		var number *MyNumber
		assert.Nil(t, number)
		assert.NotNil(t, MyNumber{})
	})

	t.Run("safe check", func(t *testing.T) {
		var number = &MyNumber{42}
		if assert.NotNil(t, number) {
			assert.Equal(t, number, &MyNumber{42})
		}
	})
}

// Same assertions as `assert`, but stops test execution when a test fails.
func TestRequire(t *testing.T) {
	t.Run("compare", func(t *testing.T) {
		require.Equal(t, 22, 22, "they should be equal")
		require.NotEqual(t, 22, 33, "they should not be equal")
	})
}

// The mock package provides `Mock`, that tracks activity on another object.
func TestMock(t *testing.T) {
}

// Mock object is embedded into a test object
type MyTestObject struct {
	mock.Mock
}
