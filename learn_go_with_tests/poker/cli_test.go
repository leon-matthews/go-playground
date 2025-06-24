package poker_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"learn_go_with_tests/poker"
)

func TestCLI(t *testing.T) {
	var mockAlerter = &MockAlerter{}

	t.Run("Alyson wins", func(t *testing.T) {
		in := strings.NewReader("Alyson wins\n")
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in, mockAlerter)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, store, "Alyson")
	})

	t.Run("Leon wins", func(t *testing.T) {
		in := strings.NewReader("Leon wins\n")
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in, mockAlerter)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, store, "Leon")
	})

	t.Run("it schedules printing of blind values", func(t *testing.T) {
		in := strings.NewReader("Chris wins\n")
		store := poker.NewPlayerStoreMock()
		alerter := &MockAlerter{}

		cli := poker.NewCLI(store, in, alerter)
		cli.PlayPoker()

		if len(alerter.alerts) != 1 {
			t.Fatal("expected a blind alert to be scheduled")
		}
		fmt.Printf("[%T]%+[1]v\n", alerter)
		t.Fail()
	})
}

type MockAlerter struct {
	alerts []struct {
		at     time.Duration
		amount int
	}
}

func (m *MockAlerter) Schedule(at time.Duration, amount int) {
	alert := struct {
		at     time.Duration
		amount int
	}{at, amount}
	m.alerts = append(m.alerts, alert)
}
