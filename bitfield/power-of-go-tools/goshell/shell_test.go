package shell_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shell"
)

func TestCmdFromString(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		input := "/bin/ls -l shell.go"
		want := exec.Command("/bin/ls", "-l", "shell.go")
		got, err := shell.CmdFromString(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty", func(t *testing.T) {
		input := ""
		got, err := shell.CmdFromString(input)
		assert.ErrorContains(t, err, "no command given")
		assert.Nil(t, got)
	})
}
