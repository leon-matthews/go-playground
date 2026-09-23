package main

import "errors"

var ErrExisting = errors.New("word already exists")
var ErrNotFound = errors.New("word not found")

type Dictionary map[string]string

func (d Dictionary) Add(key, value string) error {
	// RACE!
	_, exists := d[key]
	if exists {
		return ErrExisting
	}
	d[key] = value
	return nil
}

func (d Dictionary) Search(key string) (string, error) {
	v, ok := d[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (d Dictionary) Update(word, definition string) error {
	_, ok := d[word]
	if !ok {
		return ErrNotFound
	}
	d[word] = definition
	return nil
}
