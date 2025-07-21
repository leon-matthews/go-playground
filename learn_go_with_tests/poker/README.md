# Poker

Sample app from Learn [Go with Tests](https://quii.gitbook.io/learn-go-with-tests)

## TODO

* Continue from chapter on [Time](https://quii.gitbook.io/learn-go-with-tests/build-an-application/time#write-the-test-first-3)
* Play around with [BadgerDB](https://docs.hypermode.com/badger/overview)

## BoltDB

Bolt is a pure-GO key-value store inspired by LMDB. Features:

1. Database files are a single file
2. Extremely fast reads, slower writes.
3. Database files are locked, so that one application can open a file at a time.
4. Key/value pairs are contained in a 'bucket'.
5. A database contains mulitple buckets.
6. Buckets are organised in a tree, like folders in a file system.
7. Keys and values are both byte-strings

### BoltDB CLI Utilities

Install `bbolt` command-line utility:

	$ go install go.etcd.io/bbolt/cmd/bbolt@latest

Optionally install tool to export to JSON/YAML:

	$ go install github.com/konoui/boltdb-exporter@latest
	$ boltdb-exporter --db poker.db
	{
	  "scores": {
		"alyson": 10430,
		"leon": 5000
	  }
	}
