package poker_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"poker"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	t.Run("record Alyson winning", func(t *testing.T) {
		in := strings.NewReader("Alyson wins\n")
		storage := poker.NewStorageMock(poker.League{})
		cli := poker.NewCLI(storage, in)

		cli.PlayPoker()

		if len(storage.WinCalls) != 1 {
			t.Fatalf("expected one win call but got %d", len(storage.WinCalls))
		}
		got := storage.WinCalls[0]
		want := "Alyson"
		assert.Equalf(t, want, got, "wrong winner, expected %q, got %q", want, got)
	})

	t.Run("record Leon winning", func(t *testing.T) {
		in := strings.NewReader("Leon wins\n")
		storage := poker.NewStorageMock(poker.League{})
		cli := poker.NewCLI(storage, in)

		cli.PlayPoker()

		if len(storage.WinCalls) != 1 {
			t.Fatalf("expected one win call but got %d", len(storage.WinCalls))
		}
		got := storage.WinCalls[0]
		want := "Leon"
		assert.Equalf(t, want, got, "wrong winner, expected %q, got %q", want, got)
	})
}

func TestExtractWinner(t *testing.T) {
	cases := []struct{ given, want string }{
		{"Leon wins", "Leon"},
		{"Alyson wins", "Alyson"},
	}

	for _, c := range cases {
		name := fmt.Sprintf("extract %s", strings.ToLower(c.want))
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.want, poker.ExtractWinner(c.given))
		})
	}
}
