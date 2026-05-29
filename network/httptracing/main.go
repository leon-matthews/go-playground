// Command httptracing measures the phases of a single HTTP request using
// the stdlib net/http/httptrace package.
package main

import (
	"context"
	"crypto/tls"
	"io"
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
	// Drain the body so 'content transfer' covers the whole response, not just the headers.
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		log.Fatal(err)
	}
	res.Body.Close()
	t.finished = time.Now()

	log.Printf("DNS lookup:       %v\n", t.dnsLookup())
	log.Printf("TCP connect:      %v\n", t.tcpConnect())
	log.Printf("TLS handshake:    %v\n", t.tlsHandshake())
	log.Printf("First byte:       %v\n", t.serverProcessing())
	log.Printf("Content transfer: %v\n", t.contentTransfer())
	log.Printf("Total:            %v\n", t.total())
}

// timings tracks the key timestamps of an HTTP request
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

// dnsLookup returns the time spent resolving the host name.
func (t *timings) dnsLookup() time.Duration { return t.dnsDone.Sub(t.dnsStart) }

// tcpConnect returns the time spent establishing the TCP connection.
func (t *timings) tcpConnect() time.Duration { return t.connectDone.Sub(t.connectStart) }

// tlsHandshake returns the time spent on the TLS handshake.
func (t *timings) tlsHandshake() time.Duration { return t.tlsDone.Sub(t.tlsStart) }

// serverProcessing returns the wait from having a connection to the first response byte.
func (t *timings) serverProcessing() time.Duration { return t.firstByte.Sub(t.gotConn) }

// contentTransfer returns the time spent reading the response body.
func (t *timings) contentTransfer() time.Duration { return t.finished.Sub(t.firstByte) }

// total returns the elapsed time for the whole request.
func (t *timings) total() time.Duration { return t.finished.Sub(t.started) }

// newTrace uses the stdlib httptrace package to populate (most) fields of given timings.
//
// The DNS, TCP and TLS timestamps are only set on a fresh connection; a reused
// keep-alive connection leaves them zero, so their deltas are meaningless.
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
