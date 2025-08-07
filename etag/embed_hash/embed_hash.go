// Package embed_hash exposes pre-computed hashes from inside embed.FS
package embed_hash

import (
	"encoding/base64"
	"embed"
	"io"
	"io/fs"
	"reflect"
)

// EmbedHashFS simply embeds [embed.FS] so we can modify its methods
type EmbedHashFS struct {
	embed.FS
}

// Open is a copy of the same method from [embed.FS.Open()]
func (e EmbedHashFS) Open(name string) (fs.File, error) {
	f, err := e.FS.Open(name)
	if err != nil {
		return nil, err
	}
	if rf, ok := f.(fs.ReadDirFile); ok {
		return EmbedHashDir{rf}, nil
	}
	return EmbedHashFile{f, f.(io.Seeker)}, nil
}

// EmbedHashDir is needed to provide a concrete type, as [embed.FS] doesn't export theirs
type EmbedHashDir struct {
	fs.ReadDirFile
}

type EmbedHashFile struct {
	fs.File
	io.Seeker // needed by http.File
}

// Content hash returns pre-computed file hash from embedded file
// Reflection dark magic from https://github.com/golang/go/issues/60940
func (f EmbedHashFile) ContentHash() string {
	openFile := reflect.ValueOf(f.File).Elem()
	file := openFile.Field(0).Elem()
	hashField := file.Field(2)
	hex := base64.RawStdEncoding.EncodeToString(hashField.Bytes())
	return hex
}
