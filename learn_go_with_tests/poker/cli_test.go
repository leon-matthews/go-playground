package poker_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"poker"
	"strings"
	"testing"
)

var dummyAlerter = &poker.BlindAlerterMock{}

func TestCLI(t *testing.T) {
	t.Run("record Alyson winning", func(t *testing.T) {
		in := strings.NewReader("Alyson wins\n")
		storage := poker.NewStorageMock(poker.League{})
		cli := poker.NewCLI(storage, in, dummyAlerter)

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
		cli := poker.NewCLI(storage, in, dummyAlerter)

		cli.PlayPoker()

		if len(storage.WinCalls) != 1 {
			t.Fatalf("expected one win call but got %d", len(storage.WinCalls))
		}
		got := storage.WinCalls[0]
		want := "Leon"
		assert.Equalf(t, want, got, "wrong winner, expected %q, got %q", want, got)
	})

	t.Run("schedule printing of blind values", func(t *testing.T) {
		in := strings.NewReader("Alyson Wins\n")
		storage := poker.NewStorageMock(poker.League{})
		alerter := &poker.BlindAlerterMock{}

		cli := poker.NewCLI(storage, in, alerter)
		cli.PlayPoker()

		if len(alerter.Alerts) != 1 {
			t.Fatal("expected one blind alert to be scheduled")
		}
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
