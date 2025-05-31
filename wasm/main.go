package main

import (
    "strings"
    "syscall/js"
)

// uppercaseFunction returns a js.Func that acts as an event listener for the input field. 
// It receives the input field via the `this` parameter.
func uppercaseFunction() js.Func {
    return js.FuncOf(func(this js.Value, args []js.Value) any {
        // Get the input field's value
        inputText := this.Get("value").String()

        // Process the value
        uppercaseText := strings.ToUpper(inputText)

        // Get the DOM document object
        document := js.Global().Get("document")

        // Find the output element and write the uppercase text to it
        outputElement := document.Call("getElementById", "output")
        outputElement.Set("textContent", uppercaseText)

        return nil
    })
}

func main() {
    // Add the uppercase function as an event listener to the input field
    js.Global().Get("document").Call("getElementById", "input").Call("addEventListener", "input", uppercaseFunction())

    // Keep the WebAssembly program running indefinitely
    select {}
}
