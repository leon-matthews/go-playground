// Command pwnedcache maintains and queries a local mirror of the Have I Been Pwned password database.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"
)

// Exit codes: 1 for a failed command, 2 for a bad command line
const (
	exitFailure = 1
	exitUsage   = 2
)

func main() {
	// Cancel the run cleanly on Ctrl-C; state is safe in the database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:])
	stop()
	os.Exit(code)
}

// run parses the command line, builds what the chosen command needs, and runs it.
func run(ctx context.Context, args []string) int {
	// Phase 1: parse only; --help prints and exits here
	var root cli
	parser := kong.Must(&root,
		kong.Name("pwnedcache"),
		kong.Description("Maintain and query a local mirror of the Have I Been Pwned password database."),
	)
	kctx, err := parser.Parse(args)
	if err != nil {
		if parseErr, ok := errors.AsType[*kong.ParseError](err); ok {
			_ = parseErr.Context.PrintUsage(false)
		}
		parser.Errorf("%s", err)
		return exitUsage
	}

	// Phase 2: dependencies, now that the command and flags are known
	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&root)
	if kctx.Selected().Name == "update" {
		logs, err := setupLogging(root.Verbose)
		if err != nil {
			slog.Error("logging setup", "error", err)
			return exitFailure
		}
		defer logs.Close()
		kctx.Bind(logs)
	}

	// Phase 3: run; a failure is logged before the deferred close above
	if err := kctx.Run(); err != nil {
		slog.Error("command failed", "error", err)
		return exitFailure
	}
	return 0
}
