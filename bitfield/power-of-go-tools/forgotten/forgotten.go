package forgotten

import (
	"os/exec"
	"strings"
)

type Status struct {
	NumFiles int
}

func GetDiffNumStat() (string, error) {
	output, err := exec.Command("git", "diff", "--numstat").CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ParseDiffNumStat builds a Status from the string output of 'git diff --numstat' command
func ParseDiffNumStat(s string) (Status, error) {
	numFiles := 0
	for range strings.Lines(s) {
		numFiles++
	}
	status := Status{
		NumFiles: numFiles,
	}
	return status, nil
}
