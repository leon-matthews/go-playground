package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"pwncache/database"
)

// newCheckCmd builds the "check" sub-command, which checks passwords against
// the local database.
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <password>...",
		Short: "Check one or more passwords against the local database",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer logs.logFile.Close()
			return runCheck(cmd.Context(), cmd.OutOrStdout(), args)
		},
	}
	return cmd
}

// runCheck prints a table reporting how many times each password appears in
// the local database.
func runCheck(ctx context.Context, out io.Writer, passwords []string) error {
	queries, db, err := database.Open(ctx, databasePath)
	if err != nil {
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
