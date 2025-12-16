package main

import (
	"fmt"
	"log"
	"net/url"

    "github.com/jackc/pgconn"
    "github.com/lib/pq"
)


func main() {
    username := "Leon 'cool dude' Matthews"
    password := "Is this secure?"

    // pq.QuoteIdentifier
    fmt.Printf("pq.QuoteIdentifier: %s:%s\n", pq.QuoteLiteral(username), pq.QuoteLiteral(password))

    // pq.QuoteLiteral
    fmt.Printf("pq.QuoteLiteral: %s:%s\n", pq.QuoteLiteral(username), pq.QuoteLiteral(password))

    // net/url
    info := url.UserPassword(username, password)
    fmt.Printf("net/url: %s\n", info.String())

    url := fmt.Sprintf("postgres://%s@pexample.com:5432/sales", info.String())
    fmt.Println(url)
    config, err := pgconn.ParseConfig(url)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%+v\n", config)
}
