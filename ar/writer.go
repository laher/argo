package ar

/*
   Copyright 2013 Am Laher

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"
	"errors"
	"io"
	"strings"
)

var (
	ErrWriteTooLong    = errors.New("archive/ar: write too long")
	ErrFieldTooLong    = errors.New("archive/ar: header field too long")
	ErrWriteAfterClose = errors.New("archive/ar: write after close")
	errNameTooLong     = errors.New("archive/ar: name too long")
	errInvalidHeader   = errors.New("archive/ar: header field too long or contains invalid values")
	arFileHeader       = "!<arch>\n"
)

/* example showing ar file entries ...
!<arch>
debian-binary   1282478016  0     0     100644  4         `
2.0
control.tar.gz  1282478016  0     0     100644  444       `
.....binary-data.....
*/
// A Writer provides sequential writing of an ar archive.
// An ar archive consists of a sequence of files.
// Call WriteHeader to begin a new file, and then call Write to supply that file's data,
// writing at most hdr.Size bytes in total.
type Writer struct {
	w          io.Writer
	arFileHeaderWritten   bool
	err        error
	nb         int64 // number of unwritten bytes for current file entry
	pad        bool  // whether the file will be padded an extra byte (i.e. if ther's an odd number of bytes in the file)
	closed     bool
}

// NewWriter creates a new Writer writing to w.
func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }

// Flush finishes writing the current file (optional).
func (aw *Writer) Flush() error {
	if aw.nb > 0 {
		aw.err = fmt.Errorf("archive/tar: missed writing %d bytes", aw.nb)
		return aw.err
	}
	if !aw.arFileHeaderWritten {
		_, aw.err = aw.w.Write([]byte(arFileHeader))
		if aw.err != nil {
			return aw.err
		}
		aw.arFileHeaderWritten = true
	}
/*
	n := aw.nb
	if aw.pad {
		n += 1
	}
	for n > 0 && aw.err == nil {
		nr := n
		if nr > blockSize {
			nr = blockSize
		}
		var nw int
		nw, aw.err = aw.w.Write(zeroBlock[0:nr])
		n -= int64(nw)
	}
	*/
	aw.nb = 0
	aw.pad = false
	return aw.err
}

// WriteHeader writes hdr and prepares to accept the file's contents.
// WriteHeader calls Flush if it is not the first header.
// Calling after a Close will return ErrWriteAfterClose.
func (aw *Writer) WriteHeader(hdr *Header) error {
	return aw.writeHeader(hdr)
}

// WriteHeader writes hdr and prepares to accept the file's contents.
// WriteHeader calls Flush if it is not the first header.
// Calling after a Close will return ErrWriteAfterClose.
func (aw *Writer) writeHeader(hdr *Header) error {
	if aw.closed {
		return ErrWriteAfterClose
	}
	if aw.err == nil {
		aw.Flush()
	}
	if aw.err != nil {
		return aw.err
	}
	fmodTimestamp := fmt.Sprintf("%d", hdr.ModTime.Unix())
	//use root (for deb). These files are only for dpkg to extract as root anyway
	//this behaviour could be made configurable if 'Ar' gets used for anything else beyond .deb creation
	uid := fmt.Sprintf("%d", hdr.Uid)
	gid := fmt.Sprintf("%d", hdr.Gid)
	//Files only atm (not dirs)
	mode := fmt.Sprintf("10%d", hdr.Mode)
	size := fmt.Sprintf("%d", hdr.Size)
	line := fmt.Sprintf("%s%s%s%s%s%s`\n", pad(hdr.Name, 16), pad(fmodTimestamp, 12), pad(gid, 6), pad(uid, 6), pad(mode, 8), pad(size, 10))
	if _, err := aw.Write([]byte(line)); err != nil {
		return err
	}
	return nil
}

func (aw *Writer) Write(b []byte) (n int, err error) {
	if aw.closed {
		err = ErrWriteAfterClose
		return
	}
	
	if _, err := aw.w.Write(b); err != nil {
		aw.err = err
		return n, err
	}
	//0.5.4 bugfix: data section is 2-byte aligned.
	if len(b) % 2 == 1 {
		if _, err = aw.w.Write([]byte("\n")); err != nil {
			aw.err = err
			return
		}
	}

/*

	overwrite := false
	if int64(len(b)) > aw.nb {
		b = b[0:aw.nb]
		overwrite = true
	}

	n, err = aw.w.Write(b)
	aw.nb -= int64(n)
	if err == nil && overwrite {
		err = ErrWriteTooLong
		return
	}
	aw.err = err
*/
	return
}

// Close closes the ar archive, flushing any unwritten
// data to the underlying writer.
func (aw *Writer) Close() error {
	if aw.err != nil || aw.closed {
		return aw.err
	}
	aw.Flush()
	aw.closed = true
	if aw.err != nil {
		return aw.err
	}
/*
	// trailer: two zero blocks
	for i := 0; i < 2; i++ {
		_, aw.err = aw.w.Write(zeroBlock)
		if aw.err != nil {
			break
		}
	}
*/
	return aw.err
}
/*
func Ar(archiveFilename string, items []Archivable) error {
	// open output file
	fo, err := os.Create(archiveFilename)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		err := fo.Close()
		if err != nil {
			log.Printf("Error closing output file: %v", err)
		}
	}()
	header := "!<arch>\n"
	if _, err := fo.Write([]byte(header)); err != nil {
		log.Printf("Write error: %v", err)
		return err
	} else {
		for _, item := range items {
			fi, err := item.NewReader()
			if err != nil {
				return err
			} else {
				cl, ok := fi.(io.Closer)
				if ok {
					defer cl.Close()
				}
					inf, err := item.Header()
					if err != nil {
						return err
					}
					fmodTimestamp := fmt.Sprintf("%d", inf.ModTime.Unix())
					//use root (for deb). These files are only for dpkg to extract as root anyway
					//this behaviour could be made configurable if 'Ar' gets used for anything else beyond .deb creation
					uid := fmt.Sprintf("%d", inf.Uid)
					gid := fmt.Sprintf("%d", inf.Gid)
					//Files only atm (not dirs)
					mode := fmt.Sprintf("10%d", inf.Mode)
					size := fmt.Sprintf("%d", inf.Size)
					line := fmt.Sprintf("%s%s%s%s%s%s`\n", pad(inf.Name, 16), pad(fmodTimestamp, 12), pad(gid, 6), pad(uid, 6), pad(mode, 8), pad(size, 10))
					if _, err := fo.Write([]byte(line)); err != nil {
						return err
					} else {
						copyFile(fi, fo)
						//0.5.4 bugfix: data section is 2-byte aligned.
						if inf.Size % 2 == 1 {
							if _, err = fo.Write([]byte("\n")); err != nil {
								return err
							}
						}

					}
				//}
			}
		}
	}
	return err
}

func copyFile(fi io.Reader, fo io.Writer) error {
	// make a read buffer
	r := bufio.NewReader(fi)
	// make a write buffer
	w := bufio.NewWriter(fo)

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}
*/
func pad(value string, length int) string {
	plen := length - len(value)
	if plen > 0 {
		return value + strings.Repeat(" ", plen)
	}
	return value
}
