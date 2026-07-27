package main

import (
	"github.com/spf13/cobra"

	"local.dev/cine/importer"
)

// newBuildDatabaseCmd builds the "build-database" sub-command.
func newBuildDatabaseCmd() *cobra.Command {
	var allowAdult, allowUnrated, people bool

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
build.

Only the titles themselves are imported by default. Pass --people to import
name.basics, title.crew and title.principals as well, which roughly trebles the
size of the database. Every option is recorded in the build_info table, so a
query can tell what a database was never given.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			folder, out := args[0], args[1]
			if err := requireFolder(folder); err != nil {
				return err
			}
			// The filter flags name what to keep and the options what to restrict, so
			// each inverts; --people names what to add and so does not.
			options := importer.BuildOptions{
				Rated:    !allowUnrated,
				NotAdult: !allowAdult,
				People:   people,
			}
			return importer.Import(cmd.Context(), out, folder, options, newLogger())
		},
	}

	cmd.Flags().BoolVar(&allowAdult, "allow-adult", false,
		"keep titles IMDb flags as adult")
	cmd.Flags().BoolVar(&allowUnrated, "allow-unrated", false,
		"keep titles IMDb has published no rating for")
	cmd.Flags().BoolVar(&people, "people", false,
		"import the people datasets: names, crew and principals")
	return cmd
}
