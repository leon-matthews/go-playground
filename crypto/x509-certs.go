// Worked example to understand how the parts all fit together
package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

const addr = "localhost:4433"

func main() {
	// Script based on published code[1] from the Go stdlib, and an
	// article on Medium[2].
	// [1] https://go.dev/src/crypto/tls/generate_cert.go
	// [2] https://medium.com/@shaneutt/create-sign-x509-certificates-in-golang-8ac4ae49f903

	// Create self-signed Certificate Authority (CA)
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	authority := CreateCertificateAuthority(template, private.Public(), private)

	// PEM encode certificate
	authorityPEM := PEMEncodeCertificate(authority)
	log.Println("PEM encode certificate key\n", string(authorityPEM))

	// PEM encode private key
	privatePEM := PEMEncodePrivateKey(private)
	log.Println("PEM encode private key\n", string(privatePEM))

	// TODO: Create certificate signed by CA

	// Key-pair for HTTP server
	serverCert, err := tls.X509KeyPair(authorityPEM, privatePEM)

	// Create a tls.Config which will be provided to our server:
	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	// Start HTTPS server!
	/*
		mux := http.NewServeMux()
		mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, "Hello, TLS!\n") })
		server := http.Server{
			Addr:         addr,
			Handler:      mux,
			TLSConfig:    serverTLSConf,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Println("Listening on https://" + addr)
		log.Fatal(server.ListenAndServeTLS("", ""))
	*/
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "success!")
	}))
	server.TLS = serverTLSConf
	server.StartTLS()
	defer server.Close()

	// CertPool to house certificate for client connections
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(authorityPEM)

	// Create HTTP client
	clientTLSConf := &tls.Config{
		RootCAs: certpool,
	}
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}
	client := http.Client{
		Transport: transport,
	}

	// Make request
	resp, err := client.Get(server.URL)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[%T]%+[1]v\n", resp)
}

var template = &x509.Certificate{
	SerialNumber: CreateSerialNumber(),
	Subject: pkix.Name{
		Organization:  []string{"The Lost Continent of"},
		Country:       []string{"NZ"},
		Province:      []string{"Auckland"},
		Locality:      []string{"Waterview"},
		StreetAddress: []string{""},
		PostalCode:    []string{"90210"},
	},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().AddDate(10, 0, 0),
	IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	IsCA:                  true,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	BasicConstraintsValid: true,
}

// The returned slice is the CA certificate in DER encoding, as returned from x509.CreateCertificate()
func CreateCertificateAuthority(template *x509.Certificate, public crypto.PublicKey, private crypto.PrivateKey) []byte {
	// Combine to create self-signed, X.509 v3, certificate
	b, err := x509.CreateCertificate(rand.Reader, template, template, public, private)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

// CreateSerialNumber creates a cryptographically-secure 128-bit random number
func CreateSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128

	// According to the docs for Reader() and Int() from crypto/rand an error
	// is only possible on 'legacy' versions of Linux (entropy exhaustion on /dev/random).
	s, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(fmt.Sprint("Random number generation failed: %v", err))
	}
	return s
}

func PEMEncodeCertificate(cert []byte) []byte {
	buf := new(bytes.Buffer)
	pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	return buf.Bytes()
}

func PEMEncodePrivateKey(private crypto.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	pem.Encode(buf, &pem.Block{
		Bytes: privBytes,
		Type:  "PRIVATE KEY",
	})
	return buf.Bytes()
}
