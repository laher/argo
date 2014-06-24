// Copyright 2013 Am Laher.
// This code is adapted from code within the Go tree.
// See Go's licence information below:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package ar_test

import (
	"github.com/laher/argo/ar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// change this to true for generating an archive on the Filesystem
var isFs = false

func Example() {
	// Create a buffer to write our archive to.
	wtr := writer(isFs)

	// Create a new ar archive.
	aw := ar.NewWriter(wtr)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling licence."},
	}
	for _, file := range files {
		hdr := &ar.Header{
			Name: file.Name,
			Size: int64(len(file.Body)),
		}
		if err := aw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := aw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}
	// Make sure to check the error on Close.
	if err := aw.Close(); err != nil {
		log.Fatalln(err)
	}
	rdr := reader(isFs, wtr)
	arr, err := ar.NewReader(rdr)
	if err != nil {
		log.Fatalln(err)
	}

	// Iterate through the files in the archive.
	for {
		hdr, err := arr.Next()
		if err == io.EOF {
			// end of ar archive
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(os.Stdout, arr); err != nil {
			log.Fatalln(err)
		}
		fmt.Println()
	}

	// Output:
	// Contents of readme.txt:
	// This archive contains some text files.
	// Contents of gopher.txt:
	// Gopher names:
	// George
	// Geoffrey
	// Gonzo
	// Contents of todo.txt:
	// Get animal handling licence.
}

func reader(isFs bool, w io.Writer) io.Reader {
	if isFs {
		fi := w.(*os.File)
		err := fi.Close()
		if err != nil {
			log.Fatalln(err)
		}

		r, err := os.Open("tmp.ar")
		if err != nil {
			log.Fatalln(err)
		}
		return r
	} else {
		buf := w.(*bytes.Buffer)
		// Open the ar archive for reading.
		r := bytes.NewReader(buf.Bytes())
		return r
	}

}

func writer (isFs bool) io.Writer {
	if isFs {
		fi, err := os.Create("tmp.ar")
		if err != nil {
			log.Fatalln(err)
		}
		return fi
	} else {
		return new(bytes.Buffer)
	}

}

