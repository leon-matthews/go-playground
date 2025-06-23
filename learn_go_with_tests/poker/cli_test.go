package poker

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	t.Run("Alyson wins", func(t *testing.T) {
		in := strings.NewReader("Alyson wins\n")
		store := NewPlayerStoreMock()
		cli := &CLI{store, in}
		cli.PlayPoker()
		assertPlayerWin(t, store, "Alyson")
	})

	t.Run("Leon wins", func(t *testing.T) {
		in := strings.NewReader("Leon wins\n")
		store := NewPlayerStoreMock()
		cli := &CLI{store, in}
		cli.PlayPoker()
		assertPlayerWin(t, store, "Leon")
	})
}

func assertPlayerWin(t *testing.T, store *PlayerStoreMock, name string) {
	t.Helper()
	if len(store.winCalls) != 1 {
		t.Fatal("expected a win call but didn't get any")
	}
	assert.Equal(t, name, store.winCalls[0])
}
