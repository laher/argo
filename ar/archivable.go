package ar

import (
	"bytes"
	"io"
	"os"
	"syscall"
)

type Archivable interface {
	NewReader() (io.Reader, error)
	Header() (*Header, error)
}

type FilePathArchivable struct {
	Filename string
	ArchivePath string
	hdr *Header
}

func (a *FilePathArchivable) Header() (*Header, error) {
	if a.hdr == nil {
		finf, err := os.Stat(a.Filename)
		if err != nil {
			return nil, err
		}
		ai := new(Header)
		ai.Name = a.ArchivePath
		ai.ModTime = finf.ModTime()
		ai.Size = finf.Size()
		st, ok := finf.Sys().(*syscall.Stat_t)
		if ok {
			ai.Uid = int(st.Uid)
			ai.Gid = int(st.Gid)
		}
		ai.Mode = int64(finf.Mode())
		a.hdr = ai
		return ai, err
	}
	return a.hdr, nil
}

func (a *FilePathArchivable) NewReader() (io.Reader, error) {
	fi, err := os.Open(a.Filename)
	return fi, err
}

type BytesArchivable struct {
	Data []byte
	ArchivePathStr string
	Hdr *Header
}

func (a *BytesArchivable) Header() (*Header, error) {
	return a.Hdr, nil
}
func (a *BytesArchivable) NewReader() (io.Reader, error) {
	return bytes.NewBuffer(a.Data), nil
}
/*
type ReaderArchivable struct {
	Reader io.Reader
	ArchivePathStr string
	Hdr *Header
}*/
