# Database layer

The status of the download (its metadata), and optionally the password hashes 
themselves are kept in an SQLite3 database.

## sqlc

The database layer is implemented using [sqlc](https://sqlc.dev/). It is a 
code generator which produces "fully type-safe idiomatic Go code" from SQL.

    $ go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

In our case, it reads the `schema.sql` and `queries.sql` files and produces
the `database/sqlite` package.

    $ cd database
    $ ls
    query.sql schema.sql sqlc.yaml sqlite/
    $ sqlc generate
    $ ls sqlite/
    db.go  models.go  query.sql.go
