
# Go Protobuf Example


Install tools:

    $ sudo apt install protobuf-compiler
    $ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

Generate Go file by compiling *.proto file:

    $ protoc -I=. --go_out=. --go_opt=paths=source_relative addresses/addresses.proto
