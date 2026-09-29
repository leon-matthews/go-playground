package concurrency

type WebsiteChecker func(string) bool

type result struct {
    string
    bool
}

func CheckWebsites(checker WebsiteChecker, urls []string) map[string]bool {
    // Start workers
    resultChannel := make(chan result)
    for _, url := range urls {
        go func() {
            resultChannel <- result{url, checker(url)}
        }()
    }

    // Collect results
    results := make(map[string]bool, len(urls))
    for range len(urls) {
        r := <-resultChannel
        results[r.string] = r.bool
    }

    return results
}
