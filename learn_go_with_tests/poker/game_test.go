package poker_test

import (
	"fmt"
	"poker"
	"testing"
	"time"
)

func TestGame_Start(t *testing.T) {
	t.Run("schedules Alerts on game start for 5 players", func(t *testing.T) {
		blindAlerter := &BlindAlerterMock{}
		game := poker.NewGame(blindAlerter, poker.NewInMemoryStorage())

		game.Start(5)

		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{10 * time.Minute, 0},
			{20 * time.Minute, 0},
			{30 * time.Minute, 0},
			{40 * time.Minute, 0},
			{50 * time.Minute, 0},
			{60 * time.Minute, 0},
			{70 * time.Minute, 0},
			{80 * time.Minute, 0},
			{90 * time.Minute, 0},
			{100 * time.Minute, 0},
		}

		for i, want := range cases {
			t.Run(fmt.Sprint(want), func(t *testing.T) {
				if len(game.Alerter.Alerts) <= i {
					t.Fatalf("alert %d was not scheduled %v", i, game.Alerter.Alerts)
				}
				got := game.Alerter.Alerts[i]
				assertScheduledAlert(t, got, want)
			})
		}
	})

	t.Run("schedules Alerts on game start for 7 players", func(t *testing.T) {
		blindAlerter := &poker.SpyBlindAlerter{}
		game := poker.NewGame(blindAlerter, dummyPlayerStore)

		game.Start(7)

		cases := []poker.ScheduledAlert{
			{At: 0 * time.Second, Amount: 100},
			{At: 12 * time.Minute, Amount: 200},
			{At: 24 * time.Minute, Amount: 300},
			{At: 36 * time.Minute, Amount: 400},
		}

		checkSchedulingCases(cases, t, blindAlerter)
	})

}

func TestGame_Finish(t *testing.T) {
	store := &poker.StubPlayerStore{}
	game := poker.NewGame(dummyBlindAlerter, store)
	winner := "Ruth"

	game.Finish(winner)
	poker.AssertPlayerWin(t, store, winner)
}
