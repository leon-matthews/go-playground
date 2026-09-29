package concurrency

type WebsiteChecker func(string) bool

type result struct {
    string
    bool
}

func CheckWebsites(checker WebsiteChecker, urls []string) map[string]bool {
    // Start workers
    numURLs := len(urls)
    resultChannel := make(chan result)
    for i := range numURLs {
        go func() {
            resultChannel <- result{urls[i], checker(urls[i])}
        }()
    }

    // Collect results
    results := make(map[string]bool, numURLs)
    for range numURLs {
        r := <-resultChannel
        results[r.string] = r.bool
    }

    return results
}
