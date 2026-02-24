# Re-exec testing Go subprocesses

A better way to test programs that exec external binaries; a better form
of mocking.

https://rednafi.com/go/test-subprocesses/

For this to work you need to split up running an external binary into two
parts:

1. Run exec.Command() with args to build a *exec.Cmd
2. Pass the *exec.Cmd into a handler that checks outputs and error conditions.

It is the second part that this technique allows us to test.
