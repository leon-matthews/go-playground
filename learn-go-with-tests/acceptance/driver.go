package go_specs_greet

import (
	"io"
	"net/http"
)

type Driver struct {
	BaseURL string
}

func (d Driver) Greet() (string, error) {
	response, err := http.Get(d.BaseURL + "/greet")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	greeting, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(greeting), nil
}
