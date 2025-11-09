
From the *git diff* manual:

The --numstat option gives the diffstat(1) information but is designed for easier machine consumption. An entry in --numstat
output looks like this:

       1       2       README
       3       1       arch/{i386 => x86}/Makefile

   That is, from left to right:

        1. the number of added lines;
        2. a tab;
        3. the number of deleted lines;
        4. a tab;
        5. pathname (possibly with rename/copy information);
        6. a newline.