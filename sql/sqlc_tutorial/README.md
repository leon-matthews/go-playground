
# SQLC Tutorial

The [sqlc](https://docs.sqlc.dev/en/latest/index.html) project uses a compiler
to create Go types and functions directly from SQL queries. It seems to fill 
a niche between using ``database/sql`` directly, and an ORM like [GORM](https://gorm.io/).


## Migrations?

Use something like [dbmate](https://github.com/amacneil/dbmate?tab=readme-ov-file#command-line-options) 
or [Goose](https://github.com/pressly/goose) to create tables and handle migrations.

If only using SQLite, an alternative might be [zombiezen.com/go/sqlite](https://github.com/zombiezen/go-sqlite),
which provides basic schema migration, but also various SQLite-specific functionality.
