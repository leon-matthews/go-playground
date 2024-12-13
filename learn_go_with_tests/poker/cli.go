package poker

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const NumPlayersPrompt = "Please enter the number of players: "

// CLI manages the command-line interface
type CLI struct {
	in   *bufio.Scanner
	out  io.Writer
	game *Game
}

func NewCLI(in io.Reader, out io.Writer, game *Game) *CLI {
	return &CLI{
		in:   bufio.NewScanner(in),
		out:  out,
		game: game,
	}
}

func (cli *CLI) PlayPoker() {
	fmt.Fprint(cli.out, NumPlayersPrompt)

	numPlayersInput := cli.readLine()
	numPlayers, _ := strconv.Atoi(strings.Trim(numPlayersInput, "\n"))

	cli.game.Start(numPlayers)

	winnerInput := cli.readLine()
	winner := ExtractWinner(winnerInput)

	cli.game.Finish(winner)
}

func ExtractWinner(line string) string {
	return strings.Replace(line, " wins", "", 1)
}

func (cli *CLI) readLine() string {
	cli.in.Scan()
	return cli.in.Text()
}
