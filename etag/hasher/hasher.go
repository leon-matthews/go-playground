// Package hasher provides hash of file's contents
package hasher

import (
	"io/fs"
)

// FileHasher extends [fs.File] to provide a hash method
type FileHasher interface {
	fs.File
	ContentHash() string
}
