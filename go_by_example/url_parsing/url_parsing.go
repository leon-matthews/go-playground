package main

import (
	"fmt"
	"net"
	"net/url"
)

func main() {
	s := "postgres://user:pass@host.com:5432/path?k=v#f"
	fmt.Println(s)
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	// Scheme
	fmt.Println(u.Scheme)

	// Username and password
	fmt.Println(u.User)
	fmt.Println(u.User.Username())
	p, _ := u.User.Password()
	fmt.Println(p)

	// Host and port
	fmt.Println(u.Host)
	host, port, _ := net.SplitHostPort(u.Host)
	fmt.Println(host)
	fmt.Println(port)

	// Path and fragment
	fmt.Println(u.Path)
	fmt.Println(u.Fragment)

	// Query params
	fmt.Println(u.RawQuery)
	m, _ := url.ParseQuery(u.RawQuery)
	fmt.Println(m)
	fmt.Println(m["k"][0])
}
