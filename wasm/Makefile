# Minimal Go/WASM example project
#
# Requires:
#
# - TinyGo compiler (https://tinygo.org/)
# - The 'wasm-opt' WASM optimiser from the Binaryen project (https://github.com/WebAssembly/binaryen)
# - Any version of Python 3 to run the development webserver.
.PHONY: build clean run

build: main.wasm

run: build
	python3 -m http.server

clean:
	rm -f main.debug.wasm main.wasm

# Compile Go to WASM using TinyGo
# The standard Go compiler can be used, but results in a 1.7MB output vs 220kB for TinyGo
main.debug.wasm: main.go
	GOOS=js GOARCH=wasm tinygo build -o main.debug.wasm main.go
	chmod -x main.debug.wasm

# Strip debug symbols
# This gets us from 220kB to 50kB
main.wasm: main.debug.wasm
	wasm-opt --strip-debug -Os main.debug.wasm -o main.wasm
