package poker

import (
	"bufio"
	"io"
	"strings"
	"time"
)

// Alerter schedules alerts for points in the future
type Alerter interface {
	// Schedule an alert for the change to the blind to amount dollars
	Schedule(at time.Duration, amount int)
}

// CLI is the top-level struct for the poker client
type CLI struct {
	store   PlayerStore
	in      *bufio.Scanner
	alerter Alerter
}

// NewCLI returns a pointer to a new CLI
func NewCLI(store PlayerStore, in io.Reader, alerter Alerter) *CLI {
	return &CLI{
		store:   store,
		in:      bufio.NewScanner(in),
		alerter: alerter,
	}
}

// PlayPoker begins a game
func (cli *CLI) PlayPoker() error {
	// Schedule alerts for blind increases
	blinds := []int{100, 200, 300, 400, 500, 600, 800, 1000, 2000, 4000, 8000}
	blindTime := 0 * time.Second
	for _, blind := range blinds {
		cli.alerter.Schedule(blindTime, blind)
		blindTime = blindTime + (10 * time.Minute)
	}

	line := cli.readLine()
	name := extractName(line)
	err := cli.store.RecordWin(name)
	if err != nil {
		return err
	}
	return nil
}

// readLine reads the next line from the input scanner
func (cli *CLI) readLine() string {
	cli.in.Scan()
	return cli.in.Text()
}

// extractName finds and returns the player's name from a line of input
func extractName(text string) string {
	return strings.Replace(text, " wins", "", 1)
}
