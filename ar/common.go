
package ar

import (
	"errors"
	"os"
	"path"
	"time"
)

/* example showing ar file entries ...
!<arch>
debian-binary   1282478016  0     0     100644  4         `
2.0
control.tar.gz  1282478016  0     0     100644  444       `
.....binary-data.....
*/
const (
	blockSize    = 512
	headerSize   = 60
	arHeaderSize = 8

	fileNameSize = 16
	modTimeSize  = 12
	uidSize      = 6
	gidSize      = 6
	modeSize     = 8
	sizeSize     = 10
	magicSize    = 2
)

var (
	zeroBlock = make([]byte, headerSize)
	ErrHeader = errors.New("ar: invalid ar header")
)


type Header struct {
	// Name is the name of the file.
	// It must be a relative path: it must not start with a drive
	// letter (e.g. C:) or leading slash, and only forward slashes
	// are allowed.
	Name    string
	ModTime time.Time
	Uid     int
	Gid     int
	Mode    int64
	Size    int64
}

type slicer []byte

func (sp *slicer) next(n int) (b []byte) {
	s := *sp
	b, *sp = s[0:n], s[n:]
	return
}

// FileInfo returns an os.FileInfo for the Header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
}

// headerFileInfo implements os.FileInfo.
type headerFileInfo struct {
	h *Header
}

func (fi headerFileInfo) Size() int64 { return fi.h.Size }
func (fi headerFileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi headerFileInfo) ModTime() time.Time { return fi.h.ModTime }
func (fi headerFileInfo) Sys() interface{}   { return fi.h }

// Name returns the base name of the file.
func (fi headerFileInfo) Name() string {
	if fi.IsDir() {
		return path.Base(path.Clean(fi.h.Name))
	}
	return path.Base(fi.h.Name)
}

// Mode returns the permission and mode bits for the headerFileInfo.
func (fi headerFileInfo) Mode() (mode os.FileMode) {
	// Set file permission bits.
	mode = os.FileMode(fi.h.Mode).Perm()
	return mode
}

// FileInfoHeader creates a partially-populated Header from an
// os.FileInfo.
// Because os.FileInfo's Name method returns only the base name of
// the file it describes, it may be necessary to modify the Name field
// of the returned header to provide the full path name of the file.
func FileInfoHeader(fi os.FileInfo) (*Header, error) {
	size := fi.Size()
	h := &Header{
		Name:               fi.Name(),
		Size: int64(size),
	}
	fm := fi.Mode()
	h.ModTime = fi.ModTime()
	h.Mode = int64(fm.Perm())

	return h, nil
}
