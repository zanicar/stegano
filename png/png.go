// Copyright 2018 Zanicar. All rights reserved.
// Utilizes a BSD-3 license. Refer to the included LICENSE file for details.

// Package png provides a steganography implementation that outputs PNG image
// steganograms. It accepts both JPEG and PNG images as input.
package png

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/zanicar/stegano"
)

var (
	_ stegano.Stegano = &SteganoPNG{}
)

// CalculateCapacity determines the maximum number of bytes that can be
// concealed within the image of the given parameters.
func CalculateCapacity(width, height, channels, bitsPerByte int) int {
	return width * height * channels * bitsPerByte / 8
}

// SteganoPNG implements the Stegano interface for PNG image steganograms.
type SteganoPNG struct {
	hcoder [][]uint8
	hmap   map[uint8]uint8
}

// New returns a pointer to a new instance of SteganoPNG that is ready to use.
func New() *SteganoPNG {
	spng := &SteganoPNG{}
	spng.initHCoder()
	return spng
}

// dataLengthEncoder allows for the content length (specified in uint32 [4 bytes]) to be concealed
// by dividing the range of values of a single byte (0-255) into 4 slices, each representing one,
// two, three or four bytes. The content length can now be concealed in the target data by encoding
// an appropriately selected random number as the first concealed data byte to denote the number of
// bytes that represent the content length.
func (s *SteganoPNG) initHCoder() {
	s.hcoder = make([][]uint8, 4)
	s.hcoder[0] = make([]uint8, 0)
	s.hcoder[1] = make([]uint8, 0)
	s.hcoder[2] = make([]uint8, 0)
	s.hcoder[3] = make([]uint8, 0)
	s.hmap = make(map[uint8]uint8)

	for i := 0; i < 256; i++ {
		switch {
		case i%4 == 0 && i/4 > 0: // slice representing 4 bytes [4 294 967 296 -> 4GB]
			s.hcoder[3] = append(s.hcoder[3], uint8(i))
			s.hmap[uint8(i)] = 3
		case i%3 == 0 && i/3 > 0: // slice representing 3 bytes [16 777 216 -> 16MB]
			s.hcoder[2] = append(s.hcoder[2], uint8(i))
			s.hmap[uint8(i)] = 2
		case i%2 == 0 && i/2 > 0: // slice representing 2 bytes [65 536 -> 65KB]
			s.hcoder[1] = append(s.hcoder[1], uint8(i))
			s.hmap[uint8(i)] = 1
		case i%1 == 0 && i/1 > 0: // slice representing 1 byte [255]
			s.hcoder[0] = append(s.hcoder[0], uint8(i))
			s.hmap[uint8(i)] = 0
		}
	}
}

// headerBytes accepts the content length and returns its byte representation as a byte slice
// prepended with an appropritely selected random value that represents the number of bytes used.
// Thus the function returns a byte slice of length n + 1, where n is the minimum number of bytes
// required to represent the uint32 content length value.
func (s SteganoPNG) headerBytes(dlen int) ([]byte, error) {
	max := int(math.Pow(2, 32))
	if dlen > max {
		return nil, fmt.Errorf("%w: length (%v) max (%v)", stegano.ErrCapacityMax, dlen, max)
	}

	bitcount := len(fmt.Sprintf("%08b", dlen))
	bytecount := bitcount / 8
	if bitcount%8 > 0 {
		bytecount++
	}

	b := make([]byte, bytecount)
	l := dlen
	for bi := 0; bi < bytecount; bi++ {
		b[bi] |= uint8(l & 255)
		l = l >> 8
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch {
	case dlen < int(math.Pow(2, 8)):
		ri := r.Intn(len(s.hcoder[0]))
		b = append([]byte{s.hcoder[0][ri]}, b...)
	case dlen < int(math.Pow(2, 16)):
		ri := r.Intn(len(s.hcoder[1]))
		b = append([]byte{s.hcoder[1][ri]}, b...)
	case dlen < int(math.Pow(2, 24)):
		ri := r.Intn(len(s.hcoder[2]))
		b = append([]byte{s.hcoder[2][ri]}, b...)
	case dlen < int(math.Pow(2, 32)):
		ri := r.Intn(len(s.hcoder[3]))
		b = append([]byte{s.hcoder[3][ri]}, b...)
	}

	return b, nil
}

// Conceal encodes the given data to the two least significant bits of the RGB channels of the image
// decoded from reader and writes a new PNG image to the writer. The alpha channel is deliberately ommitted
// as alpha channels rarely provide sufficient noise for proper concealment. The given data is also spread
// across the available encoding bytes. The function returns an error on failure.
func (s SteganoPNG) Conceal(data []byte, r io.Reader, w io.Writer) error {
	log.Print("Conceal")
	sourceImg, _, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("image decode: %w", err)
	}

	bounds := sourceImg.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	hdata, err := s.headerBytes(len(data))
	if err != nil {
		return err
	}

	cap := CalculateCapacity(width, height, 3, 2)
	if len(data) > cap-len(hdata) {
		return fmt.Errorf("%w: length (%v) capacity (%v)", stegano.ErrCapacityOverflow, len(data), cap)
	}
	step := int(float64(cap-len(hdata)) / float64(len(data)) * 4)

	log.Printf("indexbyte=%d header=%d data=%d capacity=%d step=%d", hdata[0], len(hdata), len(data), cap, step)

	outputImg := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := sourceImg.At(x, y).RGBA()
			px := make([]byte, 3)
			// RGB values are pre-multiplied and represented as uint32 to prevent blendfactor multiplication overflow,
			// consequently they need to be normalized to a raw data byte (uint8).
			px[0] = uint8(r / 256)
			px[1] = uint8(g / 256)
			px[2] = uint8(b / 256)

			// pixel
			pxi := x + (y * width)
			// channel
			for ci := 0; ci < 3; ci++ {
				// available encoding byte
				abi := pxi*3 + ci
				// data byte index, start bit index (on data byte)
				var dbi, sbi int
				if abi < len(hdata)*4 {
					// set indices for header data
					dbi = abi / 4
					sbi = (abi % 4) * 2
				} else {
					// set indices for content data
					dbi = (abi - len(hdata)*4) / step
					sbi = ((abi - len(hdata)*4) % step) * 2
				}

				//log.Printf("Px: %04d Channel: %d ABi: %04d DBi: %04d", pxi, ci, abi, dbi)

				// conceal data bits on available encoding byte
				if sbi < 8 && dbi < len(data) {
					//log.Printf(" [conceal]")
					// encoding bits (two least significant on available encoding byte)
					for ebi := 0; ebi < 2; ebi++ {
						// bit index (on data byte)
						bi := sbi + ebi
						// bit value (from bit mask - e.g. bi:3 0000 0001 -> 0000 1000 = 8)
						var bit byte
						if abi < len(hdata)*4 {
							//log.Printf(" [header]")
							bit = hdata[dbi] & (1 << uint8(bi))
						} else {
							//log.Printf(" [content]")
							bit = data[dbi] & (1 << uint8(bi))
						}

						//log.Printf(" Bi: %d Bit: %02d", bi, bit)

						switch bit {
						case 0:
							px[ci] &^= uint8(ebi + 1) // set to 0
						default:
							px[ci] |= uint8(ebi + 1) // set to 1
						}
					}
				} /* else {
					//log.Printf(" [randomize]")
					n := rand.Intn(4)
					//log.Printf(" %v -> ", px[ci])
					px[ci] ^= uint8(n)
					//log.Printf("%v", px[ci])
				}*/
			}

			outputImg.Set(x, y, color.NRGBA{
				R: uint8(px[0]),
				G: uint8(px[1]),
				B: uint8(px[2]),
				A: uint8(a / 256),
			})
		}
	}

	if err := png.Encode(w, outputImg); err != nil {
		return err
	}

	log.Printf("%d bytes of data concealed", len(data))
	return nil
}

// Reveal uncovers any steganographic data from the PNG image decoded from reader and writes
// the output to the writer. It first decodes a mapped byte from the first available encoding byte
// to determine the length of the subsequent bytes holding the content length. The content length is
// decoded from the subsequent bytes whereafter the content is decoded from the entire data space.
// The function returns an error on failure.
func (s SteganoPNG) Reveal(r io.Reader, w io.Writer) error {
	log.Print("Reveal")
	sourceImg, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	bounds := sourceImg.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	cap := CalculateCapacity(width, height, 3, 2)

	log.Printf("capacity=%d", cap)

	// byte to rebuild
	var dbyte byte = 0
	var hdata []byte
	var data []byte
	step := 4
	clen := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := sourceImg.At(x, y).RGBA()
			px := make([]byte, 3)
			// RGB values are pre-multiplied and represented as uint32 to prevent blendfactor multiplication overflow,
			// consequently they need to be normalized to a raw data byte (uint8).
			px[0] = uint8(r / 256)
			px[1] = uint8(g / 256)
			px[2] = uint8(b / 256)

			// pixel
			pxi := x + (y * width)
			// channel
			for ci := 0; ci < 3; ci++ {
				// available encoding byte
				abi := pxi*3 + ci

				var dbi, sbi int
				if hdata == nil || abi < len(hdata)*4 {
					// set indices for header data
					dbi = abi / 4
					sbi = (abi % 4) * 2
				} else {
					// data byte index
					dbi = (abi - len(hdata)*4) / step
					// start bit index (on data byte)
					sbi = ((abi - len(hdata)*4) % step) * 2
				}

				//log.Printf("Px: %04d Channel: %d ABi: %04d DBi: %04d", pxi, ci, abi, dbi)

				if sbi < 8 && (data == nil || dbi < len(data)) {
					//log.Printf(" [reveal]")
					for ebi := 0; ebi < 2; ebi++ {
						// bit index (on data byte)
						bi := sbi + ebi

						// extract bit
						ebit := px[ci] & (1 << uint8(ebi))
						// get original bit value
						bit := ebit << uint8(bi/2*2)

						//log.Printf(" Bi: %d Bit: %02d", bi, bit)

						// rebuild byte
						dbyte |= bit

						// dbyte complete
						if bi == 7 {
							if hdata == nil || abi < len(hdata)*4 {
								//log.Printf(" [header]")
								// first byte indicates content length byte size
								if dbi == 0 {
									// hmap provides index (+1 for length, +1 for byte size byte)
									// dmap provides value (+1 for byte size byte)
									hdata = make([]byte, s.hmap[dbyte]+1+1)
									log.Printf("indexbyte=%d header=%d", dbyte, len(hdata))
									//log.Printf(" content length encoded to %v byte(s) [%v]", s.hmap[dbyte], dbyte)
								}
								hdata[dbi] = dbyte
								if dbi == len(hdata)-1 {
									// content length data
									cld := hdata[1:]
									for i := 0; i < len(cld); i++ {
										for ii := 0; ii < 8; ii++ {
											tbit := cld[i] & (1 << uint(ii))
											bit := uint(tbit) << uint(i*8)
											clen |= int(bit)
										}
									}
									data = make([]byte, clen)
									step = int(float64(cap-len(hdata)) / float64(clen) * 4)
									log.Printf("data=%d step=%d", clen, step)
									//log.Printf(" content length is %v bytes", clen)
								}
							} else {
								//log.Printf(" [content]")
								data[dbi] = dbyte
							}
							// reset dbyte to prevent data shadowing
							dbyte = 0
						}
					}
				}
				// TODO: verbose logging here
			}
		}
	}

	n, err := w.Write(data)
	if err != nil {
		return err
	}

	log.Printf("%d bytes of data revealed\n", n)
	return nil
}
