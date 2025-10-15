
# Basic Markdown Preview Tool

Creates temporary HTML file from markdown input, then opens it in the 
system browser.

* Use the `template` package to assemble complete HTML file
* Use [markdown](https://github.com/gomarkdown/markdown) parser and HTML renderer
* Use [bluemonday](https://github.com/microcosm-cc/bluemonday) HTML sanitiser
* Use [testify](https://github.com/stretchr/testify) test assertions package

The book chapter used the `blackfriday` markdown package, but I replaced that 
with `markdown`, is the former seems to be unmaintained.
