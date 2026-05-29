package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

func main() {
	url := "https://lost.co.nz/"

	t := &timings{started: time.Now()}
	trace := newTrace(t)
	ctx := httptrace.WithClientTrace(context.Background(), trace)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	res.Body.Close()
	t.finished = time.Now()

	log.Printf("DNS lookup:       %v\n", t.dnsDone.Sub(t.dnsStart))
	log.Printf("TCP connect:      %v\n", t.connectDone.Sub(t.connectStart))
	log.Printf("TLS handshake:    %v\n", t.tlsDone.Sub(t.tlsStart))
	log.Printf("First byte:       %v\n", t.firstByte.Sub(t.gotConn))
	log.Printf("Content transfer: %v\n", t.finished.Sub(t.firstByte))
	log.Printf("Total:            %v\n", t.finished.Sub(t.started))
}

// timings tracks every single step of an HTTP transaction.
type timings struct {
	started      time.Time
	dnsStart     time.Time
	dnsDone      time.Time
	connectStart time.Time
	connectDone  time.Time
	tlsStart     time.Time
	tlsDone      time.Time
	gotConn      time.Time
	firstByte    time.Time
	finished     time.Time
}

// elapsed converts values inside timings into duration values.
func (t *timings) elapsed(at time.Time) time.Duration {
	return at.Sub(t.started)
}

// newTrace uses the stdlib httptrace package to populate (most) fields of given timings.
func newTrace(t *timings) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart:          func(_ httptrace.DNSStartInfo) { t.dnsStart = time.Now() },
		DNSDone:           func(_ httptrace.DNSDoneInfo) { t.dnsDone = time.Now() },
		ConnectStart:      func(_, _ string) { t.connectStart = time.Now() },
		ConnectDone:       func(_, _ string, _ error) { t.connectDone = time.Now() },
		TLSHandshakeStart: func() { t.tlsStart = time.Now() },
		TLSHandshakeDone:  func(_ tls.ConnectionState, _ error) { t.tlsDone = time.Now() },
		GotConn: func(info httptrace.GotConnInfo) {
			t.gotConn = time.Now()
			if info.Reused {
				log.Printf("connection reused (idle for %v)", info.IdleTime)
			} else {
				log.Printf("new connection to %s", info.Conn.RemoteAddr())
			}
		},
		GotFirstResponseByte: func() { t.firstByte = time.Now() },
	}
}
