package poker

import (
	"bufio"
	"io"
	"strings"
)

type CLI struct {
	storage PlayerStorage
	scanner *bufio.Scanner
}

func NewCLI(storage PlayerStorage, in io.Reader) *CLI {
	return &CLI{
		storage: storage,
		scanner: bufio.NewScanner(in),
	}
}

func (cli *CLI) PlayPoker() {
	cli.scanner.Scan()
	winner := ExtractWinner(cli.scanner.Text())
	cli.storage.RecordWin(winner)
}

func ExtractWinner(line string) string {
	return strings.Replace(line, " wins", "", 1)
}
