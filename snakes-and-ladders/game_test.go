package main

import "testing"

// TestMoveUnmarshalJSON checks the array form decodes, and bad input is rejected.
func TestMoveUnmarshalJSON(t *testing.T) {
	var move Move
	if err := move.UnmarshalJSON([]byte(`[4,14]`)); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if move.Roll != 4 || move.Square != 14 {
		t.Errorf("move = %+v, want roll 4, square 14", move)
	}
	for _, bad := range []string{`[300,1]`, `[1,-1]`, `"text"`, `{}`} {
		if err := move.UnmarshalJSON([]byte(bad)); err == nil {
			t.Errorf("UnmarshalJSON(%s) expected an error, got none", bad)
		}
	}
}
