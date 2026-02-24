package main

import (
    "fmt"
    "log"
    "net"
)

// LocalIP finds address of outgoing IPv4 interface
// Uses an outgoing UDP connection to find local IP address
// No connection is actually made, no data is sent.
func LocalIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    addr := conn.LocalAddr().(*net.UDPAddr)
    return addr.IP
}


// LocalIPs lists all network interface addresses
func LocalIPs() ([]net.IP, error) {
    var ips []net.IP
    addresses, err := net.InterfaceAddrs()
    if err != nil {
        return nil, err
    }

    for _, addr := range addresses {
        if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                ips = append(ips, ipnet.IP)
            }
        }
    }
    return ips, nil
}


func main() {
    fmt.Println(LocalIP())

    ips, err := LocalIPs()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(ips)
}
