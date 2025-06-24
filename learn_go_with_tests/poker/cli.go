package poker

import (
	"bufio"
	"io"
	"strings"
	"time"
)

type Alerter interface {
	Schedule(at time.Duration, amount int)
}

type CLI struct {
	store   PlayerStore
	in      *bufio.Scanner
	alerter Alerter
}

func NewCLI(store PlayerStore, in io.Reader, alerter Alerter) *CLI {
	return &CLI{
		store:   store,
		in:      bufio.NewScanner(in),
		alerter: alerter,
	}
}

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

func (cli *CLI) readLine() string {
	cli.in.Scan()
	return cli.in.Text()
}

func extractName(text string) string {
	return strings.Replace(text, " wins", "", 1)
}
