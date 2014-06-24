// Copyright 2013 Am Laher.
// This code is adapted from code within the Go tree.
// See Go's licence information below:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package ar

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Reader struct {
	r   io.Reader
	err error
	nb  int64 // number of unread bytes for current file entry
	pad bool  // whether the file will be padded an extra byte (i.e. if ther's an odd number of bytes in the file)
}

// NewReader creates a new Reader reading from r.
func NewReader(r io.Reader) (*Reader, error) {
	ar := &Reader{r: r}
	arHeader := make([]byte, arHeaderSize)
	_, err := io.ReadFull(ar.r, arHeader)
	if err != nil {
		return nil, err
	}
	if string(arHeader) != "!<arch>\n" {
		return nil, errors.New("ar: Invalid ar file!")
	}
	return ar, nil
}

// skipUnread skips any unread bytes in the existing file entry, as well as any alignment padding.
func (tr *Reader) skipUnread() {
	nr := tr.nb // number of bytes to skip
	if tr.pad {
		nr += int64(1)
		tr.pad = false
	}
	tr.nb = 0
	if sr, ok := tr.r.(io.Seeker); ok {
		if _, err := sr.Seek(nr, os.SEEK_CUR); err == nil {
			return
		}
	}

	_, tr.err = io.CopyN(ioutil.Discard, tr.r, nr)
}

func (tr *Reader) octal(b []byte) int64 {
	// Check for binary format first.
	if len(b) > 0 && b[0]&0x80 != 0 {
		var x int64
		for i, c := range b {
			if i == 0 {
				c &= 0x7f // ignore signal bit in first byte
			}
			x = x<<8 | int64(c)
		}
		return x
	}

	// Removing leading spaces.
	for len(b) > 0 && b[0] == ' ' {
		b = b[1:]
	}
	// Removing trailing NULs and spaces.
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\x00') {
		b = b[0 : len(b)-1]
	}
	x, err := strconv.ParseUint(string(b), 8, 64)
	if err != nil {
		tr.err = err
	}
	return int64(x)
}

// Next advances to the next entry in the ar archive.
func (tr *Reader) Next() (*Header, error) {
	var hdr *Header
	if tr.err == nil {
		tr.skipUnread()
	}
	if tr.err != nil {
		return hdr, tr.err
	}
	hdr = tr.readHeader()
	if hdr == nil {
		return hdr, tr.err
	}
	return hdr, tr.err
}

func (tr *Reader) NextString(max int) (string, error) {
	firstLine := make([]byte, max)
	n, err := io.ReadFull(tr.r, firstLine)
	tr.nb -= int64(n)
	if err != nil {
		tr.err = err
		log.Printf("failed to read first line of PKGDEF: %v", err)
		return "", err
	}
	return string(firstLine), nil
}

func (tr *Reader) readHeader() *Header {
	header := make([]byte, headerSize)
	if _, tr.err = io.ReadFull(tr.r, header); tr.err != nil {
		return nil
	}

	// Two blocks of zero bytes marks the end of the archive.
	if bytes.Equal(header, zeroBlock[0:headerSize]) {
		if _, tr.err = io.ReadFull(tr.r, header); tr.err != nil {
			return nil
		}
		if bytes.Equal(header, zeroBlock[0:headerSize]) {
			tr.err = io.EOF
		} else {
			tr.err = ErrHeader // zero block and then non-zero block
		}
		return nil
	}

	// Unpack
	hdr := new(Header)
	s := slicer(header)

	hdr.Name = strings.TrimSpace(string(s.next(fileNameSize)))
	modTime, err := strconv.Atoi(strings.TrimSpace(string(s.next(modTimeSize))))
	if err != nil {
		log.Printf("Error: (%+v)", tr.err)
		log.Printf(" (Header: %+v)", hdr)
		return nil
	}
	hdr.ModTime = time.Unix(int64(modTime), int64(0))
	hdr.Uid, tr.err = strconv.Atoi(strings.TrimSpace(string(s.next(uidSize))))
	if tr.err != nil {
		log.Printf("Error: (%+v)", tr.err)
		log.Printf(" (Header: %+v)", hdr)
		return nil
	}
	hdr.Gid, err = strconv.Atoi(strings.TrimSpace(string(s.next(gidSize))))
	if tr.err != nil {
		log.Printf("Error: (%+v)", tr.err)
		log.Printf(" (Header: %+v)", hdr)
		return nil
	}
	modeStr := strings.TrimSpace(string(s.next(modeSize)))
	hdr.Mode, tr.err = strconv.ParseInt(modeStr, 10, 64)
	sizeStr := strings.TrimSpace(string(s.next(sizeSize)))
	hdr.Size, tr.err = strconv.ParseInt(sizeStr, 10, 64)
	if tr.err != nil {
		log.Printf("Error: (%+v)", tr.err)
		log.Printf(" (Header: %+v)", hdr)
		return nil
	}
	magic := s.next(2) // magic
	if magic[0] != 0x60 || magic[1] != 0x0a {
		log.Printf("Invalid magic Header (%x,%x)", int(magic[0]), int(magic[1]))
		log.Printf(" (Header: %+v)", hdr)
		tr.err = ErrHeader
		return nil
	}
	if tr.err != nil {
		log.Printf("Error: (%+v)", tr.err)
		log.Printf(" (Header: %+v)", hdr)
		return nil
	}

	tr.nb = hdr.Size
	if math.Mod(float64(hdr.Size), float64(2)) == float64(1) {
		tr.pad = true
	} else {
		tr.pad = false
	}
	return hdr
}


// Read reads from the current entry in the ar archive.
// It returns 0, io.EOF when it reaches the end of that entry,
// until Next is called to advance to the next entry.
func (ar *Reader) Read(b []byte) (n int, err error) {
	if ar.nb == 0 {
		// file consumed
		return 0, io.EOF
	}

	if int64(len(b)) > ar.nb {
		b = b[0:ar.nb]
	}
	n, err = ar.r.Read(b)
	ar.nb -= int64(n)

	if err == io.EOF && ar.nb > 0 {
		err = io.ErrUnexpectedEOF
	}
	ar.err = err
	return
}
