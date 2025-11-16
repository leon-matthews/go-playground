package main

import (
	"bufio"
	"fmt"
	"os"

	"shell"
)

func main() {
	fmt.Print("> ")
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Text()
		cmd, err := shell.CmdFromString(line)
		// Ignore empty lines
		if err != nil {
			continue
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Printf("%s", out)
		fmt.Print("\n> ")
	}
	fmt.Println("\nSmell you later!")
}
