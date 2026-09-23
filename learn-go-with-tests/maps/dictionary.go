package main

type Dictionary map[string]string

func Search(dict Dictionary, key string) string {
	return dict[key]
}
