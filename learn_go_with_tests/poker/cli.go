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

func (cli *CLI) PlayPoker() {
	cli.alerter.Schedule(5*time.Second, 100)
	line := cli.readLine()
	name := extractName(line)
	cli.store.RecordWin(name)
}

func (cli *CLI) readLine() string {
	cli.in.Scan()
	return cli.in.Text()
}

func extractName(text string) string {
	return strings.Replace(text, " wins", "", 1)
}
