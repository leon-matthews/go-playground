package forgotten

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type NumStat []Change

type Change struct {
	LinesAdded   int
	LinesRemoved int
	FileName     string
}

func (status NumStat) ToJSON() string {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func GetDiffNumStat() (string, error) {
	output, err := exec.Command("git", "diff", "--numstat").CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ParseDiffNumStat builds a Status from the string output of 'git diff --numstat' command
func ParseDiffNumStat(s string) (NumStat, error) {
	numStat := NumStat{}
	for line := range strings.Lines(s) {
		change, err := parse(line)
		if err != nil {
			return numStat, err
		}
		numStat = append(numStat, change)
	}
	fmt.Printf("[%T]%+[1]v\n", numStat)
	return numStat, nil
}

// parse produces a charge record from a single line of input
func parse(line string) (Change, error) {
	f := strings.Fields(line)
	if len(f) != 3 {
		return Change{}, fmt.Errorf("expected 3 fields, got %d: %s", len(f), line)
	}

	linesAdded, err := strconv.Atoi(f[0])
	if err != nil {
		return Change{}, fmt.Errorf("could not parse lines added: %w: %s", err, line)
	}

	linesRemoved, err := strconv.Atoi(f[1])
	if err != nil {
		return Change{}, fmt.Errorf("could not parse lines removed: %w: %s", err, line)
	}

	return Change{linesAdded, linesRemoved, f[2]}, nil
}
