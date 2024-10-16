package main

import "fmt"

type ServerState int

const (
	StateIdle ServerState = iota
	StateConnected
	StateError
	StateRetrying
)

var stateName = map[ServerState]string{
	StateIdle:      "idle",
	StateConnected: "connected",
	StateError:     "error",
	StateRetrying:  "retrying",
}

// Manual implementation of `Stringer` interface could be replaced with
// use of `go generate` and the `stringer` tool.
// See the file the tool generated in this folder: `serverstate_string.go`.
func (s ServerState) String() string {
	return stateName[s]
}

func transition(s ServerState) ServerState {
	switch s {
	case StateIdle:
		return StateConnected
	case StateConnected, StateRetrying:
		return StateIdle
	case StateError:
		return StateError
	default:
		panic(fmt.Errorf("unknown state: %s", s))
	}
}

func main() {
	fmt.Println(StateIdle, StateConnected, StateError, StateRetrying)
	var oldState ServerState
	var newState ServerState

	// Idle to connected
	oldState = StateIdle
	newState = transition(oldState)
	fmt.Println(oldState, "->", newState)

	// Connected to idle
	oldState = StateConnected
	newState = transition(oldState)
	fmt.Println(oldState, "->", newState)

	// Retrying to idle
	oldState = StateRetrying
	newState = transition(oldState)
	fmt.Println(oldState, "->", newState)

	// Error to error
	oldState = StateError
	newState = transition(oldState)
	fmt.Println(oldState, "->", newState)
}
