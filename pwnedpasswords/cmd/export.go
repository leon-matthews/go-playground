package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"pwnedpasswords/database"
)

// newExportCmd builds the "export" sub-command.
func newExportCmd() *cobra.Command {
	var top int
	var format string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write the most-breached passwords as a denylist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExport(cmd.Context(), cmd.OutOrStdout(), databasePath, top, format)
		},
	}
	cmd.Flags().IntVarP(&top, "top", "n", 1000, "number of passwords to write")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "output format: text or json")
	return cmd
}

// denylistEntry is one row of JSON denylist output.
type denylistEntry struct {
	Password string `json:"password"`
	Count    int64  `json:"count"`
}

// runExport writes the top passwords by breach count in the chosen format.
func runExport(ctx context.Context, out io.Writer, dbPath string, top int, format string) error {
	queries, db, err := database.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := queries.TopPasswords(ctx, int64(top))
	if err != nil {
		return err
	}

	switch format {
	case "text":
		for _, row := range rows {
			if _, err := fmt.Fprintln(out, row.Password); err != nil {
				return err
			}
		}
	case "json":
		entries := make([]denylistEntry, len(rows))
		for i, row := range rows {
			entries[i] = denylistEntry{Password: row.Password, Count: row.Count}
		}
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(entries); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q: use text or json", format)
	}
	return nil
}
