package shift_test

import (
	"bytes"
	"testing"

	"github.com/gokatas/shift"
)

var testcases = []struct {
	plaintext, ciphertext []byte
	key                   byte
}{
	{[]byte("HAL"), []byte("IBM"), 1},
	{[]byte("BEEF"), []byte("LOOP"), 10},
}

func TestEncrypt(t *testing.T) {
	for _, tc := range testcases {
		got := shift.Encrypt(tc.plaintext, tc.key)
		if !bytes.Equal(tc.ciphertext, got) {
			t.Errorf("want %v, got %v", tc.ciphertext, got)
		}
	}
}

func TestDecrypt(t *testing.T) {
	for _, tc := range testcases {
		got := shift.Decrypt(tc.ciphertext, tc.key)
		if !bytes.Equal(tc.plaintext, got) {
			t.Errorf("want %v, got %v", tc.plaintext, got)
		}
	}
}
