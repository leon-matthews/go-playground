
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

    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }

    words := make(map[string]int)

    return words, nil
}


func main() {
    log.SetFlags(0)
    if len(os.Args) != 2 {
        log.Fatal("usage: word_frequency PATH")
    }

    path := os.Args[1]
    fmt.Println(fmt.Sprintf("Counting words from %v", path))

    words, err := count_words(path)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(fmt.Sprintf("Found %v unique words", len(words)))
}
