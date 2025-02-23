# ED25519 Signatures

## Generate key-pair using `openssl`

    $ openssl genpkey -algorithm Ed25519 -out private_key.pem
    $ openssl pkey -in private_key.pem -pubout -out public_key.pem

