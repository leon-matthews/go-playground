package main

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"math/rand/v2"
)

// Move records a single turn: the dice roll, then the square ended up on.
//
// Byte-sized fields keep memory traffic low. Rolls and squares never exceed
// 6 and 100, but note that Go silently truncates values over 255.
type Move struct {
	Roll   uint8
	Square uint8
}

// MarshalJSON writes a move as a two-element array, matching the Python tuples.
func (m Move) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "[%d,%d]", m.Roll, m.Square), nil
}

// UnmarshalJSON reads the two-element array form back; json rejects values over 255.
func (m *Move) UnmarshalJSON(data []byte) error {
	var pair [2]uint8
	if err := json.Unmarshal(data, &pair); err != nil {
		return err
	}
	m.Roll, m.Square = pair[0], pair[1]
	return nil
}

// Game is the full roll and position history of a single game.
type Game []Move

// board maps a square to where its ladder or snake leads, or zero if it has neither.
var board = [101]int{
	// Ladders
	1:  38,
	4:  14,
	9:  31,
	21: 42,
	28: 84,
	36: 44,
	51: 67,
	71: 91,
	80: 100,

	// Snakes
	98: 78,
	95: 75,
	93: 73,
	87: 24,
	64: 60,
	62: 19,
	56: 53,
	49: 11,
	48: 26,
	16: 6,
}

// snakesAndLadders plays a solo game of snakes and ladders.
//
// Standard rules: you need the exact roll to land on 100, do not move if roll
// overshoots it.
//
// Returns the list of moves taken to win the game. Each move is the dice
// roll, then the square you end up on. For example, one of the two possible
// shortest, 7 move games is:
//
//	[(4, 14), (6, 20), (6, 26), (2, 84), (5, 89), (5, 94), (6, 100)]
//
// The moves buffer is reused to avoid an allocation per game: the returned
// game is only valid until the next call with the same buffer.
//
// See:
//
//	https://en.wikipedia.org/wiki/Snakes_and_ladders
func snakesAndLadders(rng *rand.PCG, moves Game) Game {
	moves = moves[:0]
	place := 0
	for {
		// Roll the dice; the high word of x*6 maps x onto 0..5, bias one part in 2^64
		hi, _ := bits.Mul64(rng.Uint64(), 6)
		roll := int(hi) + 1
		landed := place + roll

		// Too high? Stay where you are. Otherwise, special move or as rolled.
		if landed <= 100 {
			if jump := board[landed]; jump != 0 {
				place = jump
			} else {
				place = landed
			}
		}

		moves = append(moves, Move{uint8(roll), uint8(place)})

		// Won? Require exact roll.
		if place == 100 {
			return moves
		}
	}
}
