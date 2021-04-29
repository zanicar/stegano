# stegano/png

An image steganography implementation that outputs PNG steganograms. It accepts both JPEG and PNG images as input.

It satisfies the interface defined by [github.com/zanicar/stegano](https://github.com/zanicar/stegano)
A command line implementation is available [here](../cmd/stegano).

## Package Documentation

[https://pkg.go.dev/github.com/zanicar/stegano/png](https://pkg.go.dev/github.com/zanicar/stegano/png)

## Steganography in a nutshell

Steganography is the practice of concealing a message within another message. In the digital age however, we can redefine it as the practice of concealing data within other data. The general idea is to transfer or store a message or data in such a way, that nobody other than the sender and intended receiver is aware of it despite any observers. It is this that distinguishes steganography from cryptography, in cryptography any observer will be aware of the existence of the secret content.

Consider the following simple examples:

[Cryptography](https://en.wikipedia.org/wiki/Cryptography):

    Ipnaljo pz ayhjrpun fvb...
    
[Steganography](https://en.wikipedia.org/wiki/Steganography):

    Buddy, I got that episodic cluster headache. It sucks!
    Tell Randy about Chuck's key. I'm not going. You're obvioulsy unsure...
    
In the first case we use a simple Caesar Cipher (right shift by 7), but it is patently clear we are sending a message with content we want to keep secret from any oberservers. In the second case we use a simple Acrostic to hide our actual intended message and most observers would not give the message a second glace.

## How this package works

PNG images use lossless compression, thus data can be written to individual pixels without risk of being destroyed by the compression algorithm. For a JPEG implementation the data will have to be concealed within the compression algorithm's output matrices to prevent dataloss. This implementation also deliberately ignores the alpha channel for pixels, as alpha channels typically do not contain sufficient noise for concealment purposes.

The Conceal method conceals the given data in the pixels of the image decoded from reader and writes the result to writer as a new PNG image.

The Reveal method extracts data from the pixels of the image decoded from reader and writes the result to writer as a binary file.

Refer to the documentation given above for package usage.

### Why are images usefull for data concealment?

Digital images come in two primary formats, namely raster images and vector images. Raster images are comprized of many tiny dots, aka pixels, that represent light intensities in a two-dimensional spatial coordinate system. Vector images are comprized of mathematical functions that operate in terms of points on a Cartesian plane that can be transformed and rendered to a finite spatial coordinate system (a fancy way of saying they can be rendered as raster images). Raster images contain a lot of data and are ideally suited for data concealment, whereas vector images contain mostly formulas where even minor alterations may be easily noticed and are thus not suited for data concealment. In steganography this is known as the "concealment surface".

So just how much data do raster images contain?

Well, consider a standard Full HD display (1080p), this entails a display resolution of 1920x1080 pixels. That is 2 073 600 pixels, or roughly 2.1 Megapixels. Each pixel is represented by at least three color channels for Red, Green and Blue light intensities (an RGB image) where the intensities range from 0 to 255. Thus every colour pixel consists of at least 3 bytes. That gives us just over 6MB of uncompressed image data for a 1080p rgb image with no alpha channel.

To put it into perspective, "The Complete Works of William Shakespeare by William Shakespeare" as uncompressed UTF-8 text is only 5.5MB...

### But how is this relevant?

Well, every single rgb pixel has 16 777 216 different representations (256 for each of its three (RGB) channels). You may even be familiar with it:

    #831919     rgb(131, 25, 25)    // dark red
    #77ab59     rgb(119, 171, 89)   // soft green
    #0072fa     rgb(0, 114, 250)    // bright blue
    #eb7d46     rgb(235, 125, 70)   // citrus orange

Let us break it down even further using the dark red #831919:

| Base    |       Red |     Green |      Blue |
| :------ | --------: | --------: | --------: |
| Decimal |       131 |        25 |        25 |
| Hex     |        83 |        19 |        19 |
| Binary  | 1000 0011 | 0001 1001 | 0001 1001 |

What we will be doing is making small imperceptable changes to the intensities of each of the RGB channels. Would you be able to tell the difference if we changed our hex color to #801818? We will be hiding data in small changes like these.

In binary terms we will be hiding data in the two least significant bits, which means the two bits with the lowest values. Thus, going from 0x83 (facy way to denote a hex value) to 0x80 our binary would change from 10000011 to 10000000 - we changed only the two smallest bits and reduced the intensity from 131 to 128. Practically speaking we will never alter the value of any channel by more than 3, as this is the maximum value we can denote with two bits.

### Hiding data in color

Let's hide a "Cat" in the colors of the pixels given above.

So let us first obtain the data representation of "Cat". We have text containing three characters, each represented by a single byte of data. Each byte consists of eight bits and we will be hiding two bits per color channel. A pixel has three channels thus we need at least four pixels to hide our "Cat".

    1 byte = 8 bits
    2 bits per channel
    thus 4 "surface" bytes per concealed byte

Here is how we do it:

| DATA | Dec |    Binary |
| :--- | --: | --------: |
| C    |  67 | 0100 0011 |
| a    |  97 | 0110 0001 |
| t    | 116 | 0111 0100 |

| Pixel | Hex | Dec | Surface   |   | Data        | Concealed   | Dec | Hex |
| :---- | --: | --: | :-------- | - | :---------- | :---------- | --: | --: |
| 1:R   |  83 | 131 | 1000 0011 | C | 0100 00[11] | 1000 00[11] | 131 |  83 |
| 1:G   |  19 |  25 | 0001 1001 | C | 0100 [00]11 | 0001 10[00] |  24 |  18 |
| 1:B   |  19 |  25 | 0001 1001 | C | 01[00] 0011 | 0001 10[00] |  24 |  18 |
| 2:R   |  77 | 119 | 0111 0111 | C | [01]00 0011 | 0111 01[01] | 117 |  75 |
| 2:G   |  ab | 171 | 1010 1011 | a | 0110 00[01] | 1010 10[01] | 169 |  a9 |
| 2:B   |  59 |  89 | 0101 1001 | a | 0110 [00]01 | 0101 10[00] |  88 |  58 |
| 3:R   |  00 |   0 | 0000 0000 | a | 01[10] 0001 | 0000 00[10] |   2 |  02 |
| 3:G   |  72 | 114 | 0111 0010 | a | [01]10 0001 | 0111 00[01] | 113 |  71 |
| 3:B   |  fa | 250 | 1111 1010 | t | 0111 01[00] | 1111 10[00] | 248 |  f8 |
| 4:R   |  eb | 235 | 1110 1011 | t | 0111 [01]00 | 1110 10[01] | 233 |  e9 |
| 4:G   |  7d | 125 | 0111 1101 | t | 01[11] 0100 | 0111 11[11] | 127 |  7f |
| 4:B   |  46 |  70 | 0100 0110 | t | [01]11 0100 | 0100 01[01] |  69 |  45 |

We overwrite the two least significant bits of our pixels with the bits representing our "Cat" starting from the least significant bits (right to left). We get some new color codes that are imperceptably different from what we had, but we have a "Cat" hiding in there now...

Here are the old color codes alongside the new ones:

    #831919 -> #831818      rgb(131, 25, 25)    -> rgb(131, 24, 24)     // dark red
    #77ab59 -> #75a958      rgb(119, 171, 89)   -> rgb(117, 169, 88)    // soft green
    #0072fa -> #0271f8      rgb(0, 114, 250)    -> rgb(2, 113, 248)     // bright blue
    #eb7d46 -> #e97f45      rgb(235, 125, 70)   -> rgb(233, 127, 69)    // citrus orange

## Examples

This tiny version of my profile image contains the link to this repository.

![steganogram](../examples/x.png)

More examples [here](../examples)
