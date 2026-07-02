// Command pwneddb downloads the Have I Been Pwned password database to SQLite.
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"pwneddb/database"
	"pwneddb/pwned"
)

// Local SQLite file holding download metadata and hash lists
const databasePath = "pwned.db"

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// Cancel the run cleanly on Ctrl-C; state is safe in the database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	queries, db, err := database.Open(ctx, databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	downloader := pwned.NewDownloader(queries)
	if err := downloader.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
