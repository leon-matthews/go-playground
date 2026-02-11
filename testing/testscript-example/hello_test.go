package hello_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	hello "testscriptexample"
)

// Override TestMain to build and provide the binary 'hello' to test scripts.
func TestMain(m *testing.M) {
	// Binary name mapped to delegate function
	commands := map[string]func(){
		"hello": hello.Main,
	}
	testscript.Main(m, commands)
}

func TestHello(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
	})
}
