package poker_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"learn_go_with_tests/poker"
	"strings"
	"testing"
	"time"
)

func TestCLI(t *testing.T) {
	var mockAlerter = &poker.AlerterMock{}

	t.Run("reads 'Alyson wins' from input", func(t *testing.T) {
		in := strings.NewReader("Alyson wins\n")
		out := &bytes.Buffer{}
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in, out, mockAlerter)
		err := cli.PlayPoker()
		require.NoError(t, err)
		poker.AssertPlayerWin(t, store, "Alyson")
	})

	t.Run("reads 'Leon wins' from input", func(t *testing.T) {
		in := strings.NewReader("Leon wins\n")
		out := &bytes.Buffer{}
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in, out, mockAlerter)
		err := cli.PlayPoker()
		require.NoError(t, err)
		poker.AssertPlayerWin(t, store, "Leon")
	})

	t.Run("schedule printing of blind values", func(t *testing.T) {
		in := strings.NewReader("Chris wins\n")
		out := &bytes.Buffer{}
		store := poker.NewPlayerStoreMock()
		alerter := &poker.AlerterMock{}

		cli := poker.NewCLI(store, in, out, alerter)
		err := cli.PlayPoker()
		require.NoError(t, err)

		cases := []struct {
			expectedTime   time.Duration
			expectedAmount int
		}{
			{0 * time.Second, 100},
			{10 * time.Minute, 200},
			{20 * time.Minute, 300},
			{30 * time.Minute, 400},
			{40 * time.Minute, 500},
			{50 * time.Minute, 600},
			{60 * time.Minute, 800},
			{70 * time.Minute, 1000},
			{80 * time.Minute, 2000},
			{90 * time.Minute, 4000},
			{100 * time.Minute, 8000},
		}

		for i, c := range cases {
			name := fmt.Sprintf("alert %d", i)
			t.Run(name, func(t *testing.T) {
				if len(alerter.Alerts) <= i {
					t.Fatalf("%s for $%d at %s", name, c.expectedAmount, c.expectedTime)
				}
				alert := alerter.Alerts[i]
				assert.Equal(t, c.expectedAmount, alert.Amount)
				assert.Equal(t, c.expectedTime, alert.At)
			})
		}
	})

	t.Run("prompt user for number of players", func(t *testing.T) {
		alerter := &poker.AlerterMock{}
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		store := poker.NewPlayerStoreMock()
		cli := poker.NewCLI(store, in, out, alerter)

		cli.PlayPoker()

		assert.Equal(t, poker.NumPlayerPrompt, out.String())
	})
}
