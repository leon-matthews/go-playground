package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Sha256 calculates hexadecimal SHA-256 checksome of file's contents
func Sha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("calculating SHA-256: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("calculating SHA-256: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
