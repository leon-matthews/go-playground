package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"local.dev/pwnedcache/internal/database"
)

// checkCmd holds the arguments of the "check" command.
type checkCmd struct {
	Passwords []string `arg:"" help:"Passwords to check."`
}

// Run prints how many times each password appears in the local database.
func (c *checkCmd) Run(ctx context.Context, global *cli) error {
	return runCheck(ctx, os.Stdout, global.Database, c.Passwords)
}

// runCheck prints a table reporting how many times each password appears in
// the local database.
func runCheck(ctx context.Context, out io.Writer, databasePath string, passwords []string) error {
	queries, db, err := database.OpenReadOnly(ctx, databasePath)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("%w - run 'pwnedcache update' first", err)
		}
		return err
	}
	defer db.Close()

	tw := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "PASSWORD\tBREACHES")
	for _, password := range passwords {
		hash := sha1.Sum([]byte(password))
		count, err := queries.GetHashCount(ctx, hash[:])
		if errors.Is(err, sql.ErrNoRows) {
			count = 0
		} else if err != nil {
			return fmt.Errorf("checking %q: %w", password, err)
		}
		fmt.Fprintf(tw, "%s\t%d\n", password, count)
	}
	return tw.Flush()
}
