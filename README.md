# stegano

A Golang steganography module that contains an interface, a PNG image steganography implementation, and a command cline tool (CLI) thats supports ZIP compression and an AES GCM block cipher.

## Documentation

[https://pkg.go.dev/github.com/zanicar/stegano](https://pkg.go.dev/github.com/zanicar/stegano)

## Steganography in a nutshell

Steganography is the practice of concealing a message within another message. In the digital age however, we can redefine it as the practice of concealing data within other data. The general idea is to transfer or store a message or data in such a way, that nobody other than the sender and intended receiver is aware of it despite the presence of any observers. It is this attribute that distinguishes steganography from cryptography, in cryptography any observer will be aware of the existence of the secret content.

Consider the following simple examples:

[Cryptography](https://en.wikipedia.org/wiki/Cryptography):

    Ipnaljo pz ayhjrpun fvb...

Clearly a secret message (or a string of garbage).
    
[Steganography](https://en.wikipedia.org/wiki/Steganography):

    Buddy, I got that episodic cluster headache. It sucks!
    Tell Randy about Chuck's key. I'm not going. You're obvioulsy unsure...

A seemingly normal message (containing the same secret message as the cipher above).
    
In the first case we use a simple Caesar Cipher (right shift by 7), but it is patently clear we are sending a message with content we want to keep secret from any observers. In the second case we use a simple Acrostic to hide our actual intended message and most observers would not give the message a second glance.

## Steganography #TLDR (the deep dive)

![watching](parker-coffman-8EYMcqG5GRU-unsplash.jpg)
Photo by [Parker Coffman](https://unsplash.com/photos/8EYMcqG5GRU?utm_source=unsplash&utm_medium=referral&utm_content=creditShareLink) on [Unsplash](https://unsplash.com/s/photos/spy?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText)

To effectively use steganography we require a medium (**carrier**) capable of concealing our secret message (**payload**) somewhere or somehow (**channel**) within the carrier. Obviously the carrier must necessarily be larger than the payload in order to encode it (**encoding density**).

Any message that is considered to contain a payload is a **suspect** and any identified suspect is referred to as a **candidate**. A good steganography implementation seeks to minimize the chances of carriers being considered suspects and for suspects to be identified as candidates.

### Why are images so useful for data concealment?

Digital images come in two primary formats, namely raster images and vector images. A raster image is a collection of dots or pixels representing light intensities in a two-dimensional spatial coordinate system, this is the system utilized by cameras, screens and even your eye. A vector image is comprised of mathematical formulas that describe lines, curves, graphs and gradients that are in turn transformed and rendered to a finite coordinate system (a fancy way of saying they are rendered as raster images).

Raster images contain a lot of data that can easily be altered without distorting the image, thus making them perfectly suited for data concealment. In steganography this is referred to as the **concealment surface**.

_So how much data do typical raster images contain?_

A standard Full HD Display (1080p) has a display resolution of 1920x1080 pixels, that is precisely 2,073,600 pixels or roughly 2.1 MP (Megapixels). Each pixel is represented by at least three separate color channels for Red, Green and Blue light intensities with values ranging from 0-255 (8-bit / one byte per channel for 24-bit RGB color). That yields a total of just over 6 megabytes of uncompressed data for a single 2.1 Megapixel image with no alpha channel. Most Smart Phones have a primary camera that far exceeds 10 Megapixels with four channels (RGBA image) per pixel.

To put it all into perspective, "[The Complete Works of William Shakespeare by William Shakespeare](https://www.gutenberg.org/ebooks/100)" is merely 5.5MB (Megabytes) of uncompressed UTF-8 text... You can get it from [Project Gutenberg](https://www.gutenberg.org).

### So how do we do it?

Every single RGB pixel has 16 777 216 different representations (256 for each of its three RGB channels). You may even be familiar with the following color codes:

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

What we will be doing is making small imperceptible changes to the intensities of each of the RGB channels. Would you be able to tell the difference if we changed our hex color to #801818? We will be hiding data in small changes like these.

In binary terms we will be hiding data in the two least significant bits, which means the two bits with the lowest values. Thus, going from 0x83 (fancy way to denote a hex value) to 0x80 our binary would change from 10000011 to 10000000 - we changed only the two smallest bits and reduced the intensity from 131 to 128. Practically speaking we will never alter the value of any channel by more than 3, as this is the maximum value we can denote with two bits.

### Example: Hiding data in color

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

| Pixel | Hex | Dec | Surface   |   | Data          | Concealed     | Dec | Hex |
| :---- | --: | --: | :-------- | - | :------------ | :------------ | --: | --: |
| 1:R   |  83 | 131 | 1000 0011 | C | 0100 00**11** | 1000 00**11** | 131 |  83 |
| 1:G   |  19 |  25 | 0001 1001 | C | 0100 **00**11 | 0001 10**00** |  24 |  18 |
| 1:B   |  19 |  25 | 0001 1001 | C | 01**00** 0011 | 0001 10**00** |  24 |  18 |
| 2:R   |  77 | 119 | 0111 0111 | C | **01**00 0011 | 0111 01**01** | 117 |  75 |
| 2:G   |  ab | 171 | 1010 1011 | a | 0110 00**01** | 1010 10**01** | 169 |  a9 |
| 2:B   |  59 |  89 | 0101 1001 | a | 0110 **00**01 | 0101 10**00** |  88 |  58 |
| 3:R   |  00 |   0 | 0000 0000 | a | 01**10** 0001 | 0000 00**10** |   2 |  02 |
| 3:G   |  72 | 114 | 0111 0010 | a | **01**10 0001 | 0111 00**01** | 113 |  71 |
| 3:B   |  fa | 250 | 1111 1010 | t | 0111 01**00** | 1111 10**00** | 248 |  f8 |
| 4:R   |  eb | 235 | 1110 1011 | t | 0111 **01**00 | 1110 10**01** | 233 |  e9 |
| 4:G   |  7d | 125 | 0111 1101 | t | 01**11** 0100 | 0111 11**11** | 127 |  7f |
| 4:B   |  46 |  70 | 0100 0110 | t | **01**11 0100 | 0100 01**01** |  69 |  45 |

We overwrite the two least significant bits of our pixels with the bits representing our "Cat" starting from the least significant bits (right to left). We get some new color codes that are imperceptibly different from what we had, but we have a "Cat" hiding in there now...

Here are the old color codes alongside the new ones:

    #831919 -> #831818      rgb(131, 25, 25)    -> rgb(131, 24, 24)     // dark red
    #77ab59 -> #75a958      rgb(119, 171, 89)   -> rgb(117, 169, 88)    // soft green
    #0072fa -> #0271f8      rgb(0, 114, 250)    -> rgb(2, 113, 248)     // bright blue
    #eb7d46 -> #e97f45      rgb(235, 125, 70)   -> rgb(233, 127, 69)    // citrus orange

## Examples

This tiny version of my profile image contains the link to this repository.

![steganogram](examples/x.png)

More examples [here](examples/)
