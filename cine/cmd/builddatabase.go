package main

import (
	"github.com/spf13/cobra"

	"local.dev/cine/importer"
)

// newBuildDatabaseCmd builds the "build-database" sub-command.
func newBuildDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build-database <imdb-data-folder> <output.db>",
		Short: "Import the IMDb dataset files into a new SQLite database",
		Long: `Import the IMDb dataset files into a new SQLite database.

The database is built in a temporary file and renamed into place only once the
whole import succeeds, so an interrupted or failed run leaves any database
already at the output path untouched. Per-file progress and the total build time
are logged as the import runs.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			folder, out := args[0], args[1]
			if err := requireFolder(folder); err != nil {
				return err
			}
			return importer.Import(cmd.Context(), out, folder, newLogger())
		},
	}
}
