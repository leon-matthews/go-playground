// Certificate generation and loading based on this blog post
// https://www.mejaz.in/posts/ed25519-digital-signatures-in-go
package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// load private key
	privateKey, err := loadPrivateKey(filepath.Join("certs", "private.pem"))
	if err != nil {
		fmt.Println("loading private key:", err)
		return
	}

	// Sign message
	message := []byte("hey, I am coming for movies tonight!")
	signature := ed25519.Sign(privateKey, message)
	signatureHex := hex.EncodeToString(signature)

	// Check signature
	publicKey, err := loadPublicKey(filepath.Join("certs", "public.pem"))
	if err != nil {
		fmt.Println("loading public key:", err)
		return
	}

	// convert hex string back to bytes
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		fmt.Println("error decoding signature hex:", err)
		return
	}

	// verify signature
	isValid := ed25519.Verify(publicKey, message, signatureBytes)

	fmt.Println("signature verification:", isValid)
}

// loadPrivateKey reads and parses private key from given path to PEM file
func loadPrivateKey(filename string) (ed25519.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey.(ed25519.PrivateKey), nil
}

// loadPublicKey reads and parses public key from given path to PEM file
func loadPublicKey(filename string) (ed25519.PublicKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey.(ed25519.PublicKey), nil
}
