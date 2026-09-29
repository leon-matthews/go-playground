package concurrency

import (
    "maps"
    "testing"
    "time"
)

func mockWebsiteChecker(url string) bool {
    return url != "wot://silly.one"
}

func slowWebsiteChecker(_ string) bool {
    time.Sleep(20 * time.Millisecond)
    return true
}

func TestCheckWebsites(t *testing.T) {
    urls := []string{
        "https://google.com",
        "https://lost.co.nz",
        "wot://silly.one",
    }

    want := map[string]bool{
        "https://google.com": true,
        "https://lost.co.nz": true,
        "wot://silly.one": false,
    }

    got := CheckWebsites(mockWebsiteChecker, urls)
    if !maps.Equal(got, want) {
        t.Errorf("got %v, wanted %v", got, want)
    }
}

func BenchmarkCheckWebsites(b *testing.B) {
    urls := make([]string, 100)
    for i := range urls {
        urls[i] = "https://example.com"
    }

    for b.Loop() {
        CheckWebsites(slowWebsiteChecker, urls)
    }
}
