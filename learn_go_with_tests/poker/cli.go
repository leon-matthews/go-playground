package poker

import (
	"bufio"
	"io"
	"strings"
	"time"
)

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
	blinds := []int{100, 200, 300, 400, 500, 600, 800, 1000, 2000, 4000, 8000}
	blindTime := 0 * time.Second
	for _, amount := range blinds {
		cli.alerter.ScheduleAlert(blindTime, amount)
		blindTime = blindTime + 10*time.Second
	}
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
