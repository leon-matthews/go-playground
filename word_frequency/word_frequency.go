// Count the frequency of every unique word found in the given file.

    $ ./word_frequency ../data/voyage-of-the-beagle.txt
    Counting words from ../data/voyage-of-the-beagle.txt
    Found 23432 unique words.
    40 most popular words are:
             the 15225
              of 9362
             and 5667
               a 5095
              to 4019
              in 3816
              is 2331
            that 1877
             was 1758
               I 1744
              on 1644
    ...snip...

*/


package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "sort"
)


// Count frequency of words found in given text file
func count_words(path string) (map[string]int, error) {
    // Open file & split by words
    fileHandle, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer fileHandle.Close()
    scanner := bufio.NewScanner(fileHandle)
    scanner.Split(bufio.ScanWords)

    // Count word freqencies using a map
    words := make(map[string]int)
    for scanner.Scan() {
        words[scanner.Text()] += 1;
    }
    return words, nil
}


// Print the given mapping of words, ordered by frequency
// Will stop printing the `limit` most popular words (set to zero to
// remove limit).
func print_words(words map[string]int, limit int) {
    // Extract keys into a slice
    keys := make([]string, 0, len(words))
    for key := range words {
        keys = append(keys, key)
    }

    // Sort new keys slice by frequency from words mapping
    sort.Slice(keys, func(i, j int) bool {
        return words[keys[i]] > words[keys[j]]
    })

    // Print words and their freqencies
    for index, key := range keys {
        fmt.Printf("%12v %v\n", key, words[key])
        if index == limit - 1 {
            break
        }
    }
}


func main() {
    // Check arguments
    log.SetFlags(0)
    if len(os.Args) != 2 {
        log.Fatal("usage: word_frequency PATH")
    }
    path := os.Args[1]

    // Count word frequencies
    fmt.Printf("Counting words from %v\n", path)
    words, err := count_words(path)
    if err != nil {
        log.Fatal(err)
    }

    // Print them
    limit := 40
    fmt.Printf("Found %v unique words.\n", len(words))
    if limit > 0 {
        fmt.Printf("%v most popular words are:\n", limit)
    }
    print_words(words, limit)
}
