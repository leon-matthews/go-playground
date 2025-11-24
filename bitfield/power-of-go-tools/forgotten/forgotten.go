package forgotten

import (
	"os/exec"
)

func GetDiffNumStat() (string, error) {
	output, err := exec.Command("git", "diff", "--numstat").CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
