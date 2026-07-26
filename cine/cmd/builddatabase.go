package main

import (
	"github.com/spf13/cobra"

	"local.dev/cine/importer"
)

// newBuildDatabaseCmd builds the "build-database" sub-command.
func newBuildDatabaseCmd() *cobra.Command {
	var allowAdult, allowUnrated bool

	cmd := &cobra.Command{
		Use:   "build-database <imdb-data-folder> <output.db>",
		Short: "Import the IMDb dataset files into a new SQLite database",
		Long: `Import the IMDb dataset files into a new SQLite database.

The database is built in a temporary file and renamed into place only once the
whole import succeeds, so an interrupted or failed run leaves any database
already at the output path untouched. Per-file progress and the total build time
are logged as the import runs.

Both filters are applied by default, keeping a title only if IMDb has published
a rating for it and has not flagged it as adult. An episode is kept exactly when
its parent series is kept, whichever rule decided that, so a series is never
stored with gaps in it. Pass both --allow-adult and --allow-unrated for a full
build.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			folder, out := args[0], args[1]
			if err := requireFolder(folder); err != nil {
				return err
			}
			// The flags name what to keep and the rules what to restrict, so each inverts.
			rules := importer.FilterRules{Rated: !allowUnrated, NotAdult: !allowAdult}
			return importer.Import(cmd.Context(), out, folder, rules, newLogger())
		},
	}

	cmd.Flags().BoolVar(&allowAdult, "allow-adult", false,
		"keep titles IMDb flags as adult")
	cmd.Flags().BoolVar(&allowUnrated, "allow-unrated", false,
		"keep titles IMDb has published no rating for")
	return cmd
}
