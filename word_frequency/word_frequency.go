
package main

import (
    "bufio"
    //~ "errors"
    "fmt"
    "log"
    "os"
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
        words[scanner.Text()] = 1;
    }
    return words, nil
}


func main() {
    log.SetFlags(0)
    if len(os.Args) != 2 {
        log.Fatal("usage: word_frequency PATH")
    }

    path := os.Args[1]
    fmt.Printf("Counting words from %v\n", path)

    words, err := count_words(path)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %v unique words\n", len(words))
}
