package main

import (
	"testing"
	"testing/synctest"
)

func TestBlocking(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		trickyOrdering()
	})
}
