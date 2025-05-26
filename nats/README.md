
# [Nats](https://github.com/nats-io/nats.go)

Go interface to the NATS messaging system.


## Applications

	$ sudo apt install nats-server
	$ go install github.com/nats-io/natscli/nats@latest

On the Raspberry Pi Zero with only 512MB of memory, you'll need to restrict both
the amount of RAM the build process uses, and the number of jobs to run in parallel:

	$ export GOMEMLIMIT=750MiB
	$ go install -p 1 github.com/nats-io/natscli/nats@latest


## Go Library

Create basic usage examples.
