package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	"local.dev/pwnedcache/internal/database"
	"local.dev/pwnedcache/internal/database/sqlite"
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
		count, err := breaches(ctx, queries, password)
		if err != nil {
			return fmt.Errorf("checking %q: %w", password, err)
		}
		fmt.Fprintf(tw, "%s\t%s\n", password, count)
	}
	return tw.Flush()
}

// breaches returns the breach count for password, or "not downloaded" if its hash list is missing.
func breaches(ctx context.Context, queries *sqlite.Queries, password string) (string, error) {
	hash := sha1.Sum([]byte(password))
	count, err := queries.GetHashCount(ctx, hash[:])
	if err == nil {
		return strconv.FormatInt(count, 10), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	// An absent hash only means zero breaches if its prefix's list was stored
	prefix := hex.EncodeToString(hash[:3])[:5]
	if _, err := queries.GetPrefix(ctx, prefix); errors.Is(err, sql.ErrNoRows) {
		return "not downloaded", nil
	} else if err != nil {
		return "", err
	}
	return "0", nil
}
