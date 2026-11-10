package main

// cli is the whole command line: global flags, then one field per command.
type cli struct {
	Database string `short:"d" default:"pwnedcache.db" help:"Path to the SQLite database."`
	Verbose  bool   `short:"v" help:"Debug-level logging."`

	Update updateCmd `cmd:"" help:"Download or refresh the local password hash database."`
	Check  checkCmd  `cmd:"" help:"Check one or more passwords against the local database."`
}
