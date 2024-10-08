
# Word Frequency

Count the frequency of every unique word found in the given file.

Written as an initial exploration of Go, and to compare with similar programs
written in Python and Rust.

## Output

The output is sorted by the popularity of the words found, for example, using
the Gutenberg Project's public-domain, plain-text version of Darwin's book
'[Voyage of the Beagle](https://www.gutenberg.org/ebooks/944)' the output is:

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
             The 1602
            with 1585
              by 1570
           which 1516
              as 1512
              it 1383
            from 1304
              at 1241
             are 1153
            have 1083
             for 1045
            this 1035
            were 919
             not 915
             but 907
             one 897
           their 888
              we 875
              be 871
            they 844
              an 771
            been 766
              or 749
            very 727
             had 696
            some 641
             its 636
              so 612
           these 591

## Performance

The initial performance is fantastic, especially given that I am new to the
language and wrote it in a very straight-forward manner. On my workstation I
was able to process the 1.2MB book mentioned above in only 22.7ms!

    $ hyperfine './word_frequency ../data/voyage-of-the-beagle.txt'
      Time (mean ± σ):      22.7 ms ±   0.8 ms    [User: 20.2 ms, System: 4.2 ms]
      Range (min … max):    21.1 ms …  24.9 ms    128 runs

It will be interesting to revisit this example once I have more experience
in Go.
