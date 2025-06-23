package poker

import (
	"bufio"
	"io"
	"strings"
)

type CLI struct {
	store PlayerStore
	in    io.Reader
}

func (cli *CLI) PlayPoker() {
	reader := bufio.NewScanner(cli.in)
	reader.Scan()
	name := extractName(reader.Text())
	cli.store.RecordWin(name)
}

func extractName(text string) string {
	return strings.Replace(text, " wins", "", 1)
}
