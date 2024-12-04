package poker_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"poker"
	"strings"
	"testing"
	"time"
)

type scheduledAlert struct {
	at     time.Duration
	amount int
}

func (s scheduledAlert) String() string {
	return fmt.Sprintf("%d chips at %v", s.amount, s.at)
}

// BlindAlerterMock fakes blind altering for testing
type BlindAlerterMock struct {
	alerts []scheduledAlert
}

func (b *BlindAlerterMock) ScheduleAlert(at time.Duration, amount int) {
	b.alerts = append(b.alerts, scheduledAlert{at, amount})
}

var dummyAlerter = &BlindAlerterMock{}

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
		alerter := &BlindAlerterMock{}

		cli := poker.NewCLI(storage, in, alerter)
		cli.PlayPoker()

		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{10 * time.Minute, 200},
			{20 * time.Minute, 300},
			{30 * time.Minute, 400},
			{40 * time.Minute, 500},
			{50 * time.Minute, 600},
			{60 * time.Minute, 800},
			{70 * time.Minute, 1_000},
			{80 * time.Minute, 2_000},
			{90 * time.Minute, 4_000},
			{100 * time.Minute, 8_000},
		}

		for i, want := range cases {
			t.Run(fmt.Sprint(want), func(t *testing.T) {
				if len(alerter.alerts) <= i {
					t.Fatalf("alert %d was not scheduled %v", i, alerter.alerts)
				}
				got := alerter.alerts[i]
				assertScheduledAlert(t, got, want)
			})
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

func assertScheduledAlert(t testing.TB, got, want scheduledAlert) {
	t.Helper()
	if got.amount != want.amount {
		t.Errorf("got amount %d, want %d", got.amount, want.amount)
	}

	if got.at != want.at {
		t.Errorf("got scheduled time %d, want %d", got.at, want.at)
	}
}
