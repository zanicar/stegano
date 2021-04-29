// Copyright 2018 Zanicar. All rights reserved.
// Utilizes a BSD-3 license. Refer to the included LICENSE file for details.

// Package stegano provides a simple interface for steganography implementations.
package stegano

import (
	"errors"
	"io"
)

// ErrCapacityMax means that a conceal received a length of bytes that exceeds
// its maximum capacity.
var ErrCapacityMax = errors.New("maximum capacity exceeded")

// ErrCapacityOverflow means that a conceal requires greater concealment capacity
// on the Reader to conceal the given length of bytes.
var ErrCapacityOverflow = errors.New("concealment capacity exceeded")

// Stegano is the interface that groups the basic Conceal and Reveal methods.
type Stegano interface {
	Concealer
	Revealer
}

// Concealer is the interface that wraps the basic Conceal method.
//
// Conceal conceals data into the bytes read from reader and writes
// the result to writer.
// Conceal must not modify the data slice, even temporarily.
//
// Implementations must not retain data.
type Concealer interface {
	Conceal(data []byte, reader io.Reader, writer io.Writer) error
}

// Revealer is the interface that wraps the basic Reveal method.
//
// Reveal reveals the underlying data from reader and writes it to writer.
type Revealer interface {
	Reveal(reader io.Reader, writer io.Writer) error
}
