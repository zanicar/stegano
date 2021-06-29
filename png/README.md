# stegano/png

An image steganography implementation that outputs PNG steganograms. It accepts both JPEG and PNG images as input.

It satisfies the interface defined by [github.com/zanicar/stegano](https://github.com/zanicar/stegano)
A command line implementation is available [here](../cmd/stegano).

## Package Documentation

[https://pkg.go.dev/github.com/zanicar/stegano/png](https://pkg.go.dev/github.com/zanicar/stegano/png)

## How this package works

PNG images use lossless compression, thus data can be written to individual pixels without risk of being destroyed by the compression algorithm. For a JPEG implementation the data will have to be concealed within the compression algorithm's output matrices to prevent dataloss. This implementation also deliberately ignores the alpha channel for pixels, as alpha channels typically do not contain sufficient noise for concealment purposes.

The Conceal method conceals the given data in the pixels of the image decoded from reader and writes the result to writer as a new PNG image.

The Reveal method extracts data from the pixels of the image decoded from reader and writes the result to writer as a binary file.

Refer to the documentation given above for package usage.

## Examples

This tiny version of my profile image contains the link to this repository.

![steganogram](../examples/x.png)

More examples [here](../examples)
