package poker_test

import (
	"strings"
	"testing"

	"learn_go_with_tests/poker"
)

func TestCLI(t *testing.T) {
	t.Run("Alyson wins", func(t *testing.T) {
		in := strings.NewReader("Alyson wins\n")
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, store, "Alyson")
	})

	t.Run("Leon wins", func(t *testing.T) {
		in := strings.NewReader("Leon wins\n")
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, store, "Leon")
	})
}
