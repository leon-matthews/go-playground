package creditcard

import (
	"errors"
)

type card struct {
	number string
}

func New(number string) (*card, error) {
	if number == "" {
		return nil, errors.New("number must not be empty")
	}
	c := &card{number: number}
	return c, nil
}

func (c *card) Number() string {
	return c.number
}
