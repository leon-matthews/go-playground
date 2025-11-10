//go:build integration

package forgotten_test

import (
	"os/exec"
	"testing"

	"forgotten"
)

func TestGetDiffNumStat(t *testing.T) {
	t.Parallel()
	_, err := exec.Command("git", "diff").CombinedOutput()
	if err != nil {
		t.Skip("git not installed:", err)
	}

	text, err := forgotten.GetDiffNumStat()
	status, err := forgotten.ParseDiffNumStat(text)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Status: %v", status)
}
