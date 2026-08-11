# Self-signed TLS key

Two files created using the command:

```bash
go run /usr/local/go/src/crypto/tls/generate_cert.go --host=localhost --rsa-bits=2048
```

Both in PEM format.

- `key.pem` holds the prate key.
- `cert.pem` self-signed TLS certificate containing public key.

Examine certifacate:

```bash
openssl x509 -in cert.pem -text -noout
```
