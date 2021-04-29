// Copyright 2018 Zanicar. All rights reserved.
// Utilizes a BSD-3 license. Refer to the included LICENSE file for details.

package main

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/zanicar/stegano/png"
)

type opts struct {
	zip bool   // applies compression or decompression
	key []byte // applies encryption or decryption
}

func usage() {
	fmt.Printf("stegano: correct usage examples:\n")
	fmt.Printf("\t> stegano [options] -conceal -data {datafile} -in {inputfile} -out {outputfile}\n")
	fmt.Printf("\t> stegano [options] -reveal -in {inputfile} -out {outputfile}\n")
}

func conceal(dataFile, inputFile, outputFile string, options opts) error {
	// data
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("data file: %w", err)
	}

	// input file handler (reader)
	rfh, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("input file: %w", err)
	}
	defer rfh.Close()

	// output file handler (writer)
	wfh, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("output file: %w", err)
	}
	defer wfh.Close()

	// additional options
	if options.zip {
		zdata, err := compress(data)
		if err != nil {
			return fmt.Errorf("compress: %w", err)
		}
		data = zdata
	}

	if options.key != nil {
		cdata, err := encrypt(data, options.key)
		if err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}
		data = cdata
	}

	// steganography
	stegano := png.New()
	if err := stegano.Conceal(data, rfh, wfh); err != nil {
		return fmt.Errorf("conceal: %w", err)
	}

	return nil
}

func reveal(inputFile, outputFile string, options opts) error {
	// input file handler (reader)
	rfh, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("input file: %w", err)
	}
	defer rfh.Close()

	// output buffer
	buf := new(bytes.Buffer)

	// steganography
	stegano := png.New()
	if err := stegano.Reveal(rfh, buf); err != nil {
		return fmt.Errorf("reveal: %w", err)
	}

	// additional options
	if options.key != nil {
		pdata, err := decrypt(buf.Bytes(), options.key)
		if err != nil {
			return fmt.Errorf("decrypt: %w", err)
		}
		buf.Reset()
		if _, err := buf.Write(pdata); err != nil {
			return fmt.Errorf("decrypt: %w", err)
		}
	}

	if options.zip {
		zdata, err := decompress(buf.Bytes())
		if err != nil {
			return fmt.Errorf("decompress: %w", err)
		}
		buf.Reset()
		if _, err := buf.Write(zdata); err != nil {
			return fmt.Errorf("decompress: %w", err)
		}
	}

	// output file handler (writer)
	wfh, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("output file: %w", err)
	}
	defer wfh.Close()

	buf.WriteTo(wfh)

	return nil
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	// zip writer
	zw := zlib.NewWriter(&buf)
	n, err := zw.Write(data)
	if err != nil {
		return nil, err
	}

	// call Close explicitly to flush any unwritten data to the writer
	if err := zw.Close(); err != nil {
		return nil, err
	}

	log.Printf("%d bytes compressed to %d bytes", n, buf.Len())

	return buf.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	// input buffer
	var ibuf bytes.Buffer
	ibuf.Write(data)

	// zip reader
	zr, err := zlib.NewReader(&ibuf)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	// copy reader data to output buffer (writer) - prevents data going out of scope
	var obuf bytes.Buffer
	if _, err := io.Copy(&obuf, zr); err != nil {
		return nil, err
	}

	log.Printf("%d bytes decompressed to %d bytes", len(data), obuf.Len())

	return obuf.Bytes(), nil
}

func encrypt(data []byte, key []byte) ([]byte, error) {
	var buf bytes.Buffer

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Galois Counter Mode Block Cipher
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 96-bit nonce
	nonce := make([]byte, 12)
	if _, err := crand.Read(nonce); err != nil {
		return nil, err
	}
	buf.Write(nonce)

	// cipher data
	cd := aesgcm.Seal(data[:0], nonce, data, nil)
	buf.Write(cd)

	log.Printf("%d bytes encrypted to %d bytes", len(data), buf.Len())

	return buf.Bytes(), nil
}

func decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// nonce size is 12 bytes, the first 12 bytes of data
	nonce := data[:12]
	// cipher data follows after nonce
	cd := data[12:]

	// plain text bytes
	ptb, err := aesgcm.Open(nil, nonce, cd, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("%d bytes decrypted to %d bytes", len(data), len(ptb))

	return ptb, nil
}

func main() {
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	var fhelp bool
	flag.BoolVar(&fhelp, "h", false, "help")

	var fverbose bool
	flag.BoolVar(&fverbose, "v", false, "verbose mode")

	var fconceal, freveal bool
	flag.BoolVar(&fconceal, "conceal", false, "executes the conceal operation")
	flag.BoolVar(&freveal, "reveal", false, "executes the reveal operation")

	var dataFile, inputFile, outputFile string
	flag.StringVar(&dataFile, "data", "", "path to data file")
	flag.StringVar(&inputFile, "in", "", "path to input file")
	flag.StringVar(&outputFile, "out", "", "path to output file (create, overwrite)")

	var fzip bool
	flag.BoolVar(&fzip, "z", false, "applies zip compression or decompression")

	var key string
	flag.StringVar(&key, "key", "", "key used for encryption, decryption and message authentication (use secure key)")

	// Parse flags
	flag.Parse()

	// print help and return
	if fhelp {
		usage()
		fmt.Printf("\nflag and option details:\n")
		flag.PrintDefaults()
		return
	}

	// toggle verbose output
	if fverbose {
		log.SetOutput(os.Stderr)
	}

	// initialize execution options
	options := opts{
		zip: fzip,
		key: nil,
	}

	if key != "" {
		shaKey := sha256.Sum256([]byte(key))
		options.key = shaKey[:]
	}

	// Conceal
	if fconceal && dataFile != "" && inputFile != "" && outputFile != "" && !freveal {
		if err := conceal(dataFile, inputFile, outputFile, options); err != nil {
			log.SetOutput(os.Stderr)
			log.Fatal(err)
		}
		return
	}

	// Reveal
	if freveal && inputFile != "" && outputFile != "" && !fconceal {
		if err := reveal(inputFile, outputFile, options); err != nil {
			log.SetOutput(os.Stderr)
			log.Fatal(err)
		}
		return
	}

	usage()
}
