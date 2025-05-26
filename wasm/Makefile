.PHONY: clean run

run: main.wasm
	python3 -m http.server

clean:
	rm -f main.wasm

main.wasm: main.go
	GOOS=js GOARCH=wasm tinygo build -o main.wasm main.go
