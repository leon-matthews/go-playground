package main

import (
	"fmt"
	"testing"
)

func TestParent(t *testing.T) {
	t.Parallel()
	defer fmt.Println("Parent exits")

	t.Run("Sub1", func(t *testing.T) {
		defer fmt.Println("Sub1 exits")
	})

	t.Run("Sub2", func(t *testing.T) {
		defer fmt.Println("Sub1 exits")
	})
}

func main() {}
