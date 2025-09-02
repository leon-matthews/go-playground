package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"

	_ "modernc.org/sqlite"

	"sqlc_tutorial/app/authors"
)

//go:embed schema.sql
var ddl string

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return err
	}

	// Create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}

	queries := authors.New(db)

	// List authors
	rows, err := queries.ListAuthors(ctx)
	if err != nil {
		return err
	}
	log.Println(rows)

	// Create an author
	insertedAuthor, err := queries.CreateAuthor(ctx, authors.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio:  sql.NullString{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
	})
	if err != nil {
		return err
	}
	log.Println(insertedAuthor)

	// Get the author we just inserted
	fetchedAuthor, err := queries.GetAuthor(ctx, insertedAuthor.ID)
	if err != nil {
		return err
	}
	log.Println(fetchedAuthor)

	return nil
}
