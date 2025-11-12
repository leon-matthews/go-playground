package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"time"
)

func main() {
	// Script based on published code[1] from the Go stdlib, and an
	// article on Medium[2].
	// [1] https://go.dev/src/crypto/tls/generate_cert.go
	// [2] https://medium.com/@shaneutt/create-sign-x509-certificates-in-golang-8ac4ae49f903

	// Create self-signed Certificate Authority (CA)
	ca := CreateCertificateAuthority()

	// Create private key
	_, caPrivKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Combine to create self-signed, X.509 v3, certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, caPrivKey.Public(), caPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	// PEM encode certificate
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	log.Println("PEM encode certificate key\n", caPEM)

	// PEM encode private key
	caPrivKeyPEM := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(caPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	pem.Encode(caPrivKeyPEM, &pem.Block{Bytes: privBytes, Type: "PRIVATE KEY"})
	log.Println("PEM encode private key\n", caPrivKeyPEM)
}

// CreateCertificateAuthority creates self-signed CA
func CreateCertificateAuthority() *x509.Certificate {
	ca := &x509.Certificate{
		SerialNumber: CreateSerialNumber(),
		Subject: pkix.Name{
			Organization:  []string{"The Lost Continent of"},
			Country:       []string{"NZ"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Auckland"},
			PostalCode:    []string{"90210"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	log.Println("Create self-signed certificate, serial number:", ca.SerialNumber)
	return ca
}

// CreateSerialNumber creates a cryptographically-secure 128-bit random number
// Will panic if the crypto/rand.Reader
func CreateSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
	s, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(fmt.Sprint("Random number generation failed: %v", err))
	}
	return s
}
