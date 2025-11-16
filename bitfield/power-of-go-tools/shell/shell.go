package shell

import (
	"errors"
	"os/exec"
	"strings"
)

func CmdFromString(s string) (*exec.Cmd, error) {
	f := strings.Fields(s)
	if len(f) < 1 {
		return nil, errors.New("no command given")
	}
	cmd := exec.Command(f[0], f[1:]...)
	return cmd, nil
}
