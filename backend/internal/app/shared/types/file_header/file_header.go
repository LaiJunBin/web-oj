package fileheader

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
)

type FileHeader struct {
	filename string
	header   textproto.MIMEHeader
	size     int64

	buffer *bytes.Buffer
}

func New(fh *multipart.FileHeader) (*FileHeader, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, err
	}
	// Close the opened part: for a file spooled to a temp file this releases a
	// file descriptor, and sustained uploads without it exhaust the process FD
	// limit. The bytes are fully copied into the buffer below, so nothing needs
	// the handle afterwards.
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &FileHeader{
		filename: fh.Filename,
		header:   fh.Header,
		size:     fh.Size,
		buffer:   bytes.NewBuffer(b),
	}, nil
}

func (f *FileHeader) Filename() string {
	return f.filename
}

func (f *FileHeader) Header() textproto.MIMEHeader {
	return f.header
}

func (f *FileHeader) Size() int64 {
	return f.size
}

func (f *FileHeader) Buffer() *bytes.Buffer {
	return f.buffer
}
