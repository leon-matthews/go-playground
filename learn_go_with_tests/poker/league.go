package poker

import (
	"encoding/json"
	"fmt"
	"io"
)

type League []Player

func (l League) Find(name string) *Player {
	for i, p := range l {
		if p.Name == name {
			return &l[i]
		}
	}
	return nil
}

func NewLeague(r io.Reader) (League, error) {
	var league []Player
	err := json.NewDecoder(r).Decode(&league)
	if err != nil {
		err = fmt.Errorf("problem parsing league, %w", err)
	}
	return league, err
}
