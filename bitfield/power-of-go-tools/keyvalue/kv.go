// Package keyvalue uses encoding/gob to implement persistent key/value storage
package keyvalue

import (
	"encoding/gob"
	"errors"
	"io/fs"
	"os"
)

type KV struct {
	data map[string]string
	path string
}

// Open creates key-value store using given path.
// If path doesn't exist, it will be created when Save() called.
func Open(path string) (*KV, error) {
	kv := &KV{
		path: path,
		data: map[string]string{},
	}

	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return kv, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = gob.NewDecoder(f).Decode(&kv.data)
	if err != nil {
		return nil, err
	}

	return kv, nil
}

func (kv *KV) Get(key string) (string, bool) {
	v, ok := kv.data[key]
	return v, ok
}

func (kv *KV) Set(key, value string) {
	kv.data[key] = value
}

// Save persists data to disk
func (kv *KV) Save() error {
	f, err := os.Create(kv.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(kv.data)
}
