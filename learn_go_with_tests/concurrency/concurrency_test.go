package main

import (
	"maps"
	"testing"
	"time"
)

func mockWebsiteChecker(url string) bool {
	return url != "https://lost.co.nz"
}

func slowWebsiteChecker(url string) bool {
	time.Sleep(20 * time.Millisecond)
	return true
}

func TestCheckWebsites(t *testing.T) {
	websites := []string{
		"http://google.com",
		"https://lost.co.nz",
		"https://youtube.com",
	}

	got := CheckWebsites(mockWebsiteChecker, websites)

	want := map[string]bool{
		"http://google.com":   true,
		"https://lost.co.nz":  false,
		"https://youtube.com": true,
	}
	if !maps.Equal(want, got) {
		t.Fatalf("wanted %v, got %v", want, got)
	}
}

func BenchmarkCheckWebsites(b *testing.B) {
	urls := make([]string, 100)
	for i := 0; i < len(urls); i++ {
		urls[i] = "Some URL"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckWebsites(slowWebsiteChecker, urls)
	}
}
