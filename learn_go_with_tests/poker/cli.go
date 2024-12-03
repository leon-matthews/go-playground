package poker

import (
	"bufio"
	"io"
	"strings"
	"time"
)

// BlindAlerter alerts users to blind increases after some time interval
type BlindAlerter interface {
	ScheduleAlert(duration time.Duration, amount int)
}

// CLI manages the command-line interface
type CLI struct {
	storage PlayerStorage
	scanner *bufio.Scanner
	alerter BlindAlerter
}

func NewCLI(storage PlayerStorage, in io.Reader, alerter BlindAlerter) *CLI {
	return &CLI{
		storage: storage,
		scanner: bufio.NewScanner(in),
		alerter: alerter,
	}
}

func (cli *CLI) PlayPoker() {
	cli.alerter.ScheduleAlert(5*time.Second, 100)
	winner := ExtractWinner(cli.readLine())
	cli.storage.RecordWin(winner)
}

func ExtractWinner(line string) string {
	return strings.Replace(line, " wins", "", 1)
}

func (cli *CLI) readLine() string {
	cli.scanner.Scan()
	return cli.scanner.Text()
}
