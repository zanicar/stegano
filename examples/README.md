# examples

This folder contains a number of files that demonstrate the application of steganography.

## Files

[The Complete Works of William Shakespeare by William Shakespeare](https://www.gutenberg.org/ebooks/100) from [Project Gutenberg](https://www.gutenberg.org)

* 100-0.txt     (Plain Text UTF-8)
* pg100.epub    (EPUB - no images)

[Bust of William Shakespeare (1564-1616), 1760. Sculptor: John Michael Rysbrack](https://unsplash.com/photos/L2sbcLBJwOc?utm_source=unsplash&utm_medium=referral&utm_content=creditShareLink) from [Unsplash](https://unsplash.com)

* birmingham-museums-trust-L2sbcLBJwOc-unsplash.jpg

Steganograms

* shakespeare.png   (contains pg100.epub)
* x.png             (contains the link to this repository)

Reveal with ([cmd/stegano](https://github.com/zanicar/stegano/cmd/stegano)):

    stegano -reveal -in shakespeare.png -out shakespeare.epub -z
    stegano -reveal -in x.png -out repo.txt -z
