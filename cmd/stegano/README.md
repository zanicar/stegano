# stegano-cli

A simple steganography command line tool written in Golang that supports ZIP compression and AES GCM block cipher encryption and message authentication.

This implementation is limited to data concealment in images, accepting both JPEG and PNG formats, with the output steganogram as PNG.

See [github.com/zanicar/stegano/png](https://github.com/zanicar/stegano/png) for further information.

## Features

* Fast data concealment in JPEG or PNG images (output as PNG)
* ZIP compression and decompression
* AES GCM Block Cipher encryption, decryption and message authentication

## Usage

    > stegano [options] -conceal -data {datafile} -in {inputfile} -out {outputfile}
    > stegano [options] -reveal -in {inputfile} -out {outputfile}

or

    > stegano [options] -conceal -data={datafile} -in={inputfile} -out={outputfile}
    > stegano [options] -reveal -in={inputfile} -out={outputfile}

Flags:

    -h  help
    -v  verbose mode
    -z  applies zip compression or decompression

Encryption / Decryption:

    -key {keystring}
    -key={keystring}

A SHA256 hash is computed for the key string to ensure it is exactly 32 bytes (as is required for the AES block cipher).

It is recommended to utilize a cryptographic key derivation function (such as scrypt or argon2) to produce highly secure keys. Such a key can then be passed as a hex or base64 encoded string.
